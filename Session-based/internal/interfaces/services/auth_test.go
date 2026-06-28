package services_test

import (
	"context"
	"errors"
	"sessions-based/internal/domain"
	"sessions-based/internal/domain/dto"
	"sessions-based/internal/domain/models"
	"sessions-based/internal/interfaces/services"
	srvmocks "sessions-based/internal/interfaces/services/mock"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

var ErrDBDown = errors.New("db down")
var ErrRedisDown = errors.New("redis down")
var ErrSaveSession = errors.New("save session error")

type testDeps struct {
	session *srvmocks.MockSessionRepository
	auth    *srvmocks.MockAuthRepo
	srvs    *services.AuthService
}

func newDeps(t *testing.T) *testDeps {
	ctrl := gomock.NewController(t)
	session := srvmocks.NewMockSessionRepository(ctrl)
	auth := srvmocks.NewMockAuthRepo(ctrl)
	srvs := services.NewSessionService(session, auth)
	return &testDeps{session: session, auth: auth, srvs: srvs}
}

func TestRegister(t *testing.T) {

	testCases := []struct {
		name    string
		data    dto.RegisterRequest
		setup   func(mockRepo *srvmocks.MockAuthRepo, mockSession *srvmocks.MockSessionRepository)
		wantErr error
		success bool
	}{
		{
			name: "invalid credentials",
			data: dto.RegisterRequest{
				UserName:      "bob1",
				Email:         "bob1@gmail.com",
				Password:      "password1",
				RetryPassword: "password2",
			},
			setup:   func(mockRepo *srvmocks.MockAuthRepo, mockSession *srvmocks.MockSessionRepository) {},
			wantErr: domain.ErrInvalidCredentials,
			success: false,
		},

		{
			name: "a user with this email already exists",
			data: dto.RegisterRequest{
				UserName:      "bob1",
				Email:         "bob1@gmail.com",
				Password:      "password1",
				RetryPassword: "password1",
			},
			setup: func(mockRepo *srvmocks.MockAuthRepo, mockSession *srvmocks.MockSessionRepository) {
				mockRepo.EXPECT().UserExists(gomock.Any(), gomock.Any()).Return(true, domain.ErrUserConflict).AnyTimes()
			},
			wantErr: domain.ErrUserConflict,
			success: false,
		},

		{
			name: "s.authRepo.UserExists error",
			data: dto.RegisterRequest{
				UserName:      "bob1",
				Email:         "bob1@gmail.com",
				Password:      "password1",
				RetryPassword: "password1",
			},
			setup: func(mockRepo *srvmocks.MockAuthRepo, mockSession *srvmocks.MockSessionRepository) {
				mockRepo.EXPECT().UserExists(gomock.Any(), gomock.Any()).Return(false, ErrDBDown).AnyTimes()
			},
			wantErr: ErrDBDown,
			success: false,
		},

		{
			name: "s.authRepo.RegisterUser error",
			data: dto.RegisterRequest{
				UserName:      "bob1",
				Email:         "bob1@gmail.com",
				Password:      "password1",
				RetryPassword: "password1",
			},
			setup: func(mockRepo *srvmocks.MockAuthRepo, mockSession *srvmocks.MockSessionRepository) {
				mockRepo.EXPECT().UserExists(gomock.Any(), gomock.Any()).Return(false, nil)
				mockRepo.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).Return(uuid.Nil, ErrDBDown).AnyTimes()
			},
			wantErr: ErrDBDown,
			success: false,
		},
		{
			name: "s.authRepo.SaveSession error",
			data: dto.RegisterRequest{
				UserName:      "bob1",
				Email:         "bob1@gmail.com",
				Password:      "password1",
				RetryPassword: "password1",
			},
			setup: func(mockRepo *srvmocks.MockAuthRepo, mockSession *srvmocks.MockSessionRepository) {
				mockRepo.EXPECT().UserExists(gomock.Any(), gomock.Any()).Return(false, nil)
				mockRepo.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).Return(uuid.New(), nil)
				mockSession.EXPECT().SaveSession(gomock.Any(), gomock.Any(), gomock.Any()).Return(ErrSaveSession).AnyTimes()
			},
			wantErr: ErrSaveSession,
			success: false,
		},

		{
			name: "Register ok",
			data: dto.RegisterRequest{
				UserName:      "bob1",
				Email:         "bob1@gmail.com",
				Password:      "password1",
				RetryPassword: "password1",
			},
			setup: func(mockRepo *srvmocks.MockAuthRepo, mockSession *srvmocks.MockSessionRepository) {
				mockRepo.EXPECT().UserExists(gomock.Any(), gomock.Any()).Return(false, nil)
				mockRepo.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).Return(uuid.New(), nil)
				mockSession.EXPECT().SaveSession(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			success: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := newDeps(t)
			tc.setup(d.auth, d.session)

			_, err := d.srvs.Register(context.Background(), tc.data)
			if !tc.success {
				assert.Error(t, err)
				if tc.wantErr != nil {
					assert.ErrorIs(t, err, tc.wantErr)
				}
			} else {
				assert.NoError(t, err)
			}

		})
	}

}

