package interfaces

import (
	"github.com/gin-gonic/gin"
	"github.com/lriverd/big-service/internal/platform/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *RatingHandler, authMw *middleware.AuthMiddleware) {
	ratings := rg.Group("/spots/:id/ratings")
	{
		ratings.GET("", handler.ListBySpot)

		auth := ratings.Group("")
		auth.Use(authMw.RequireAuth())
		{
			auth.POST("", handler.CreateOrUpdate)
			auth.DELETE("", handler.Delete)
		}
	}
}

