package interfaces

import (
	"github.com/gin-gonic/gin"
	"github.com/lriverd/big-service/internal/platform/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *SpeciesHandler, authMw *middleware.AuthMiddleware) {
	species := rg.Group("/species")
	{
		species.GET("", handler.List)
		species.GET("/:id", handler.GetByID)

		admin := species.Group("")
		admin.Use(authMw.RequireAuth(), middleware.RequireAdmin())
		{
			admin.POST("", handler.Create)
			admin.PUT("/:id", handler.Update)
			admin.DELETE("/:id", handler.Delete)
		}
	}
}

