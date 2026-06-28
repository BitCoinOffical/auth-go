package repo

import (
	"context"
	"sessions-based/internal/domain/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepo struct {
	pool *pgxpool.Pool
}

func NewAuthRepo(pool *pgxpool.Pool) *AuthRepo {
	return &AuthRepo{pool: pool}
}

func (r *AuthRepo) UserExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	sql := "SELECT EXISTS(SELECT 1 FROM usersdb WHERE email = $1)"
	err := r.pool.QueryRow(ctx, sql, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *AuthRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	sql := "SELECT id, password FROM usersdb WHERE email = $1"
	var user models.User
	err := r.pool.QueryRow(ctx, sql, email).Scan(&user.ID, &user.Password)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
func (r *AuthRepo) RegisterUser(ctx context.Context, user *models.User) (uuid.UUID, error) {
	sql := "INSERT INTO usersdb (username, email, password, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id"
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, sql, user.UserName, user.Email, user.Password).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}
