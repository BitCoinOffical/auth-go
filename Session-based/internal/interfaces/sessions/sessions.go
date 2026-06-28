package sessions

import (
	"context"
	"encoding/json"
	"sessions-based/internal/domain"
	"sessions-based/internal/domain/models"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	sessionPrefix = "session:"
	SessionTTL    = 24 * time.Hour
)

type SessionRepository struct {
	rdb *redis.Client
}

func NewSessionRepo(rdb *redis.Client) *SessionRepository {
	return &SessionRepository{rdb: rdb}
}

func (s *SessionRepository) SaveSession(ctx context.Context, sessionID string, session *models.Session) error {
	value, err := json.Marshal(session)
	if err != nil {
		return err
	}
	key := sessionPrefix + sessionID
	return s.rdb.Set(ctx, key, value, SessionTTL).Err()
}

func (s *SessionRepository) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	key := sessionPrefix + sessionID
	res, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrSessionNotFound
		}
		return nil, err
	}
	var user models.Session
	err = json.Unmarshal([]byte(res), &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func (s *SessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	key := sessionPrefix + sessionID
	return s.rdb.Del(ctx, key).Err()
}
