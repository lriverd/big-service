package interfaces

import (
	"github.com/gin-gonic/gin"
	"github.com/lriverd/big-service/internal/platform/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *AuthHandler, authMw *middleware.AuthMiddleware) {
	auth := rg.Group("/auth")
	{
		auth.POST("/login", handler.Login)
		auth.POST("/login/password", handler.LoginWithPassword)
		auth.POST("/register", handler.Register)
		auth.POST("/logout", authMw.RequireAuth(), handler.Logout)
	}
}

