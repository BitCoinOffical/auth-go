package models

import "github.com/google/uuid"

type LoginRequest struct {
	Email    string
	Password string
}

type User struct {
	ID       uuid.UUID
	UserName string
	Email    string
	Password string
}
