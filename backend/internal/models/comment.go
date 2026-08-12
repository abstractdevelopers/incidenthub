package models

import (
	"time"
)

type Comment struct {
	ID          string    `db:"id" json:"id"`
	IncidentID  string    `db:"incident_id" json:"incident_id"`
	UserID      string    `db:"user_id" json:"user_id"`
	Author      *User     `json:"author,omitempty"`
	Body        string    `db:"body" json:"body"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type CreateCommentInput struct {
	Body string `json:"body" binding:"required"`
}

type DashboardStats struct {
	Total         int64       `json:"total"`
	Open          int64       `json:"open"`
	Investigating int64       `json:"investigating"`
	Mitigated     int64       `json:"mitigated"`
	Resolved      int64       `json:"resolved"`
	Critical      int64       `json:"critical"`
}