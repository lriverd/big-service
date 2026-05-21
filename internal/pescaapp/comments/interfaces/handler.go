package interfaces

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lriverd/big-service/internal/pescaapp/comments/application"
	"github.com/lriverd/big-service/internal/pescaapp/comments/domain"
	apperrors "github.com/lriverd/big-service/internal/shared/errors"
	"github.com/lriverd/big-service/internal/shared/pagination"
	"github.com/lriverd/big-service/internal/shared/response"
)

type CommentHandler struct {
	service *application.CommentService
}

func NewCommentHandler(service *application.CommentService) *CommentHandler {
	return &CommentHandler{service: service}
}

func (h *CommentHandler) ListBySpot(c *gin.Context) {
	spotID := c.Param("id")
	p := pagination.Parse(c)
	sortBy := c.DefaultQuery("sortBy", "recent")
	currentUserID, _ := c.Get("userID")
	uid := ""
	if currentUserID != nil {
		uid = currentUserID.(string)
	}

	comments, total, err := h.service.ListBySpot(c.Request.Context(), spotID, p.Limit, p.Offset, sortBy, uid)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list comments")
		return
	}
	response.Paginated(c, comments, total, p.Limit, p.Offset)
}

func (h *CommentHandler) Create(c *gin.Context) {
	spotID := c.Param("id")
	var req domain.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Text must be between 10 and 500 characters")
		return
	}
	userID, _ := c.Get("userID")
	comment, err := h.service.Create(c.Request.Context(), spotID, userID.(string), req.Text)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create comment")
		return
	}
	response.Created(c, comment)
}

func (h *CommentHandler) Update(c *gin.Context) {
	commentID := c.Param("commentId")
	var req domain.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}
	userID, _ := c.Get("userID")
	comment, err := h.service.Update(c.Request.Context(), commentID, userID.(string), req.Text)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			response.Error(c, appErr.Status, appErr.Code, appErr.Message)
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update comment")
		return
	}
	response.Success(c, comment)
}

func (h *CommentHandler) Delete(c *gin.Context) {
	commentID := c.Param("commentId")
	userID, _ := c.Get("userID")
	role, _ := c.Get("userRole")
	if err := h.service.Delete(c.Request.Context(), commentID, userID.(string), role.(string)); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			response.Error(c, appErr.Status, appErr.Code, appErr.Message)
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete comment")
		return
	}
	response.NoContent(c)
}

func (h *CommentHandler) Like(c *gin.Context) {
	commentID := c.Param("commentId")
	userID, _ := c.Get("userID")
	likes, liked, err := h.service.Like(c.Request.Context(), commentID, userID.(string))
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			response.Error(c, appErr.Status, appErr.Code, appErr.Message)
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to like comment")
		return
	}
	response.Success(c, gin.H{"likes": likes, "liked": liked})
}

func (h *CommentHandler) Unlike(c *gin.Context) {
	commentID := c.Param("commentId")
	userID, _ := c.Get("userID")
	likes, liked, err := h.service.Unlike(c.Request.Context(), commentID, userID.(string))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to unlike comment")
		return
	}
	response.Success(c, gin.H{"likes": likes, "liked": liked})
}

