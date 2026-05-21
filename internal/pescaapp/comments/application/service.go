package application

import (
	"context"
	"time"

	"github.com/lriverd/big-service/internal/pescaapp/comments/domain"
	spotDomain "github.com/lriverd/big-service/internal/pescaapp/spots/domain"
	apperrors "github.com/lriverd/big-service/internal/shared/errors"
	log "github.com/sirupsen/logrus"
)

type CommentService struct {
	commentRepo  domain.CommentRepository
	likeRepo     domain.CommentLikeRepository
	userProvider domain.UserInfoProvider
	spotRepo     spotDomain.SpotRepository
}

func NewCommentService(
	commentRepo domain.CommentRepository,
	likeRepo domain.CommentLikeRepository,
	userProvider domain.UserInfoProvider,
	spotRepo spotDomain.SpotRepository,
) *CommentService {
	return &CommentService{
		commentRepo:  commentRepo,
		likeRepo:     likeRepo,
		userProvider:  userProvider,
		spotRepo:     spotRepo,
	}
}

func (s *CommentService) ListBySpot(ctx context.Context, spotID string, limit, offset int, sortBy, currentUserID string) ([]*domain.Comment, int, error) {
	comments, total, err := s.commentRepo.ListBySpot(ctx, spotID, limit, offset, sortBy)
	if err != nil {
		return nil, 0, err
	}

	for _, c := range comments {
		userInfo, err := s.userProvider.GetUserInfo(ctx, c.UserID)
		if err == nil {
			c.User = userInfo
		}
		if currentUserID != "" {
			liked, _ := s.likeRepo.Exists(ctx, c.ID, currentUserID)
			c.Liked = liked
		}
	}
	return comments, total, nil
}

func (s *CommentService) Create(ctx context.Context, spotID, userID, text string) (*domain.Comment, error) {
	comment := &domain.Comment{
		SpotID:    spotID,
		UserID:    userID,
		Text:      text,
		Likes:     0,
		CreatedAt: time.Now().UTC(),
	}

	created, err := s.commentRepo.Create(ctx, comment)
	if err != nil {
		return nil, err
	}

	_ = s.spotRepo.UpdateCommentCount(ctx, spotID, 1)
	log.WithFields(log.Fields{"spotId": spotID, "commentId": created.ID}).Info("Comment created")
	return created, nil
}

func (s *CommentService) Update(ctx context.Context, commentID, userID, text string) (*domain.Comment, error) {
	comment, err := s.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		return nil, apperrors.NotFound("Comment")
	}
	if comment.UserID != userID {
		return nil, apperrors.Forbidden("Only the author can edit this comment")
	}
	return s.commentRepo.Update(ctx, commentID, text)
}

func (s *CommentService) Delete(ctx context.Context, commentID, userID, role string) error {
	comment, err := s.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		return apperrors.NotFound("Comment")
	}
	if comment.UserID != userID && role != "admin" {
		return apperrors.Forbidden("Only the author or admin can delete this comment")
	}
	if err := s.commentRepo.SoftDelete(ctx, commentID); err != nil {
		return err
	}
	_ = s.spotRepo.UpdateCommentCount(ctx, comment.SpotID, -1)
	return nil
}

func (s *CommentService) Like(ctx context.Context, commentID, userID string) (int, bool, error) {
	exists, err := s.likeRepo.Exists(ctx, commentID, userID)
	if err != nil {
		return 0, false, err
	}
	if exists {
		return 0, true, apperrors.New(409, "CONFLICT", "Already liked")
	}

	like := &domain.CommentLike{
		CommentID: commentID,
		UserID:    userID,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.likeRepo.Create(ctx, like); err != nil {
		return 0, false, err
	}
	_ = s.commentRepo.IncrementLikes(ctx, commentID, 1)

	comment, _ := s.commentRepo.FindByID(ctx, commentID)
	likes := 0
	if comment != nil {
		likes = comment.Likes
	}
	return likes, true, nil
}

func (s *CommentService) Unlike(ctx context.Context, commentID, userID string) (int, bool, error) {
	if err := s.likeRepo.Delete(ctx, commentID, userID); err != nil {
		return 0, false, err
	}
	_ = s.commentRepo.IncrementLikes(ctx, commentID, -1)

	comment, _ := s.commentRepo.FindByID(ctx, commentID)
	likes := 0
	if comment != nil {
		likes = comment.Likes
	}
	return likes, false, nil
}

