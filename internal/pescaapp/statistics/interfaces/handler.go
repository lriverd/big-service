package interfaces

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lriverd/big-service/internal/pescaapp/statistics/application"
	apperrors "github.com/lriverd/big-service/internal/shared/errors"
	"github.com/lriverd/big-service/internal/shared/response"
)

type StatsHandler struct {
	service *application.StatsService
}

func NewStatsHandler(service *application.StatsService) *StatsHandler {
	return &StatsHandler{service: service}
}

func (h *StatsHandler) GetSpotStats(c *gin.Context) {
	spotID := c.Param("id")
	stats, err := h.service.GetSpotStats(c.Request.Context(), spotID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			response.Error(c, appErr.Status, appErr.Code, appErr.Message)
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get spot stats")
		return
	}
	response.Success(c, stats)
}

func (h *StatsHandler) GetUserStats(c *gin.Context) {
	userID := c.Param("userId")
	stats, err := h.service.GetUserStats(c.Request.Context(), userID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			response.Error(c, appErr.Status, appErr.Code, appErr.Message)
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get user stats")
		return
	}
	response.Success(c, stats)
}

func (h *StatsHandler) GetPopularSpots(c *gin.Context) {
	limit := 10
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}
	orderBy := c.DefaultQuery("orderBy", "rating")

	spots, err := h.service.GetPopularSpots(c.Request.Context(), limit, orderBy)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get popular spots")
		return
	}
	response.Success(c, gin.H{"data": spots})
}

