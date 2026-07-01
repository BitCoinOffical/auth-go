package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sessions-based/internal/api/handlers"
	hmocks "sessions-based/internal/api/handlers/mock"
	"sessions-based/internal/domain"
	"sessions-based/internal/domain/models"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupRouters(h *handlers.AuthHandler, SetSession bool, Session *models.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	auth := gin.New()
	auth.GET("/auth/me", func(c *gin.Context) {
		if SetSession {
			c.Set("session", Session)
		}
		h.Me(c)
	})
	auth.POST("/auth/register", h.Register)
	auth.POST("/auth/login", h.Login)
	auth.DELETE("/auth/logout", h.Logout)
	return auth
}

type testDeps struct {
	r   *gin.Engine
	svc *hmocks.MockAuthService
}

func newDeps(t *testing.T, SetSession bool, Session *models.Session) *testDeps {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc := hmocks.NewMockAuthService(ctrl)
	h := handlers.NewAuthHandler(svc)
	r := setupRouters(h, SetSession, Session)
	return &testDeps{r: r, svc: svc}
}

func TestRegister(t *testing.T) {
	testCases := []struct {
		Name           string
		Body           map[string]string
		setup          func(svc *hmocks.MockAuthService)
		ExpectedStatus int
	}{
		{
			Name: "valid user",
			Body: map[string]string{
				"username":       "bob123",
				"email":          "bob@gmail.com",
				"password":       "secret12",
				"retry_password": "secret12",
			},
			setup: func(svc *hmocks.MockAuthService) {
				svc.EXPECT().Register(gomock.Any(), gomock.Any()).Return(uuid.New().String(), nil)
			},
			ExpectedStatus: http.StatusCreated,
		},
		{
			Name: "username too short",
			Body: map[string]string{
				"username":       "bo",
				"email":          "bob@gmail.com",
				"password":       "secret12",
				"retry_password": "secret12",
			},
			setup: func(svc *hmocks.MockAuthService) {

			},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name: "invalid email",
			Body: map[string]string{
				"username":       "bob123",
				"email":          "notanemail",
				"password":       "secret12",
				"retry_password": "secret12",
			},
			setup: func(svc *hmocks.MockAuthService) {

			},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name: "password too short",
			Body: map[string]string{
				"username":       "bob123",
				"email":          "bob@gmail.com",
				"password":       "short",
				"retry_password": "short",
			},
			setup: func(svc *hmocks.MockAuthService) {

			},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "empty body",
			Body:           map[string]string{},
			setup:          func(svc *hmocks.MockAuthService) {},
			ExpectedStatus: http.StatusBadRequest,
		},

		{
			Name: "user already exists",
			Body: map[string]string{
				"username":       "bob123",
				"email":          "bob@gmail.com",
				"password":       "secret12",
				"retry_password": "secret12",
			},
			setup: func(svc *hmocks.MockAuthService) {
				svc.EXPECT().Register(gomock.Any(), gomock.Any()).Return("", domain.ErrUserConflict)
			},
			ExpectedStatus: http.StatusConflict,
		},
		{
			Name: "invalid credentials",
			Body: map[string]string{
				"username":       "bob123",
				"email":          "bob@gmail.com",
				"password":       "secret12",
				"retry_password": "secret12",
			},
			setup: func(svc *hmocks.MockAuthService) {
				svc.EXPECT().Register(gomock.Any(), gomock.Any()).Return("", domain.ErrInvalidCredentials)
			},
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			deps := newDeps(t, false, nil)
			tc.setup(deps.svc)

			body, err := json.Marshal(tc.Body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			deps.r.ServeHTTP(w, req)

			assert.Equal(t, tc.ExpectedStatus, w.Code)

		})
	}
}

func TestMe(t *testing.T) {
	testCases := []struct {
		Name           string
		Session        *models.Session
		SetSession     bool
		setup          func(svc *hmocks.MockAuthService)
		ExpectedStatus int
	}{
		{
			Name:           "valid session",
			Session:        &models.Session{UserID: uuid.New(), Email: "bob@gmail.com"},
			SetSession:     true,
			setup:          func(svc *hmocks.MockAuthService) {},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "no session in context",
			SetSession:     false,
			setup:          func(svc *hmocks.MockAuthService) {},
			ExpectedStatus: http.StatusUnauthorized,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			deps := newDeps(t, tc.SetSession, tc.Session)
			tc.setup(deps.svc)

			req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)

			w := httptest.NewRecorder()
			deps.r.ServeHTTP(w, req)

			assert.Equal(t, tc.ExpectedStatus, w.Code)

		})
	}
}

func TestLogin(t *testing.T) {
	testCases := []struct {
		Name           string
		Body           map[string]string
		setup          func(svc *hmocks.MockAuthService)
		ExpectedStatus int
		Success        bool
	}{
		{
			Name: "valid login",
			Body: map[string]string{"email": "bob@gmail.com", "password": "secret12"},
			setup: func(svc *hmocks.MockAuthService) {
				svc.EXPECT().Login(gomock.Any(), gomock.Any()).Return(uuid.New().String(), nil)
			},
			ExpectedStatus: http.StatusOK,
			Success:        true,
		},
		{
			Name:           "invalid body",
			Body:           map[string]string{},
			setup:          func(svc *hmocks.MockAuthService) {},
			ExpectedStatus: http.StatusBadRequest,
			Success:        false,
		},
		{
			Name: "invalid credentials",
			Body: map[string]string{"email": "bob@gmail.com", "password": "wrongpass"},
			setup: func(svc *hmocks.MockAuthService) {
				svc.EXPECT().Login(gomock.Any(), gomock.Any()).Return("", domain.ErrInvalidCredentials)
			},
			ExpectedStatus: http.StatusUnauthorized,
			Success:        false,
		},
		{
			Name: "internal error",
			Body: map[string]string{"email": "bob@gmail.com", "password": "secret12"},
			setup: func(svc *hmocks.MockAuthService) {
				svc.EXPECT().Login(gomock.Any(), gomock.Any()).Return("", errors.New("db error"))
			},
			ExpectedStatus: http.StatusInternalServerError,
			Success:        false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			deps := newDeps(t, false, nil)
			tc.setup(deps.svc)

			body, err := json.Marshal(tc.Body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			deps.r.ServeHTTP(w, req)

			assert.Equal(t, tc.ExpectedStatus, w.Code)
		})
	}
}

func TestLogout(t *testing.T) {
	testCases := []struct {
		Name           string
		Cookie         string
		setup          func(svc *hmocks.MockAuthService)
		ExpectedStatus int
	}{
		{
			Name:   "valid logout",
			Cookie: "session-id-123",
			setup: func(svc *hmocks.MockAuthService) {
				svc.EXPECT().Logout(gomock.Any(), "session-id-123").Return(nil)
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "no cookie",
			Cookie:         "",
			setup:          func(svc *hmocks.MockAuthService) {},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:   "internal error",
			Cookie: "session-id-123",
			setup: func(svc *hmocks.MockAuthService) {
				svc.EXPECT().Logout(gomock.Any(), "session-id-123").Return(errors.New("redis error"))
			},
			ExpectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			deps := newDeps(t, false, nil)
			tc.setup(deps.svc)

			req := httptest.NewRequest(http.MethodDelete, "/auth/logout", nil)
			if tc.Cookie != "" {
				req.AddCookie(&http.Cookie{Name: "session_id", Value: tc.Cookie})
			}

			w := httptest.NewRecorder()
			deps.r.ServeHTTP(w, req)

			assert.Equal(t, tc.ExpectedStatus, w.Code)
		})
	}
}
