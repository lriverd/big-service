package interfaces

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, handler *SearchHandler) {
	rg.GET("/search", handler.Search)
}

