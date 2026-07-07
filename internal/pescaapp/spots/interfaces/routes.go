package interfaces

import (
	"github.com/gin-gonic/gin"
	"github.com/lriverd/big-service/internal/platform/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *SpotHandler, authMw *middleware.AuthMiddleware) {
	spots := rg.Group("/spots")
	{
		auth := spots.Group("")
		auth.Use(authMw.RequireAuth())
		{
			auth.GET("/mine", handler.Mine)
			auth.GET("/nearby-duplicates", handler.NearbyDuplicates)
			auth.POST("", handler.Create)
			auth.PUT("/:id", handler.Update)
			auth.DELETE("/:id", handler.Delete)

			admin := auth.Group("")
			admin.Use(middleware.RequireAdmin())
			{
				admin.PATCH("/:id/status", handler.UpdateStatus)
			}
		}

		spots.GET("", handler.List)
		spots.GET("/:id", handler.GetByID)
		spots.GET("/:id/nearby", handler.FindNearby)
	}
}
