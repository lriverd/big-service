package domain

import (
	"context"
	"time"
)

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
	UpdateStatus(ctx context.Context, id string, status SpotStatus) (*Spot, error)
	FindByCreatedByUserID(ctx context.Context, userID string, limit, offset int) ([]*Spot, int, error)
	CountCreatedSince(ctx context.Context, userID string, since time.Time) (int, error)
	FindNearbyForDuplicateCheck(ctx context.Context, lat, lng, radiusMeters float64, maxResults int) ([]DuplicateCandidate, error)
}

type SpotSpeciesRepository interface {
	ListBySpot(ctx context.Context, spotID string) ([]SpotSpecies, error)
	SetForSpot(ctx context.Context, spotID string, species []SpotSpecies) error
	DeleteBySpot(ctx context.Context, spotID string) error
}

// DailyLimitProvider is a small, consumer-defined port that lets SpotService
// ask whether a user has a temporary override of the default daily
// spot-creation limit (e.g. applied by a penalty), without depending on the
// full users module.
type DailyLimitProvider interface {
	GetDailySpotLimitOverride(ctx context.Context, userID string) (*int, error)
}

// ReputationRecorder is a small, consumer-defined port that lets SpotService
// log a reputation event (and its score delta) for a spot's owner, without
// depending on the reputation module's domain types.
type ReputationRecorder interface {
	RecordReputationEvent(ctx context.Context, userID, eventType string, delta int, relatedSpotID, reason string) error
}

// ReputationConfig bundles the reputation wiring and the configured deltas
// for admin-driven spot status transitions, keeping NewSpotService's
// parameter list from growing unbounded as more reputation-aware
// transitions are added.
type ReputationConfig struct {
	Recorder      ReputationRecorder
	DeltaVerified int
	DeltaHidden   int
	DeltaDeleted  int
}
