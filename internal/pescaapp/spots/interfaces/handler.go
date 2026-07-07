package interfaces

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lriverd/big-service/internal/pescaapp/spots/application"
	"github.com/lriverd/big-service/internal/pescaapp/spots/domain"
	apperrors "github.com/lriverd/big-service/internal/shared/errors"
	"github.com/lriverd/big-service/internal/shared/pagination"
	"github.com/lriverd/big-service/internal/shared/response"
)

type SpotHandler struct {
	service *application.SpotService
}

func NewSpotHandler(service *application.SpotService) *SpotHandler {
	return &SpotHandler{service: service}
}

func (h *SpotHandler) List(c *gin.Context) {
	p := pagination.Parse(c)
	filter := domain.SpotFilter{
		Region:    c.Query("region"),
		WaterType: c.Query("waterType"),
		SortBy:    c.Query("sortBy"),
	}

	if br := c.Query("boatRequired"); br != "" {
		val := br == "true"
		filter.BoatRequired = &val
	}
	if lat := c.Query("latitude"); lat != "" {
		if v, err := strconv.ParseFloat(lat, 64); err == nil {
			filter.Latitude = &v
		}
	}
	if lng := c.Query("longitude"); lng != "" {
		if v, err := strconv.ParseFloat(lng, 64); err == nil {
			filter.Longitude = &v
		}
	}
	if r := c.Query("radius"); r != "" {
		if v, err := strconv.ParseFloat(r, 64); err == nil {
			filter.RadiusKm = &v
		}
	}

	spots, total, err := h.service.List(c.Request.Context(), p.Limit, p.Offset, filter)
	if err != nil {
		response.InternalError(c, err, "Failed to list spots")
		return
	}
	response.Paginated(c, spots, total, p.Limit, p.Offset)
}

func (h *SpotHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	spot, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			response.Error(c, appErr.Status, appErr.Code, appErr.Message)
			return
		}
		response.InternalError(c, err, "Failed to get spot")
		return
	}
	response.Success(c, spot)
}

func (h *SpotHandler) Create(c *gin.Context) {
	var req domain.CreateSpotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}
	userID, _ := c.Get("userID")

	userEmail, _ := c.Get("userEmail")
	userName, _ := c.Get("userName")

	email := ""
	if userEmail != nil {
		email = userEmail.(string)
	}
	name := ""
	if userName != nil {
		name = userName.(string)
	}

	spot, err := h.service.Create(c.Request.Context(), req, userID.(string), email, name)
	if err != nil {
		if dupErr, ok := err.(*domain.DuplicateCandidatesError); ok {
			response.ErrorWithPayload(c, http.StatusConflict, "TOO_CLOSE_TO_EXISTING_SPOT",
				"Ya existe un spot a menos de la distancia mínima permitida de esta ubicación. Elige un punto distinto.",
				dupErr.Candidates)
			return
		}
		if appErr, ok := err.(*apperrors.AppError); ok {
			response.Error(c, appErr.Status, appErr.Code, appErr.Message)
			return
		}
		response.InternalError(c, err, "Failed to create spot")
		return
	}
	response.Created(c, spot)
}

func (h *SpotHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req domain.UpdateSpotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}
	userID, _ := c.Get("userID")
	role, _ := c.Get("userRole")
	spot, err := h.service.Update(c.Request.Context(), id, req, userID.(string), role.(string))
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			response.Error(c, appErr.Status, appErr.Code, appErr.Message)
			return
		}
		response.InternalError(c, err, "Failed to update spot")
		return
	}
	response.Success(c, spot)
}

func (h *SpotHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")
	role, _ := c.Get("userRole")
	if err := h.service.Delete(c.Request.Context(), id, userID.(string), role.(string)); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			response.Error(c, appErr.Status, appErr.Code, appErr.Message)
			return
		}
		response.InternalError(c, err, "Failed to delete spot")
		return
	}
	response.NoContent(c)
}

func (h *SpotHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req domain.UpdateSpotStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}
	switch req.Status {
	case domain.SpotStatusVerified, domain.SpotStatusHidden, domain.SpotStatusDeleted:
	default:
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "status must be one of VERIFIED, HIDDEN, DELETED")
		return
	}

	spot, err := h.service.UpdateStatusByAdmin(c.Request.Context(), id, req.Status, req.Reason)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			response.Error(c, appErr.Status, appErr.Code, appErr.Message)
			return
		}
		response.InternalError(c, err, "Failed to update spot status")
		return
	}
	response.Success(c, spot)
}

func (h *SpotHandler) NearbyDuplicates(c *gin.Context) {
	lat, latErr := strconv.ParseFloat(c.Query("lat"), 64)
	lng, lngErr := strconv.ParseFloat(c.Query("lng"), 64)
	if latErr != nil || lngErr != nil {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "lat and lng query parameters are required")
		return
	}

	var radiusOverride *float64
	if r := c.Query("radius"); r != "" {
		if v, err := strconv.ParseFloat(r, 64); err == nil {
			radiusOverride = &v
		}
	}

	candidates, err := h.service.FindDuplicateCandidates(c.Request.Context(), lat, lng, radiusOverride)
	if err != nil {
		response.InternalError(c, err, "Failed to search for nearby duplicates")
		return
	}
	response.Success(c, gin.H{"data": candidates})
}

func (h *SpotHandler) Mine(c *gin.Context) {
	p := pagination.Parse(c)
	userID, _ := c.Get("userID")

	spots, total, err := h.service.Mine(c.Request.Context(), userID.(string), p.Limit, p.Offset)
	if err != nil {
		response.InternalError(c, err, "Failed to list your spots")
		return
	}
	response.Paginated(c, spots, total, p.Limit, p.Offset)
}

func (h *SpotHandler) FindNearby(c *gin.Context) {
	spotID := c.Param("id")
	radiusKm := 10.0
	limit := 10

	if r := c.Query("radiusKm"); r != "" {
		if v, err := strconv.ParseFloat(r, 64); err == nil {
			radiusKm = v
		}
	}
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}

	spots, err := h.service.FindNearby(c.Request.Context(), spotID, radiusKm, limit)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			response.Error(c, appErr.Status, appErr.Code, appErr.Message)
			return
		}
		response.InternalError(c, err, "Failed to find nearby spots")
		return
	}
	response.Success(c, gin.H{"data": spots})
}
