package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// AccessLogger is gin's default access logger with the request ID appended,
// so a line like `[GIN] 500 POST /v1/spots reqId=...` can be matched to the
// detailed error log emitted for that same request (see RequestID and
// internal/shared/response.InternalError).
func AccessLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(p gin.LogFormatterParams) string {
		reqID, _ := p.Keys[requestIDContextKey].(string)
		return fmt.Sprintf("[GIN] %s | %3d | %13v | %15s | %-7s %#v | reqId=%s\n",
			p.TimeStamp.Format("2006/01/02 - 15:04:05"),
			p.StatusCode,
			p.Latency,
			p.ClientIP,
			p.Method,
			p.Path,
			reqID,
		)
	})
}
