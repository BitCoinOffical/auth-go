package sessions_test

import (
	"context"
	"sessions-based/internal/domain/models"
	"sessions-based/internal/interfaces/sessions"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rdbtest "github.com/testcontainers/testcontainers-go/modules/redis"
)

func setupRedis(t *testing.T) *redis.Client {
	t.Helper()

	ctx := context.Background()

	container, err := rdbtest.Run(ctx, "redis:7")
	require.NoError(t, err)

	t.Cleanup(func() {
		container.Terminate(ctx)
	})

	addr, err := container.Endpoint(ctx, "")
	require.NoError(t, err)

	return redis.NewClient(&redis.Options{Addr: addr})
}

func TestSaveSession(t *testing.T) {
	rdb := setupRedis(t)
	rp := sessions.NewSessionRepo(rdb)
	err := rp.SaveSession(context.Background(), "id", &models.Session{
		UserID:    uuid.New(),
		Email:     "bob@gmail.com",
		CreatedAt: time.Now(),
	})
	assert.NoError(t, err)
}
func TestGetSession(t *testing.T) {
	testCases := []struct {
		name      string
		sessionID string
		success   bool
	}{
		{
			name:      "get session ok",
			sessionID: "id",
			success:   true,
		},
		{
			name:      "get session error",
			sessionID: "",
			success:   false,
		},
	}
	rdb := setupRedis(t)
	rp := sessions.NewSessionRepo(rdb)
	err := rp.SaveSession(context.Background(), "id", &models.Session{
		UserID:    uuid.New(),
		Email:     "bob@gmail.com",
		CreatedAt: time.Now(),
	})
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := rp.GetSession(context.Background(), tc.sessionID)

			if tc.success {
				assert.NoError(t, err)
				assert.NotEmpty(t, s)
			} else {
				assert.Error(t, err)
				assert.Empty(t, s)
			}

		})
	}

}

func TestDeleteSession(t *testing.T) {
	testCases := []struct {
		name      string
		sessionID string
		success   bool
	}{
		{
			name:      "delete session",
			sessionID: "id",
			success:   true,
		},
	}
	rdb := setupRedis(t)
	rp := sessions.NewSessionRepo(rdb)
	err := rp.SaveSession(context.Background(), "id", &models.Session{
		UserID:    uuid.New(),
		Email:     "bob@gmail.com",
		CreatedAt: time.Now(),
	})
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := rp.DeleteSession(context.Background(), tc.sessionID)
			assert.NoError(t, err)
		})
	}
}
