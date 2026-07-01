package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sessions-based/internal/api/middleware"
	middlemocks "sessions-based/internal/api/middleware/mock"
	"sessions-based/internal/domain/models"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rdbtest "github.com/testcontainers/testcontainers-go/modules/redis"
	"go.uber.org/mock/gomock"
)

type Deps struct {
	m   *middleware.Middleware
	svc *middlemocks.MockAuthService
}

func newDeps(t *testing.T) *Deps {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc := middlemocks.NewMockAuthService(ctrl)

	m := middleware.NewMiddleware(svc, nil)
	return &Deps{m: m, svc: svc}
}

func TestAuthMiddleware(t *testing.T) {
	testCases := []struct {
		Name           string
		Cookie         string
		setup          func(svc *middlemocks.MockAuthService)
		ExpectedStatus int
	}{
		{
			Name:   "valid session",
			Cookie: "valid-session-id",
			setup: func(svc *middlemocks.MockAuthService) {
				svc.EXPECT().GetSession(gomock.Any(), "valid-session-id").
					Return(&models.Session{Email: "bob@gmail.com"}, nil)
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "no cookie",
			Cookie:         "",
			setup:          func(svc *middlemocks.MockAuthService) {},
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:   "session not found",
			Cookie: "invalid-session-id",
			setup: func(svc *middlemocks.MockAuthService) {
				svc.EXPECT().GetSession(gomock.Any(), "invalid-session-id").
					Return(nil, errors.New("session not found"))
			},
			ExpectedStatus: http.StatusUnauthorized,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			deps := newDeps(t)

			tc.setup(deps.svc)

			gin.SetMode(gin.TestMode)
			r := gin.New()

			r.GET("/protected", deps.m.AuthMiddleware(), func(c *gin.Context) {
				c.Status(tc.ExpectedStatus)
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tc.Cookie != "" {
				req.AddCookie(&http.Cookie{Name: "session_id", Value: tc.Cookie})
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.ExpectedStatus, w.Code)
		})
	}
}

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
func TestRateLimiter(t *testing.T) {
	rdb := setupRedis(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc := middlemocks.NewMockAuthService(ctrl)

	m := middleware.NewMiddleware(svc, rdb)
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/auth/register", m.RateLimiter(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for i := range 5 {
		req := httptest.NewRequest(http.MethodPost, "/auth/register", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "request %d should pass", i+1)
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/register", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code, "request 6 should pass")
}
