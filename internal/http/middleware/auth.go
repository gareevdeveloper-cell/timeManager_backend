package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"testik/pkg/response"
)

// AuthMiddleware создаёт middleware для проверки JWT.
// TokenProvider — интерфейс для валидации токена (например, *auth.AuthService).
type TokenProvider interface {
	ValidateToken(tokenString string) (userID string, err error)
}

// Auth возвращает middleware, извлекающий Bearer token и проверяющий его.
func Auth(provider TokenProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "unauthorized", "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, http.StatusUnauthorized, "unauthorized", "invalid authorization format")
			c.Abort()
			return
		}

		userID, err := provider.ValidateToken(parts[1])
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "unauthorized", "invalid or expired token")
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
