package api

import (
	"context"
	"net/http"
	"sessions-based/internal/api/handlers"
	"time"

	"github.com/gin-gonic/gin"
)

type Middleware interface {
	AuthMiddleware() gin.HandlerFunc
	RateLimiter() gin.HandlerFunc
}

type Server struct {
	engine     *gin.Engine
	h          *handlers.AuthHandler
	middleware Middleware
	serv       *http.Server
}

func NewServer(h *handlers.AuthHandler, middleware Middleware) *Server {
	engine := gin.New()
	return &Server{
		engine:     engine,
		h:          h,
		middleware: middleware,
		serv: &http.Server{
			Addr:              ":8080",
			Handler:           engine,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
}

func (s *Server) Run() error {

	auth := s.engine.Group("/auth")
	{
		auth.GET("/me", s.middleware.AuthMiddleware(), s.h.Me)
		auth.POST("/register", s.middleware.RateLimiter(), s.h.Register)
		auth.POST("/login", s.middleware.RateLimiter(), s.h.Login)
		auth.DELETE("/logout", s.middleware.AuthMiddleware(), s.h.Logout)
	}

	return s.serv.ListenAndServe()
}
func (s *Server) Shutdown(ctx context.Context) error {
	return s.serv.Shutdown(ctx)
}
