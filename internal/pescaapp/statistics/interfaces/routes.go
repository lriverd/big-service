package interfaces

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, handler *StatsHandler) {
	rg.GET("/spots/:id/stats", handler.GetSpotStats)
	rg.GET("/user/:userId/stats", handler.GetUserStats)
	rg.GET("/statistics/popular-spots", handler.GetPopularSpots)
}

