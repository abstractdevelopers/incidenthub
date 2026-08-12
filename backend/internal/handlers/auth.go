package handlers

import (
	"database/sql"
	"net/http"
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

func (h *AuthHandler) Register(c *gin.Context) {
	var input models.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.User
	err := h.db.Get(&existing, "SELECT id FROM users WHERE email = $1", strings.ToLower(input.Email))
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}
	if err != sql.ErrNoRows {
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
		id, strings.ToLower(input.Email), hashedPassword, input.Name,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    id,
		"email": strings.ToLower(input.Email),
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