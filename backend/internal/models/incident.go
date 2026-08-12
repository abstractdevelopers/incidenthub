package models

import (
	"time"
)

type Severity string

const (
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

type Status string

const (
	StatusOpen         Status = "OPEN"
	StatusInvestigating Status = "INVESTIGATING"
	StatusMitigated    Status = "MITIGATED"
	StatusResolved     Status = "RESOLVED"
)

type Incident struct {
	ID          string    `db:"id" json:"id"`
	Title       string    `db:"title" json:"title"`
	Description string    `db:"description" json:"description"`
	Severity    Severity  `db:"severity" json:"severity"`
	Status      Status    `db:"status" json:"status"`
	AssigneeID  *string   `db:"assignee_id" json:"assignee_id,omitempty"`
	Assignee    *User     `json:"assignee,omitempty"`
	CreatedBy   string    `db:"created_by" json:"created_by"`
	Creator     *User     `json:"creator,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
	ResolvedAt  *time.Time `db:"resolved_at" json:"resolved_at,omitempty"`
}

type CreateIncidentInput struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description" binding:"required"`
	Severity    Severity `json:"severity" binding:"required,oneof=LOW MEDIUM HIGH CRITICAL"`
	Status      Status   `json:"status" binding:"oneof=OPEN INVESTIGATING MITIGATED RESOLVED"`
	AssigneeID  *string  `json:"assignee_id,omitempty"`
}

type UpdateIncidentInput struct {
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	Severity    *Severity `json:"severity,omitempty"`
	Status      *Status  `json:"status,omitempty"`
	AssigneeID  *string  `json:"assignee_id,omitempty"`
}

type IncidentListParams struct {
	Status   *Status `form:"status"`
	Severity *Severity `form:"severity"`
	Search   *string `form:"search"`
	Assignee *string `form:"assignee"`
	Page     int     `form:"page"`
	Limit    int     `form:"limit"`
}

type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}