package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"incidenthub/backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type IncidentHandler struct {
	db *sqlx.DB
}

func NewIncidentHandler(db *sqlx.DB) *IncidentHandler {
	return &IncidentHandler{db: db}
}

func (h *IncidentHandler) List(c *gin.Context) {
	var params models.IncidentListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 100 {
		params.Limit = 20
	}

	whereClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if params.Status != nil {
		whereClauses = append(whereClauses, "i.status = $"+strconv.Itoa(argIdx))
		args = append(args, *params.Status)
		argIdx++
	}
	if params.Severity != nil {
		whereClauses = append(whereClauses, "i.severity = $"+strconv.Itoa(argIdx))
		args = append(args, *params.Severity)
		argIdx++
	}
	if params.Search != nil && *params.Search != "" {
		searchTerm := "%" + strings.ToLower(*params.Search) + "%"
		whereClauses = append(whereClauses, "(LOWER(i.title) LIKE $"+strconv.Itoa(argIdx)+" OR LOWER(i.description) LIKE $"+strconv.Itoa(argIdx+1)+")")
		args = append(args, searchTerm, searchTerm)
		argIdx += 2
	}
	if params.Assignee != nil && *params.Assignee != "" {
		whereClauses = append(whereClauses, "i.assignee_id = $"+strconv.Itoa(argIdx))
		args = append(args, *params.Assignee)
		argIdx++
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	countQuery := "SELECT COUNT(*) FROM incidents i " + whereSQL
	var total int64
	if err := h.db.Get(&total, countQuery, args...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count incidents"})
		return
	}

	totalPages := int(total) / params.Limit
	if int(total)%params.Limit > 0 {
		totalPages++
	}

	offset := (params.Page - 1) * params.Limit
	listQuery := `SELECT i.id, i.title, i.description, i.severity, i.status, i.assignee_id,
		i.created_by, i.created_at, i.updated_at, i.resolved_at
		FROM incidents i ` + whereSQL + `
		ORDER BY i.created_at DESC
		LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)
	args = append(args, params.Limit, offset)

	var incidents []models.Incident
	if err := h.db.Select(&incidents, listQuery, args...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list incidents"})
		return
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Items:      incidents,
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	})
}

func (h *IncidentHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var incident models.Incident
	err := h.db.Get(&incident, `SELECT i.id, i.title, i.description, i.severity, i.status,
		i.assignee_id, i.created_by, i.created_at, i.updated_at, i.resolved_at
		FROM incidents i WHERE i.id = $1`, id)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "incident not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get incident"})
		return
	}

	if incident.AssigneeID != nil {
		var assignee models.User
		if err := h.db.Get(&assignee, "SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1", *incident.AssigneeID); err == nil {
			incident.Assignee = &assignee
		}
	}

	var creator models.User
	if err := h.db.Get(&creator, "SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1", incident.CreatedBy); err == nil {
		incident.Creator = &creator
	}

	c.JSON(http.StatusOK, incident)
}

func (h *IncidentHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")

	var input models.CreateIncidentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Status == "" {
		input.Status = models.StatusOpen
	}

	var resolvedAt *time.Time
	if input.Status == models.StatusResolved {
		now := time.Now()
		resolvedAt = &now
	}

	id := generateUUID()
	_, err := h.db.Exec(
		"INSERT INTO incidents (id, title, description, severity, status, assignee_id, created_by, resolved_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		id, input.Title, input.Description, input.Severity, input.Status, input.AssigneeID, userID, resolvedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create incident"})
		return
	}

	var incident models.Incident
	h.db.Get(&incident, "SELECT id, title, description, severity, status, assignee_id, created_by, created_at, updated_at, resolved_at FROM incidents WHERE id = $1", id)
	c.JSON(http.StatusCreated, incident)
}

func (h *IncidentHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var input models.UpdateIncidentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if input.Title != nil {
		setClauses = append(setClauses, "title = $"+strconv.Itoa(argIdx))
		args = append(args, *input.Title)
		argIdx++
	}
	if input.Description != nil {
		setClauses = append(setClauses, "description = $"+strconv.Itoa(argIdx))
		args = append(args, *input.Description)
		argIdx++
	}
	if input.Severity != nil {
		setClauses = append(setClauses, "severity = $"+strconv.Itoa(argIdx))
		args = append(args, *input.Severity)
		argIdx++
	}
	if input.Status != nil {
		setClauses = append(setClauses, "status = $"+strconv.Itoa(argIdx))
		args = append(args, *input.Status)
		argIdx++
		if *input.Status == models.StatusResolved {
			setClauses = append(setClauses, "resolved_at = NOW()")
		} else if *input.Status != models.StatusResolved {
			existingStatus := h.getIncidentStatus(id)
			if existingStatus == models.StatusResolved || existingStatus == "" {
				setClauses = append(setClauses, "resolved_at = NULL")
			}
		}
	}
	if input.AssigneeID != nil {
		setClauses = append(setClauses, "assignee_id = $"+strconv.Itoa(argIdx))
		args = append(args, *input.AssigneeID)
		argIdx++
	}

	if len(setClauses) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	args = append(args, id)

	query := "UPDATE incidents SET " + strings.Join(setClauses, ", ") + " WHERE id = $" + strconv.Itoa(argIdx)
	result, err := h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update incident"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "incident not found"})
		return
	}

	var incident models.Incident
	h.db.Get(&incident, "SELECT id, title, description, severity, status, assignee_id, created_by, created_at, updated_at, resolved_at FROM incidents WHERE id = $1", id)
	c.JSON(http.StatusOK, incident)
}

func (h *IncidentHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	result, err := h.db.Exec("DELETE FROM incidents WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete incident"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "incident not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *IncidentHandler) getIncidentStatus(id string) models.Status {
	var status models.Status
	h.db.Get(&status, "SELECT status FROM incidents WHERE id = $1", id)
	return status
}