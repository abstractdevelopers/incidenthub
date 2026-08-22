package handlers

import (
	"net/http"
	"regexp"
	"strings"

	"incidenthub/backend/internal/auth"
	"incidenthub/backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type AuthHandler struct {
	db     *sqlx.DB
	secret string
}

func NewAuthHandler(db *sqlx.DB, secret string) *AuthHandler {
	return &AuthHandler{db: db, secret: secret}
}

// passwordRegex enforces: min 8 chars, at least one uppercase, one lowercase, one digit.
var passwordRegex = regexp.MustCompile(`^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).{8}$`)

func (h *AuthHandler) Register(c *gin.Context) {
	var input models.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enforce password complexity: min 8 chars, uppercase, lowercase, digit
	if !passwordRegex.MatchString(input.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 characters and contain uppercase, lowercase, and a digit"})
		return
	}

	email := strings.ToLower(input.Email)

	// Check for existing user with row-level lock to prevent race conditions
	var existing models.User
	err := h.db.Get(&existing, "SELECT id FROM users WHERE email = $1 FOR UPDATE", email)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}
	if err.Error() != "sql: no rows in result set" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	hashedPassword, err := auth.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	id := generateUUID()
	result, err := h.db.Exec(
		"INSERT INTO users (id, email, password_hash, name) VALUES ($1, $2, $3, $4)",
		id, email, hashedPassword, input.Name,
	)
	if err != nil {
		// Check if this is a unique violation from a race condition
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify registration"})
		return
	}
	if rowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    id,
		"email": email,
		"name":  input.Name,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input models.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	err := h.db.Get(&user, "SELECT id, email, password_hash, name FROM users WHERE email = $1", strings.ToLower(input.Email))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	if !auth.CheckPasswordHash(input.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Email, user.Name, h.secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	resp := models.LoginResponse{
		Token: token,
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString("user_id")
	email := c.GetString("email")
	name := c.GetString("name")

	c.JSON(http.StatusOK, gin.H{
		"id":    userID,
		"email": email,
		"name":  name,
	})
}
