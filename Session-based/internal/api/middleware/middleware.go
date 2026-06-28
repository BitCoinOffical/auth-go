package middleware

import (
	"context"
	"log"
	"net/http"
	"sessions-based/internal/domain"
	"sessions-based/internal/domain/models"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	limitergin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	redisstore "github.com/ulule/limiter/v3/drivers/store/redis"
)

type AuthService interface {
	GetSession(ctx context.Context, sessionID string) (*models.Session, error)
}

type Middleware struct {
	service      AuthService
	ratelimitter *redis.Client
}

func NewMiddleware(service AuthService, ratelimitter *redis.Client) *Middleware {
	return &Middleware{service: service, ratelimitter: ratelimitter}
}

func (m *Middleware) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. достать session_id из cookie
		sessionId, err := c.Cookie("session_id")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		// 2. вызвать service.GetSession
		// 3. если ошибка — вернуть 401 и c.Abort()
		model, err := m.service.GetSession(c.Request.Context(), sessionId)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		// 4. положить сессию в контекст через c.Set
		c.Set("session", model)
		c.Next()
		// 5. c.Next()
	}
}

func (m *Middleware) RateLimiter() gin.HandlerFunc {
	rate, err := limiter.NewRateFromFormatted("5-M")
	if err != nil {
		log.Fatal(err)
	}

	store, err := redisstore.NewStoreWithOptions(m.ratelimitter, limiter.StoreOptions{
		Prefix: "rate_limiter",
	})
	if err != nil {
		log.Fatal(err)
	}
	instance := limiter.New(store, rate)
	return limitergin.NewMiddleware(instance, limitergin.WithLimitReachedHandler(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": domain.ErrTooManyRequests.Error()})
	}))
}
