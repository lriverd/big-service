package interfaces

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lriverd/big-service/internal/pescaapp/search/application"
	"github.com/lriverd/big-service/internal/shared/response"
)

type SearchHandler struct {
	service *application.SearchService
}

func NewSearchHandler(service *application.SearchService) *SearchHandler {
	return &SearchHandler{service: service}
}

func (h *SearchHandler) Search(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Query parameter 'q' is required")
		return
	}
	searchType := c.DefaultQuery("type", "all")
	limit := 20
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}

	result, err := h.service.Search(c.Request.Context(), q, searchType, limit)
	if err != nil {
		response.InternalError(c, err, "Search failed")
		return
	}
	response.Success(c, gin.H{"results": result})
}
