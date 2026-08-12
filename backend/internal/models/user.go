package models

import (
	"time"
)

type User struct {
	ID        string      `db:"id" json:"id"`
	Email     string      `db:"email" json:"email"`
	Password  string      `db:"password_hash" json:"-"`
	Name      string      `db:"name" json:"name"`
	CreatedAt time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt time.Time   `db:"updated_at" json:"updated_at"`
}

type CreateUserInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
}