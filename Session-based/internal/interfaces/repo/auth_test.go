package repo_test

import (
	"context"
	"sessions-based/internal/adapters/migrations"
	"sessions-based/internal/domain/models"
	"sessions-based/internal/interfaces/repo"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()
	container, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2)))

	require.NoError(t, err)

	t.Cleanup(func() {
		container.Terminate(ctx)
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	err = migrations.RunMigrations(pool)
	require.NoError(t, err)

	return pool
}

func TestRegisterUser(t *testing.T) {
	testCases := []struct {
		Name    string
		Data    *models.User
		Success bool
	}{
		{
			Name: "register user",
			Data: &models.User{
				UserName: "Bob",
				Email:    "bob@gmail.com",
				Password: "secret",
			},
			Success: true,
		},

		{
			Name: "register exists user failed",
			Data: &models.User{
				UserName: "Bob",
				Email:    "bob@gmail.com",
				Password: "secret",
			},
			Success: false,
		},
	}
	pool := setupPostgres(t)
	rp := repo.NewAuthRepo(pool)

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			uuid, err := rp.RegisterUser(context.Background(), tc.Data)
			if tc.Success {
				assert.NoError(t, err)
				assert.NotEmpty(t, uuid)
			} else {
				assert.Error(t, err)
				assert.Empty(t, uuid)
			}
		})
	}
}

func TestGetByEmail(t *testing.T) {
	testCases := []struct {
		Name    string
		Data    *models.User
		Success bool
	}{
		{
			Name: "register user",
			Data: &models.User{
				UserName: "Bob",
				Email:    "bob@gmail.com",
				Password: "secret",
			},
			Success: true,
		},

		{
			Name: "not register user",
			Data: &models.User{
				UserName: "test",
				Email:    "test@gmail.com",
				Password: "testsecret",
			},
			Success: false,
		},
	}

	pool := setupPostgres(t)
	rp := repo.NewAuthRepo(pool)

	uuid, err := rp.RegisterUser(context.Background(),
		&models.User{
			UserName: "Bob",
			Email:    "bob@gmail.com",
			Password: "secret",
		})
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {

			user, err := rp.GetByEmail(context.Background(), tc.Data.Email)
			if tc.Success {
				assert.NoError(t, err)
				assert.Equal(t, "secret", user.Password)
				assert.Equal(t, user.ID, uuid)
			} else {
				assert.Error(t, err)
				assert.Empty(t, user)
			}
		})
	}
}

func TestUserExists(t *testing.T) {
	testCases := []struct {
		Name        string
		Data        *models.User
		Is_register bool
	}{
		{
			Name: "user conflict",
			Data: &models.User{
				UserName: "Bob",
				Email:    "bob@gmail.com",
				Password: "secret",
			},
			Is_register: true,
		},

		{
			Name: "user not conflict",
			Data: &models.User{
				UserName: "test",
				Email:    "test@gmail.com",
				Password: "testsecret",
			},
			Is_register: false,
		},

		{
			Name: "empty email",
			Data: &models.User{},
		},
	}

	pool := setupPostgres(t)
	rp := repo.NewAuthRepo(pool)

	_, err := rp.RegisterUser(context.Background(),
		&models.User{
			UserName: "Bob",
			Email:    "bob@gmail.com",
			Password: "secret",
		})
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {

			exists, err := rp.UserExists(context.Background(), tc.Data.Email)

			assert.NoError(t, err)
			if tc.Is_register {
				assert.Equal(t, true, exists)
			} else {
				assert.Equal(t, false, exists)
			}

		})
	}

}
