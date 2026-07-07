package interfaces

import (
	"github.com/gin-gonic/gin"
	"github.com/lriverd/big-service/internal/platform/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *ReputationHandler, authMw *middleware.AuthMiddleware) {
	users := rg.Group("/users/:id")
	users.Use(authMw.RequireAuth())
	{
		users.GET("/reputation", handler.GetSummary)
		users.GET("/reputation/history", handler.GetHistory)
		users.GET("/penalties", handler.GetPenalties)
	}
}
