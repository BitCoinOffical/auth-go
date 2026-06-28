package handlers

import (
	"errors"
	"net/http"
	"sessions-based/internal/domain"
	"sessions-based/internal/domain/dto"
	"sessions-based/internal/domain/models"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	auth AuthService
}

func NewAuthHandler(auth AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Me(c *gin.Context) {
	val, exists := c.Get("session")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	session, ok := val.(*models.Session)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessionId, err := h.auth.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, domain.ErrUserConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "session_id",
		Value:    sessionId,
		Path:     "/",
		MaxAge:   86400,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	c.Status(http.StatusCreated)
}

func (h *AuthHandler) Login(c *gin.Context) {
	// 1. распарсить body в dto.LoginRequest
	var user dto.LoginRequest
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 2. вызвать service.Login
	sessionId, err := h.auth.Login(c.Request.Context(), user)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 3. положить session_id в cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "session_id",
		Value:    sessionId,
		Path:     "/",
		MaxAge:   86400,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	// 4. вернуть статус
	c.Status(http.StatusOK)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// 1. достать cookie
	sessionId, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 2. вызвать service.Logout
	err = h.auth.Logout(c.Request.Context(), sessionId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 3. удалить cookie у клиента
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "session_id",
		Value:    sessionId,
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	c.Status(http.StatusOK)
}
