package services

import (
	"context"
	"errors"
	"sessions-based/internal/domain"
	"sessions-based/internal/domain/dto"
	"sessions-based/internal/domain/models"
	generate "sessions-based/pkg/session"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	sessionRepo SessionRepository
	authRepo    AuthRepo
}

func NewSessionService(sessionRepo SessionRepository, authRepo AuthRepo) *AuthService {
	return &AuthService{sessionRepo: sessionRepo, authRepo: authRepo}
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (string, error) {
	if req.Password != req.RetryPassword {
		return "", domain.ErrInvalidCredentials
	}

	ok, err := s.authRepo.UserExists(ctx, req.Email)
	if err != nil {
		return "", err
	}

	if ok {
		return "", domain.ErrUserConflict
	}

	pass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user := &models.User{
		UserName: req.UserName,
		Email:    req.Email,
		Password: string(pass),
	}

	id, err := s.authRepo.RegisterUser(ctx, user)
	if err != nil {
		return "", err
	}

	sessionID, err := generate.GenerateSessionID()
	if err != nil {
		return "", err
	}

	session := &models.Session{
		UserID:    id,
		Email:     user.Email,
		CreatedAt: time.Now(),
	}
	err = s.sessionRepo.SaveSession(ctx, sessionID, session)
	if err != nil {
		return "", err
	}

	return sessionID, nil
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (string, error) {
	user, err := s.authRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrInvalidCredentials
		}
		return "", err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	sessionID, err := generate.GenerateSessionID()
	if err != nil {
		return "", err
	}

	session := &models.Session{
		UserID:    user.ID,
		Email:     req.Email,
		CreatedAt: time.Now(),
	}
	err = s.sessionRepo.SaveSession(ctx, sessionID, session)
	if err != nil {
		return "", err
	}

	return sessionID, nil
}

func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	return s.sessionRepo.DeleteSession(ctx, sessionID)
}

func (s *AuthService) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	return s.sessionRepo.GetSession(ctx, sessionID)
}
