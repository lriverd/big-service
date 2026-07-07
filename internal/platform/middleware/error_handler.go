package middleware

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// ErrorHandler catches errors registered via c.Error(err) that a handler
// didn't already turn into a response itself. Most handlers log and respond
// directly (see internal/shared/response.InternalError) — this is a safety
// net for anything that slips through without doing so.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 && !c.Writer.Written() {
			err := c.Errors.Last()
			log.WithError(err.Err).WithFields(log.Fields{
				"path":      c.Request.URL.Path,
				"method":    c.Request.Method,
				"requestId": GetRequestID(c),
			}).Error("Unhandled error")
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
		log.WithFields(log.Fields{
			"panic":     recovered,
			"path":      c.Request.URL.Path,
			"method":    c.Request.Method,
			"requestId": GetRequestID(c),
			"stack":     string(debug.Stack()),
		}).Error("Panic recovered")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Internal server error",
			},
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})
}
