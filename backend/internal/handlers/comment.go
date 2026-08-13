package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"incidenthub/backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type CommentHandler struct {
	db *sqlx.DB
}

func NewCommentHandler(db *sqlx.DB) *CommentHandler {
	return &CommentHandler{db: db}
}

func (h *CommentHandler) List(c *gin.Context) {
	incidentID := c.Param("id")

	var comments []models.Comment
	err := h.db.Select(&comments, `SELECT c.id, c.incident_id, c.user_id, c.body, c.created_at, c.updated_at
		FROM comments c WHERE c.incident_id = $1 ORDER BY c.created_at ASC`, incidentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list comments"})
		return
	}

	for i := range comments {
		var author models.User
		if err := h.db.Get(&author, "SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1", comments[i].UserID); err == nil {
			comments[i].Author = &author
		}
	}

	c.JSON(http.StatusOK, comments)
}

func (h *CommentHandler) Create(c *gin.Context) {
	incidentID := c.Param("id")
	userID := c.GetString("user_id")

	var incident models.Incident
	err := h.db.Get(&incident, "SELECT id FROM incidents WHERE id = $1", incidentID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "incident not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify incident"})
		return
	}

	var input models.CreateCommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if strings.TrimSpace(input.Body) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "comment body is required"})
		return
	}

	id := generateUUID()
	_, err = h.db.Exec(
		"INSERT INTO comments (id, incident_id, user_id, body) VALUES ($1, $2, $3, $4)",
		id, incidentID, userID, input.Body,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create comment"})
		return
	}

	var comment models.Comment
	h.db.Get(&comment, "SELECT id, incident_id, user_id, body, created_at, updated_at FROM comments WHERE id = $1", id)

	var author models.User
	if err := h.db.Get(&author, "SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1", userID); err == nil {
		comment.Author = &author
	}

	c.JSON(http.StatusCreated, comment)
}