package interfaces

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lriverd/big-service/internal/pescaapp/users/application"
	"github.com/lriverd/big-service/internal/shared/pagination"
	"github.com/lriverd/big-service/internal/shared/response"
)

type FavoriteHandler struct {
	service *application.FavoriteService
}

func NewFavoriteHandler(service *application.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{service: service}
}

func (h *FavoriteHandler) ListFavorites(c *gin.Context) {
	userID := c.Param("id")
	p := pagination.Parse(c)

	favs, total, err := h.service.ListFavorites(c.Request.Context(), userID, p.Limit, p.Offset)
	if err != nil {
		response.InternalError(c, err, "Failed to list favorites")
		return
	}
	response.Paginated(c, favs, total, p.Limit, p.Offset)
}

func (h *FavoriteHandler) AddFavorite(c *gin.Context) {
	userID := c.Param("id")
	spotID := c.Param("spotId")

	authUserID, _ := c.Get("userID")
	if authUserID.(string) != userID {
		response.Error(c, http.StatusForbidden, "FORBIDDEN", "Can only manage own favorites")
		return
	}

	if err := h.service.AddFavorite(c.Request.Context(), userID, spotID); err != nil {
		response.Error(c, http.StatusConflict, "CONFLICT", err.Error())
		return
	}
	response.Success(c, gin.H{
		"isFavorite": true,
		"addedAt":    time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *FavoriteHandler) RemoveFavorite(c *gin.Context) {
	userID := c.Param("id")
	spotID := c.Param("spotId")

	authUserID, _ := c.Get("userID")
	if authUserID.(string) != userID {
		response.Error(c, http.StatusForbidden, "FORBIDDEN", "Can only manage own favorites")
		return
	}

	if err := h.service.RemoveFavorite(c.Request.Context(), userID, spotID); err != nil {
		response.InternalError(c, err, "Failed to remove favorite")
		return
	}
	response.NoContent(c)
}
