package handlers

import (
	"context"
	"sessions-based/internal/api/middleware"
	"sessions-based/internal/domain/dto"
	"sessions-based/internal/domain/models"
	"sessions-based/internal/interfaces/repo"
	"sessions-based/internal/interfaces/services"
	"sessions-based/internal/interfaces/sessions"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (string, error)
	Login(ctx context.Context, req dto.LoginRequest) (string, error)
	Logout(ctx context.Context, sessionID string) error
	GetSession(ctx context.Context, sessionID string) (*models.Session, error)
}

type Services struct {
	Service    AuthService
	Middleware *middleware.Middleware
}

func NewServices(sessionRDB, LimitterRDB *redis.Client, db *pgxpool.Pool) *Services {
	sessionrepo := sessions.NewSessionRepo(sessionRDB)
	authrepo := repo.NewAuthRepo(db)
	service := services.NewSessionService(sessionrepo, authrepo)
	middleware := middleware.NewMiddleware(service, LimitterRDB)
	return &Services{Service: service, Middleware: middleware}
}

type Handlers struct {
	Auth *AuthHandler
}

func NewHandlers(srv *Services) *Handlers {
	auth := NewAuthHandler(srv.Service)
	return &Handlers{Auth: auth}
}
