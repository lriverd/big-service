package domain

import "context"

type RatingRepository interface {
	FindBySpotAndUser(ctx context.Context, spotID, userID string) (*Rating, error)
	ListBySpot(ctx context.Context, spotID string, limit, offset int) ([]*Rating, int, error)
	Create(ctx context.Context, rating *Rating) (*Rating, error)
	Update(ctx context.Context, id string, stars int) (*Rating, error)
	Delete(ctx context.Context, spotID, userID string) error
	GetStats(ctx context.Context, spotID string) (*RatingStats, error)
	CountByUser(ctx context.Context, userID string) (int, error)
}