func TestLogin(t *testing.T) {
	testCases := []struct {
		name    string
		data    dto.LoginRequest
		setup   func(mockRepo *srvmocks.MockAuthRepo, mockSession *srvmocks.MockSessionRepository)
		wantErr error
		success bool
	}{
		{
			name: "not found user by email",
			data: dto.LoginRequest{
				Email:    "bob1@gmail.com",
				Password: "password1",
			},
			setup: func(mockRepo *srvmocks.MockAuthRepo, mockSession *srvmocks.MockSessionRepository) {
				mockRepo.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(nil, pgx.ErrNoRows)
			},
			wantErr: domain.ErrInvalidCredentials,
			success: false,
		},
		{
			name: "get by email error",
			data: dto.LoginRequest{
				Email:    "bob1@gmail.com",
				Password: "password1",
			},
			setup: func(mockRepo *srvmocks.MockAuthRepo, mockSession *srvmocks.MockSessionRepository) {
				mockRepo.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(nil, ErrDBDown)
			},
			wantErr: ErrDBDown,
			success: false,
		},
		{
			name: "save session error",
			data: dto.LoginRequest{
				Email:    "bob1@gmail.com",
				Password: "password1",
			},
			setup: func(mockRepo *srvmocks.MockAuthRepo, mockSession *srvmocks.MockSessionRepository) {
				pass, err := bcrypt.GenerateFromPassword([]byte("password1"), bcrypt.DefaultCost)
				assert.NoError(t, err)
				mockRepo.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(&models.User{
					ID:       uuid.New(),
					Password: string(pass),
				}, nil)
				mockSession.EXPECT().SaveSession(gomock.Any(), gomock.Any(), gomock.Any()).Return(ErrRedisDown)
			},
			wantErr: ErrRedisDown,
			success: false,
		},
		{
			name: "login ok",
			data: dto.LoginRequest{
				Email:    "bob1@gmail.com",
				Password: "password1",
			},
			setup: func(mockRepo *srvmocks.MockAuthRepo, mockSession *srvmocks.MockSessionRepository) {
				pass, err := bcrypt.GenerateFromPassword([]byte("password1"), bcrypt.DefaultCost)
				assert.NoError(t, err)
				mockRepo.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(&models.User{
					ID:       uuid.New(),
					Password: string(pass),
				}, nil)
				mockSession.EXPECT().SaveSession(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: nil,
			success: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := newDeps(t)
			tc.setup(d.auth, d.session)

			_, err := d.srvs.Login(context.Background(), tc.data)
			if !tc.success {
				assert.Error(t, err)
				if tc.wantErr != nil {
					assert.ErrorIs(t, err, tc.wantErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLogut(t *testing.T) {
	testCases := []struct {
		name    string
		data    string
		setup   func(mockSession *srvmocks.MockSessionRepository)
		wantErr error
		success bool
	}{
		{
			name: "logout error",
			data: "",
			setup: func(mockSession *srvmocks.MockSessionRepository) {
				mockSession.EXPECT().DeleteSession(gomock.Any(), gomock.Any()).Return(ErrRedisDown)
			},
			wantErr: ErrRedisDown,
			success: false,
		},
		{
			name: "loguot ok",
			data: "session-id",
			setup: func(mockSession *srvmocks.MockSessionRepository) {
				mockSession.EXPECT().DeleteSession(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: nil,
			success: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := newDeps(t)
			tc.setup(d.session)

			err := d.srvs.Logout(context.Background(), tc.data)
			if !tc.success {
				assert.Error(t, err)
				if tc.wantErr != nil {
					assert.ErrorIs(t, err, tc.wantErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetSession(t *testing.T) {
	testCases := []struct {
		name    string
		data    string
		setup   func(mockSession *srvmocks.MockSessionRepository)
		wantErr error
		success bool
	}{
		{
			name: "get session error",
			data: "",
			setup: func(mockSession *srvmocks.MockSessionRepository) {
				mockSession.EXPECT().GetSession(gomock.Any(), gomock.Any()).Return(nil, ErrRedisDown)
			},
			wantErr: ErrRedisDown,
			success: false,
		},
		{
			name: "get session ok",
			data: "session-id",
			setup: func(mockSession *srvmocks.MockSessionRepository) {
				mockSession.EXPECT().GetSession(gomock.Any(), gomock.Any()).Return(&models.Session{}, nil)
			},
			wantErr: nil,
			success: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := newDeps(t)
			tc.setup(d.session)
			_, err := d.srvs.GetSession(context.Background(), tc.data)
			if !tc.success {
				assert.Error(t, err)
				if tc.wantErr != nil {
					assert.ErrorIs(t, err, tc.wantErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSaveSession(t *testing.T) {
	testCase := []struct {
		name    string
		data    string
		setup   func(mockSession *srvmocks.MockSessionRepository)
		wantErr error
		success bool
	}{
		{
			name: "save session error",
			data: "session-id",
			setup: func(mockSession *srvmocks.MockSessionRepository) {
				mockSession.EXPECT().SaveSession(gomock.Any(), gomock.Any(), gomock.Any()).Return(ErrSaveSession)
			},
			wantErr: ErrSaveSession,
			success: false,
		},
		{
			name: "save session ok",
			data: "session-id",
			setup: func(mockSession *srvmocks.MockSessionRepository) {
				mockSession.EXPECT().SaveSession(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: nil,
			success: true,
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			d := newDeps(t)
			tc.setup(d.session)
			err := d.session.SaveSession(context.Background(), tc.data, &models.Session{
				UserID:    uuid.New(),
				Email:     "bob1@gmail.com",
				CreatedAt: time.Now(),
			})

			if !tc.success {
				assert.Error(t, err)
				if tc.wantErr != nil {
					assert.ErrorIs(t, err, tc.wantErr)
				}
			} else {
				assert.NoError(t, tc.wantErr)
			}
		})
	}
}
