package interfaces

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lriverd/big-service/internal/pescaapp/species/application"
	"github.com/lriverd/big-service/internal/pescaapp/species/domain"
	apperrors "github.com/lriverd/big-service/internal/shared/errors"
	"github.com/lriverd/big-service/internal/shared/pagination"
	"github.com/lriverd/big-service/internal/shared/response"
)

type SpeciesHandler struct {
	service *application.SpeciesService
}

func NewSpeciesHandler(service *application.SpeciesService) *SpeciesHandler {
	return &SpeciesHandler{service: service}
}

func (h *SpeciesHandler) List(c *gin.Context) {
	p := pagination.Parse(c)
	search := c.Query("search")

	species, total, err := h.service.List(c.Request.Context(), p.Limit, p.Offset, search)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list species")
		return
	}
	response.Paginated(c, species, total, p.Limit, p.Offset)
}

func (h *SpeciesHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	sp, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			response.Error(c, appErr.Status, appErr.Code, appErr.Message)
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get species")
		return
	}
	response.Success(c, sp)
}

func (h *SpeciesHandler) Create(c *gin.Context) {
	var req domain.CreateSpeciesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}
	sp, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create species")
		return
	}
	response.Created(c, sp)
}

func (h *SpeciesHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req domain.UpdateSpeciesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}
	sp, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update species")
		return
	}
	response.Success(c, sp)
}

func (h *SpeciesHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete species")
		return
	}
	response.NoContent(c)
}

