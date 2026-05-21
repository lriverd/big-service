package domain

import "context"

type CommentRepository interface {
	FindByID(ctx context.Context, id string) (*Comment, error)
	ListBySpot(ctx context.Context, spotID string, limit, offset int, sortBy string) ([]*Comment, int, error)
	Create(ctx context.Context, comment *Comment) (*Comment, error)
	Update(ctx context.Context, id string, text string) (*Comment, error)
	SoftDelete(ctx context.Context, id string) error
	IncrementLikes(ctx context.Context, id string, delta int) error
}

type CommentLikeRepository interface {
	Exists(ctx context.Context, commentID, userID string) (bool, error)
	Create(ctx context.Context, like *CommentLike) error
	Delete(ctx context.Context, commentID, userID string) error
}

type UserInfoProvider interface {
	GetUserInfo(ctx context.Context, userID string) (*UserInfo, error)
}

