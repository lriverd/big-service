package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type Params struct {
	Limit  int
	Offset int
}

func Parse(c *gin.Context) Params {
	limit := parseIntParam(c, "limit", 20)
	offset := parseIntParam(c, "offset", 0)

	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 1
	}
	if offset < 0 {
		offset = 0
	}

	return Params{Limit: limit, Offset: offset}
}

func parseIntParam(c *gin.Context, key string, defaultVal int) int {
	v := c.Query(key)
	if v == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return i
}

