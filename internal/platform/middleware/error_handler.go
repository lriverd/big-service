package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			log.WithError(err.Err).WithField("path", c.Request.URL.Path).Error("Unhandled error")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Internal server error",
				},
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
		}
	}
}

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.WithField("panic", recovered).Error("Panic recovered")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Internal server error",
			},
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})
}

