package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/lriverd/big-service/internal/platform/middleware"
)

type PaginationResponse struct {
	Total   int  `json:"total"`
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	HasMore bool `json:"hasMore"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func Paginated(c *gin.Context, data interface{}, total, limit, offset int) {
	c.JSON(http.StatusOK, gin.H{
		"data": data,
		"pagination": PaginationResponse{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: offset+limit < total,
		},
	})
}

// InternalError logs the real underlying error (with request ID, path and
// method for correlation with the access log) and writes a generic 500
// envelope to the client. Use this — instead of calling Error directly with
// StatusInternalServerError — anywhere a handler hits an error it didn't
// expect, so the cause is never silently dropped.
func InternalError(c *gin.Context, err error, message string) {
	log.WithError(err).WithFields(log.Fields{
		"path":      c.Request.URL.Path,
		"method":    c.Request.Method,
		"requestId": middleware.GetRequestID(c),
	}).Error(message)
	Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// ErrorWithPayload writes an error response that also carries structured
// data (e.g. a list of conflicting resources) alongside the error envelope.
func ErrorWithPayload(c *gin.Context, status int, code, message string, payload interface{}) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
		"data":      payload,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func ErrorWithDetails(c *gin.Context, status int, code, message, details string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
			"details": details,
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
