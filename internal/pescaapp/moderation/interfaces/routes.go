package interfaces

import (
	"github.com/gin-gonic/gin"
	"github.com/lriverd/big-service/internal/platform/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *ReportHandler, authMw *middleware.AuthMiddleware) {
	reports := rg.Group("/spots/:id/reports")
	reports.Use(authMw.RequireAuth())
	{
		reports.POST("", handler.Create)

		admin := reports.Group("")
		admin.Use(middleware.RequireAdmin())
		{
			admin.GET("", handler.ListBySpot)
		}
	}
}
