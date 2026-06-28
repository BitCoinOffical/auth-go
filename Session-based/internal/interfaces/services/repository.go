package services

import (
	"context"
	"sessions-based/internal/domain/models"

	"github.com/google/uuid"
)

type AuthRepo interface {
	UserExists(ctx context.Context, email string) (bool, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	RegisterUser(ctx context.Context, user *models.User) (uuid.UUID, error)
}

type SessionRepository interface {
	SaveSession(ctx context.Context, sessionID string, session *models.Session) error
	GetSession(ctx context.Context, sessionID string) (*models.Session, error)
	DeleteSession(ctx context.Context, sessionID string) error
}
