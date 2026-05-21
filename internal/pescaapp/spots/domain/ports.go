package domain

import "context"

type SpotRepository interface {
	FindByID(ctx context.Context, id string) (*Spot, error)
	List(ctx context.Context, limit, offset int, filter SpotFilter) ([]*Spot, int, error)
	Create(ctx context.Context, spot *Spot) (*Spot, error)
	Update(ctx context.Context, id string, req UpdateSpotRequest) (*Spot, error)
	Delete(ctx context.Context, id string) error
	FindNearby(ctx context.Context, lat, lng, radiusKm float64, limit int) ([]*Spot, error)
	IncrementViews(ctx context.Context, id string) error
	UpdateRatingStats(ctx context.Context, id string, avgRating float64, totalRatings int) error
	UpdateCommentCount(ctx context.Context, id string, delta int) error
	Search(ctx context.Context, query string, limit int) ([]*Spot, error)
}

type SpotSpeciesRepository interface {
	ListBySpot(ctx context.Context, spotID string) ([]SpotSpecies, error)
	SetForSpot(ctx context.Context, spotID string, species []SpotSpecies) error
	DeleteBySpot(ctx context.Context, spotID string) error
}

