package middleware

import (
	"auth/internal/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	sessionService *services.SessionService
}

func NewAuthMiddleware(sessionService *services.SessionService) *AuthMiddleware {
	return &AuthMiddleware{
		sessionService: sessionService,
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return m.authenticate(false)
}

func (m *AuthMiddleware) RequireSessionRotation() gin.HandlerFunc {
	return m.authenticate(true)
}

func (m *AuthMiddleware) authenticate(rotate bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No session found"})
			return
		}

		session, err := m.sessionService.GetSession(c.Request.Context(), sessionID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
			return
		}

		if time.Now().After(session.ExpiresAt) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Session expired"})
			return
		}

		if rotate {
			session, err = m.sessionService.RotateSession(c.Request.Context(), sessionID)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to rotate session"})
				return
			}
			c.SetCookie("session_id", session.ID, int(24*time.Hour.Seconds()), "/", "localhost", false, true)
		}

		c.Set("session", session)
		c.Next()
	}
}
