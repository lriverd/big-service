package interfaces

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lriverd/big-service/internal/pescaapp/moderation/application"
	"github.com/lriverd/big-service/internal/pescaapp/moderation/domain"
	apperrors "github.com/lriverd/big-service/internal/shared/errors"
	"github.com/lriverd/big-service/internal/shared/pagination"
	"github.com/lriverd/big-service/internal/shared/response"
)

type ReportHandler struct {
	service *application.ReportService
}

func NewReportHandler(service *application.ReportService) *ReportHandler {
	return &ReportHandler{service: service}
}

func (h *ReportHandler) Create(c *gin.Context) {
	spotID := c.Param("id")

	var req domain.CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}
	userID, _ := c.Get("userID")

	report, autoHidden, err := h.service.Report(c.Request.Context(), spotID, userID.(string), req.Reason, req.Details)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			response.Error(c, appErr.Status, appErr.Code, appErr.Message)
			return
		}
		response.InternalError(c, err, "Failed to create report")
		return
	}
	response.Created(c, gin.H{"report": report, "spotHidden": autoHidden})
}

func (h *ReportHandler) ListBySpot(c *gin.Context) {
	spotID := c.Param("id")
	p := pagination.Parse(c)

	reports, total, err := h.service.ListBySpot(c.Request.Context(), spotID, p.Limit, p.Offset)
	if err != nil {
		response.InternalError(c, err, "Failed to list reports")
		return
	}
	response.Paginated(c, reports, total, p.Limit, p.Offset)
}
