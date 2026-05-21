package interfaces

import (
	"github.com/gin-gonic/gin"
	"github.com/lriverd/big-service/internal/platform/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *CommentHandler, authMw *middleware.AuthMiddleware) {
	// Spot comments
	spotComments := rg.Group("/spots/:id/comments")
	{
		spotComments.GET("", handler.ListBySpot)

		auth := spotComments.Group("")
		auth.Use(authMw.RequireAuth())
		{
			auth.POST("", handler.Create)
		}
	}

	// Comment operations
	comments := rg.Group("/comments")
	comments.Use(authMw.RequireAuth())
	{
		comments.PUT("/:commentId", handler.Update)
		comments.DELETE("/:commentId", handler.Delete)
		comments.POST("/:commentId/like", handler.Like)
		comments.DELETE("/:commentId/like", handler.Unlike)
	}
}

