package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	SessionCookieName = "auth"
)

type Session struct {
	UserID    uuid.UUID
	Email     string
	CreatedAt time.Time
}
