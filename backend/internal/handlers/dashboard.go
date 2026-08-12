package handlers

import (
	"net/http"

	"incidenthub/backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type DashboardHandler struct {
	db *sqlx.DB
}

func NewDashboardHandler(db *sqlx.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

func (h *DashboardHandler) GetStats(c *gin.Context) {
	var stats models.DashboardStats

	if err := h.db.Get(&stats.Total, "SELECT COUNT(*) FROM incidents"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}
	if err := h.db.Get(&stats.Open, "SELECT COUNT(*) FROM incidents WHERE status = $1", models.StatusOpen); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}
	if err := h.db.Get(&stats.Investigating, "SELECT COUNT(*) FROM incidents WHERE status = $1", models.StatusInvestigating); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}
	if err := h.db.Get(&stats.Mitigated, "SELECT COUNT(*) FROM incidents WHERE status = $1", models.StatusMitigated); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}
	if err := h.db.Get(&stats.Resolved, "SELECT COUNT(*) FROM incidents WHERE status = $1", models.StatusResolved); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}
	if err := h.db.Get(&stats.Critical, "SELECT COUNT(*) FROM incidents WHERE severity = $1", models.SeverityCritical); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	var recent []models.Incident
	h.db.Select(&recent, `SELECT id, title, description, severity, status, assignee_id, created_by, created_at, updated_at, resolved_at
		FROM incidents ORDER BY created_at DESC LIMIT 5`)

	c.JSON(http.StatusOK, gin.H{
		"stats":  stats,
		"recent": recent,
	})
}