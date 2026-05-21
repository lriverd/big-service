package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lriverd/big-service/internal/platform/config"
	log "github.com/sirupsen/logrus"
)

type AuthMiddleware struct {
	jwtSecret string
}

func NewAuthMiddleware(cfg *config.Config) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: cfg.JWTSecret}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Missing authorization header",
				},
			})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Invalid authorization header format",
				},
			})
			return
		}

		claims, err := ParseJWT(parts[1], m.jwtSecret)
		if err != nil {
			log.WithError(err).Warn("Invalid JWT token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Invalid or expired token",
				},
			})
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userRole", claims.Role)
		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists || role.(string) != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "Admin access required",
				},
			})
			return
		}
		c.Next()
	}
}

