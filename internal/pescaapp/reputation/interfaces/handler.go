package interfaces

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lriverd/big-service/internal/pescaapp/reputation/application"
	apperrors "github.com/lriverd/big-service/internal/shared/errors"
	"github.com/lriverd/big-service/internal/shared/pagination"
	"github.com/lriverd/big-service/internal/shared/response"
)

const recentEventsLimit = 10

type ReputationHandler struct {
	service        *application.ReputationService
	penaltyService *application.PenaltyEvaluator
}

func NewReputationHandler(service *application.ReputationService, penaltyService *application.PenaltyEvaluator) *ReputationHandler {
	return &ReputationHandler{service: service, penaltyService: penaltyService}
}

func requireSelfOrAdmin(c *gin.Context, targetUserID string) bool {
	userID, _ := c.Get("userID")
	role, _ := c.Get("userRole")
	if userID == targetUserID {
		return true
	}
	if r, ok := role.(string); ok && r == "admin" {
		return true
	}
	response.Error(c, http.StatusForbidden, "FORBIDDEN", "You can only view your own reputation")
	return false
}

func (h *ReputationHandler) GetSummary(c *gin.Context) {
	targetUserID := c.Param("id")
	if !requireSelfOrAdmin(c, targetUserID) {
		return
	}

	score, recent, err := h.service.GetSummary(c.Request.Context(), targetUserID, recentEventsLimit)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			response.Error(c, appErr.Status, appErr.Code, appErr.Message)
			return
		}
		response.InternalError(c, err, "Failed to get reputation summary")
		return
	}
	response.Success(c, gin.H{"score": score, "recentEvents": recent})
}

func (h *ReputationHandler) GetHistory(c *gin.Context) {
	targetUserID := c.Param("id")
	if !requireSelfOrAdmin(c, targetUserID) {
		return
	}

	p := pagination.Parse(c)
	events, total, err := h.service.ListHistory(c.Request.Context(), targetUserID, p.Limit, p.Offset)
	if err != nil {
		response.InternalError(c, err, "Failed to get reputation history")
		return
	}
	response.Paginated(c, events, total, p.Limit, p.Offset)
}

func (h *ReputationHandler) GetPenalties(c *gin.Context) {
	targetUserID := c.Param("id")
	if !requireSelfOrAdmin(c, targetUserID) {
		return
	}

	p := pagination.Parse(c)
	penalties, total, err := h.penaltyService.ListByUser(c.Request.Context(), targetUserID, p.Limit, p.Offset)
	if err != nil {
		response.InternalError(c, err, "Failed to get penalties")
		return
	}
	response.Paginated(c, penalties, total, p.Limit, p.Offset)
}
