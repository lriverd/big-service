package interfaces

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lriverd/big-service/internal/pescaapp/ratings/application"
	"github.com/lriverd/big-service/internal/pescaapp/ratings/domain"
	apperrors "github.com/lriverd/big-service/internal/shared/errors"
	"github.com/lriverd/big-service/internal/shared/pagination"
	"github.com/lriverd/big-service/internal/shared/response"
)

type RatingHandler struct {
	service *application.RatingService
}

func NewRatingHandler(service *application.RatingService) *RatingHandler {
	return &RatingHandler{service: service}
}

func (h *RatingHandler) ListBySpot(c *gin.Context) {
	spotID := c.Param("id")
	p := pagination.Parse(c)

	ratings, total, stats, err := h.service.ListBySpot(c.Request.Context(), spotID, p.Limit, p.Offset)
	if err != nil {
		response.InternalError(c, err, "Failed to list ratings")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  ratings,
		"stats": stats,
		"pagination": gin.H{
			"total":   total,
			"limit":   p.Limit,
			"offset":  p.Offset,
			"hasMore": p.Offset+p.Limit < total,
		},
	})
}

func (h *RatingHandler) CreateOrUpdate(c *gin.Context) {
	spotID := c.Param("id")
	var req domain.CreateRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Stars must be between 1 and 5")
		return
	}
	userID, _ := c.Get("userID")
	rating, created, err := h.service.CreateOrUpdate(c.Request.Context(), spotID, userID.(string), req.Stars)
	if err != nil {
		response.InternalError(c, err, "Failed to create rating")
		return
	}
	if created {
		response.Created(c, rating)
	} else {
		response.Success(c, rating)
	}
}

func (h *RatingHandler) Delete(c *gin.Context) {
	spotID := c.Param("id")
	userID, _ := c.Get("userID")
	if err := h.service.Delete(c.Request.Context(), spotID, userID.(string)); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			response.Error(c, appErr.Status, appErr.Code, appErr.Message)
			return
		}
		response.InternalError(c, err, "Failed to delete rating")
		return
	}
	response.NoContent(c)
}
