package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/lriverd/big-service/internal/pescaapp/ratings/application"
	"github.com/lriverd/big-service/internal/pescaapp/ratings/domain"
	spotsDomain "github.com/lriverd/big-service/internal/pescaapp/spots/domain"
)

// fakeSpotRepo implements only what RatingService actually calls
// (FindBySpotAndUser flows through ratingRepo, not this); the rest panic if
// reached, which would signal an unexpected code path in these tests.
type fakeSpotRepo struct {
	spotsDomain.SpotRepository
	findByIDFn func(ctx context.Context, id string) (*spotsDomain.Spot, error)
}

func (f *fakeSpotRepo) FindByID(ctx context.Context, id string) (*spotsDomain.Spot, error) {
	return f.findByIDFn(ctx, id)
}
func (f *fakeSpotRepo) UpdateRatingStats(ctx context.Context, id string, avgRating float64, totalRatings int) error {
	return nil
}

type fakeRatingRepo struct {
	existing *domain.Rating
}

func (f *fakeRatingRepo) FindBySpotAndUser(ctx context.Context, spotID, userID string) (*domain.Rating, error) {
	if f.existing != nil {
		return f.existing, nil
	}
	return nil, context.DeadlineExceeded // any non-nil error signals "not found" to CreateOrUpdate
}
func (f *fakeRatingRepo) ListBySpot(ctx context.Context, spotID string, limit, offset int) ([]*domain.Rating, int, error) {
	return nil, 0, nil
}
func (f *fakeRatingRepo) Create(ctx context.Context, rating *domain.Rating) (*domain.Rating, error) {
	rating.ID = "rating-1"
	return rating, nil
}
func (f *fakeRatingRepo) Update(ctx context.Context, id string, stars int) (*domain.Rating, error) {
	return &domain.Rating{ID: id, Stars: stars}, nil
}
func (f *fakeRatingRepo) Delete(ctx context.Context, spotID, userID string) error { return nil }
func (f *fakeRatingRepo) GetStats(ctx context.Context, spotID string) (*domain.RatingStats, error) {
	return &domain.RatingStats{}, nil
}
func (f *fakeRatingRepo) CountByUser(ctx context.Context, userID string) (int, error) { return 0, nil }

type fakeReputationRecorder struct {
	calls []call
}

type call struct {
	userID, eventType string
	delta             int
}

func (f *fakeReputationRecorder) RecordReputationEvent(ctx context.Context, userID, eventType string, delta int, relatedSpotID, reason string) error {
	f.calls = append(f.calls, call{userID: userID, eventType: eventType, delta: delta})
	return nil
}

func TestRatingService_CreateOrUpdate_RewardsOwnerOnGoodNewRating(t *testing.T) {
	spotRepo := &fakeSpotRepo{findByIDFn: func(ctx context.Context, id string) (*spotsDomain.Spot, error) {
		return &spotsDomain.Spot{ID: id, CreatedByUserID: "owner-1"}, nil
	}}
	recorder := &fakeReputationRecorder{}
	svc := application.NewRatingService(&fakeRatingRepo{}, spotRepo, recorder, 2, 4)

	_, created, err := svc.CreateOrUpdate(context.Background(), "spot-1", "rater-1", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Fatal("expected a new rating to be created")
	}
	if len(recorder.calls) != 1 || recorder.calls[0].userID != "owner-1" || recorder.calls[0].delta != 2 {
		t.Errorf("expected a GOOD_RATING_RECEIVED +2 event for owner-1, got %+v", recorder.calls)
	}
}

func TestRatingService_CreateOrUpdate_DoesNotRewardBelowThreshold(t *testing.T) {
	spotRepo := &fakeSpotRepo{findByIDFn: func(ctx context.Context, id string) (*spotsDomain.Spot, error) {
		return &spotsDomain.Spot{ID: id, CreatedByUserID: "owner-1"}, nil
	}}
	recorder := &fakeReputationRecorder{}
	svc := application.NewRatingService(&fakeRatingRepo{}, spotRepo, recorder, 2, 4)

	_, _, err := svc.CreateOrUpdate(context.Background(), "spot-1", "rater-1", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recorder.calls) != 0 {
		t.Errorf("expected no reputation event below the stars threshold, got %+v", recorder.calls)
	}
}

func TestRatingService_CreateOrUpdate_DoesNotRewardSelfRating(t *testing.T) {
	spotRepo := &fakeSpotRepo{findByIDFn: func(ctx context.Context, id string) (*spotsDomain.Spot, error) {
		return &spotsDomain.Spot{ID: id, CreatedByUserID: "owner-1"}, nil
	}}
	recorder := &fakeReputationRecorder{}
	svc := application.NewRatingService(&fakeRatingRepo{}, spotRepo, recorder, 2, 4)

	_, _, err := svc.CreateOrUpdate(context.Background(), "spot-1", "owner-1", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recorder.calls) != 0 {
		t.Errorf("expected no reputation event for a spot owner rating their own spot, got %+v", recorder.calls)
	}
}

func TestRatingService_CreateOrUpdate_DoesNotRewardOnUpdate(t *testing.T) {
	spotRepo := &fakeSpotRepo{findByIDFn: func(ctx context.Context, id string) (*spotsDomain.Spot, error) {
		return &spotsDomain.Spot{ID: id, CreatedByUserID: "owner-1"}, nil
	}}
	recorder := &fakeReputationRecorder{}
	svc := application.NewRatingService(&fakeRatingRepo{existing: &domain.Rating{ID: "r1", SpotID: "spot-1", UserID: "rater-1", Stars: 3, CreatedAt: time.Now(), UpdatedAt: time.Now()}}, spotRepo, recorder, 2, 4)

	_, created, err := svc.CreateOrUpdate(context.Background(), "spot-1", "rater-1", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created {
		t.Fatal("expected an update, not a creation")
	}
	if len(recorder.calls) != 0 {
		t.Errorf("expected no reputation event for updating an existing rating, got %+v", recorder.calls)
	}
}
