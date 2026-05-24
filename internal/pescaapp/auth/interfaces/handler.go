package interfaces

import (
	"net/http"

	"github.com/gin-gonic/gin"
	authDomain "github.com/lriverd/big-service/internal/pescaapp/auth/domain"
	apperrors "github.com/lriverd/big-service/internal/shared/errors"
	"github.com/lriverd/big-service/internal/shared/response"
)

type AuthHandler struct {
	service authDomain.AuthService
}

func NewAuthHandler(service authDomain.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req authDomain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	res, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed")
		return
	}

	response.Created(c, res)
}

func (h *AuthHandler) LoginWithPassword(c *gin.Context) {
	var req authDomain.PasswordLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	res, err := h.service.LoginWithPassword(c.Request.Context(), req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			response.Error(c, appErr.Status, appErr.Code, appErr.Message)
			return
		}
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid email or password")
		return
	}

	response.Created(c, res)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req authDomain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	res, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			response.Error(c, appErr.Status, appErr.Code, appErr.Message)
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Registration failed")
		return
	}

	response.Created(c, res)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, _ := c.Get("userID")
	if err := h.service.Logout(c.Request.Context(), userID.(string)); err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Logout failed")
		return
	}
	response.Success(c, gin.H{"message": "Logged out successfully"})
}

