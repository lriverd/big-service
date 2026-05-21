package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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

func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
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

