package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-Id"
const requestIDContextKey = "requestId"

// RequestID assigns a short unique ID to every request (reusing one the
// caller already provides via the X-Request-Id header, if present) and
// echoes it back in the response header. Every log line emitted while
// handling the request should include this ID, so a generic access-log
// line like `[GIN] 500 POST /v1/spots` can be matched to the detailed
// error log that explains why.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(requestIDContextKey, id)
		c.Writer.Header().Set(RequestIDHeader, id)
		c.Next()
	}
}

// GetRequestID reads the ID assigned by RequestID, if any.
func GetRequestID(c *gin.Context) string {
	if id, ok := c.Get(requestIDContextKey); ok {
		if s, ok := id.(string); ok {
			return s
		}
	}
	return ""
}
