package interfaces

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lriverd/big-service/internal/pescaapp/users/application"
	"github.com/lriverd/big-service/internal/pescaapp/users/domain"
	apperrors "github.com/lriverd/big-service/internal/shared/errors"
	"github.com/lriverd/big-service/internal/shared/pagination"
	"github.com/lriverd/big-service/internal/shared/response"
)

type UserHandler struct {
	service *application.UserService
}

func NewUserHandler(service *application.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	user, err := h.service.GetUser(c.Request.Context(), id)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			response.Error(c, appErr.Status, appErr.Code, appErr.Message)
			return
		}
		response.InternalError(c, err, "Failed to get user")
		return
	}
	response.Success(c, user)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")
	if userID.(string) != id {
		role, _ := c.Get("userRole")
		if role.(string) != "admin" {
			response.Error(c, http.StatusForbidden, "FORBIDDEN", "Can only update own profile")
			return
		}
	}

	var req domain.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	user, err := h.service.UpdateUser(c.Request.Context(), id, req)
	if err != nil {
		response.InternalError(c, err, "Failed to update user")
		return
	}
	response.Success(c, user)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	p := pagination.Parse(c)
	search := c.Query("search")

	users, total, err := h.service.ListUsers(c.Request.Context(), p.Limit, p.Offset, search)
	if err != nil {
		response.InternalError(c, err, "Failed to list users")
		return
	}
	response.Paginated(c, users, total, p.Limit, p.Offset)
}
