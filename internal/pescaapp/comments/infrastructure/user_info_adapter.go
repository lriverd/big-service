package infrastructure

import (
	"context"

	"github.com/lriverd/big-service/internal/pescaapp/comments/domain"
	userDomain "github.com/lriverd/big-service/internal/pescaapp/users/domain"
)

// UserInfoAdapter adapts UserRepository to CommentService's UserInfoProvider
type UserInfoAdapter struct {
	userRepo userDomain.UserRepository
}

func NewUserInfoAdapter(userRepo userDomain.UserRepository) *UserInfoAdapter {
	return &UserInfoAdapter{userRepo: userRepo}
}

func (a *UserInfoAdapter) GetUserInfo(ctx context.Context, userID string) (*domain.UserInfo, error) {
	user, err := a.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &domain.UserInfo{
		ID:       user.ID,
		Name:     user.Name,
		PhotoURL: user.PhotoURL,
	}, nil
}

