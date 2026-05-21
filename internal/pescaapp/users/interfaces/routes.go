package interfaces

import (
	"github.com/gin-gonic/gin"
	"github.com/lriverd/big-service/internal/platform/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, userHandler *UserHandler, favHandler *FavoriteHandler, authMw *middleware.AuthMiddleware) {
	users := rg.Group("/users")
	users.Use(authMw.RequireAuth())
	{
		users.GET("", userHandler.ListUsers)
		users.GET("/:id", userHandler.GetUser)
		users.PUT("/:id", userHandler.UpdateUser)
		users.GET("/:id/favorites", favHandler.ListFavorites)
		users.POST("/:id/favorites/:spotId", favHandler.AddFavorite)
		users.DELETE("/:id/favorites/:spotId", favHandler.RemoveFavorite)
	}
}

