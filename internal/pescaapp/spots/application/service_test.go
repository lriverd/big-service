package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lriverd/big-service/internal/pescaapp/spots/application"
	"github.com/lriverd/big-service/internal/pescaapp/spots/domain"
	"github.com/lriverd/big-service/internal/platform/cache"
	apperrors "github.com/lriverd/big-service/internal/shared/errors"
)

type mockSpotRepo struct {
	findByIDFn          func(ctx context.Context, id string) (*domain.Spot, error)
	listFn              func(ctx context.Context, limit, offset int, filter domain.SpotFilter) ([]*domain.Spot, int, error)
	createFn            func(ctx context.Context, spot *domain.Spot) (*domain.Spot, error)
	updateFn            func(ctx context.Context, id string, req domain.UpdateSpotRequest) (*domain.Spot, error)
	deleteFn            func(ctx context.Context, id string) error
	findNearbyFn        func(ctx context.Context, lat, lng, radiusKm float64, limit int) ([]*domain.Spot, error)
	incrementViewsFn    func(ctx context.Context, id string) error
	updateRatingStatsFn func(ctx context.Context, id string, avgRating float64, totalRatings int) error
	updateCommentCntFn  func(ctx context.Context, id string, delta int) error
	searchFn            func(ctx context.Context, query string, limit int) ([]*domain.Spot, error)
	updateStatusFn      func(ctx context.Context, id string, status domain.SpotStatus) (*domain.Spot, error)
	findByCreatorFn     func(ctx context.Context, userID string, limit, offset int) ([]*domain.Spot, int, error)
	countCreatedSinceFn func(ctx context.Context, userID string, since time.Time) (int, error)
	duplicateCheckFn    func(ctx context.Context, lat, lng, radiusMeters float64, maxResults int) ([]domain.DuplicateCandidate, error)
}

func (m *mockSpotRepo) FindByID(ctx context.Context, id string) (*domain.Spot, error) {
	return m.findByIDFn(ctx, id)
}
func (m *mockSpotRepo) List(ctx context.Context, limit, offset int, filter domain.SpotFilter) ([]*domain.Spot, int, error) {
	return m.listFn(ctx, limit, offset, filter)
}
func (m *mockSpotRepo) Create(ctx context.Context, spot *domain.Spot) (*domain.Spot, error) {
	if m.createFn != nil {
		return m.createFn(ctx, spot)
	}
	spot.ID = "generated-id"
	return spot, nil
}
func (m *mockSpotRepo) Update(ctx context.Context, id string, req domain.UpdateSpotRequest) (*domain.Spot, error) {
	return m.updateFn(ctx, id, req)
}
func (m *mockSpotRepo) Delete(ctx context.Context, id string) error { return m.deleteFn(ctx, id) }
func (m *mockSpotRepo) FindNearby(ctx context.Context, lat, lng, radiusKm float64, limit int) ([]*domain.Spot, error) {
	return m.findNearbyFn(ctx, lat, lng, radiusKm, limit)
}
func (m *mockSpotRepo) IncrementViews(ctx context.Context, id string) error {
	if m.incrementViewsFn != nil {
		return m.incrementViewsFn(ctx, id)
	}
	return nil
}
func (m *mockSpotRepo) UpdateRatingStats(ctx context.Context, id string, avgRating float64, totalRatings int) error {
	return m.updateRatingStatsFn(ctx, id, avgRating, totalRatings)
}
func (m *mockSpotRepo) UpdateCommentCount(ctx context.Context, id string, delta int) error {
	return m.updateCommentCntFn(ctx, id, delta)
}
func (m *mockSpotRepo) Search(ctx context.Context, query string, limit int) ([]*domain.Spot, error) {
	return m.searchFn(ctx, query, limit)
}
func (m *mockSpotRepo) UpdateStatus(ctx context.Context, id string, status domain.SpotStatus) (*domain.Spot, error) {
	return m.updateStatusFn(ctx, id, status)
}
func (m *mockSpotRepo) FindByCreatedByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Spot, int, error) {
	return m.findByCreatorFn(ctx, userID, limit, offset)
}
func (m *mockSpotRepo) CountCreatedSince(ctx context.Context, userID string, since time.Time) (int, error) {
	if m.countCreatedSinceFn != nil {
		return m.countCreatedSinceFn(ctx, userID, since)
	}
	return 0, nil
}
func (m *mockSpotRepo) FindNearbyForDuplicateCheck(ctx context.Context, lat, lng, radiusMeters float64, maxResults int) ([]domain.DuplicateCandidate, error) {
	if m.duplicateCheckFn != nil {
		return m.duplicateCheckFn(ctx, lat, lng, radiusMeters, maxResults)
	}
	return nil, nil
}

type mockSpotSpeciesRepo struct{}

func (m *mockSpotSpeciesRepo) ListBySpot(ctx context.Context, spotID string) ([]domain.SpotSpecies, error) {
	return nil, nil
}
func (m *mockSpotSpeciesRepo) SetForSpot(ctx context.Context, spotID string, species []domain.SpotSpecies) error {
	return nil
}
func (m *mockSpotSpeciesRepo) DeleteBySpot(ctx context.Context, spotID string) error { return nil }

type mockDailyLimitProvider struct {
	overrideFn func(ctx context.Context, userID string) (*int, error)
}

func (m *mockDailyLimitProvider) GetDailySpotLimitOverride(ctx context.Context, userID string) (*int, error) {
	if m.overrideFn != nil {
		return m.overrideFn(ctx, userID)
	}
	return nil, nil
}

func TestSpotService_Create_AllowsUnderDailyLimit(t *testing.T) {
	repo := &mockSpotRepo{
		countCreatedSinceFn: func(ctx context.Context, userID string, since time.Time) (int, error) {
			return 2, nil
		},
	}
	svc := application.NewSpotService(repo, &mockSpotSpeciesRepo{}, cache.New(time.Minute, time.Minute), 3, &mockDailyLimitProvider{}, 50, 5, domain.ReputationConfig{})

	spot, err := svc.Create(context.Background(), domain.CreateSpotRequest{Name: "Test"}, "u1", "u1@test.com", "User One")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spot.ID != "generated-id" {
		t.Errorf("expected created spot, got %+v", spot)
	}
}

func TestSpotService_Create_BlocksAtDailyLimit(t *testing.T) {
	repo := &mockSpotRepo{
		countCreatedSinceFn: func(ctx context.Context, userID string, since time.Time) (int, error) {
			return 3, nil
		},
	}
	svc := application.NewSpotService(repo, &mockSpotSpeciesRepo{}, cache.New(time.Minute, time.Minute), 3, &mockDailyLimitProvider{}, 50, 5, domain.ReputationConfig{})

	_, err := svc.Create(context.Background(), domain.CreateSpotRequest{Name: "Test"}, "u1", "u1@test.com", "User One")
	if err == nil {
		t.Fatal("expected daily limit error, got nil")
	}
	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("expected *apperrors.AppError, got %T", err)
	}
	if appErr.Code != "DAILY_LIMIT_EXCEEDED" || appErr.Status != 429 {
		t.Errorf("unexpected error shape: %+v", appErr)
	}
}

func TestSpotService_Create_UsesActivePenaltyOverrideInsteadOfDefault(t *testing.T) {
	overriddenLimit := 1
	repo := &mockSpotRepo{
		countCreatedSinceFn: func(ctx context.Context, userID string, since time.Time) (int, error) {
			return 1, nil // below the default of 3, but at the penalized override of 1
		},
	}
	provider := &mockDailyLimitProvider{
		overrideFn: func(ctx context.Context, userID string) (*int, error) {
			return &overriddenLimit, nil
		},
	}
	svc := application.NewSpotService(repo, &mockSpotSpeciesRepo{}, cache.New(time.Minute, time.Minute), 3, provider, 50, 5, domain.ReputationConfig{})

	_, err := svc.Create(context.Background(), domain.CreateSpotRequest{Name: "Test"}, "u1", "u1@test.com", "User One")
	if err == nil {
		t.Fatal("expected the penalized override limit to apply instead of the default")
	}
	appErr, ok := err.(*apperrors.AppError)
	if !ok || appErr.Code != "DAILY_LIMIT_EXCEEDED" {
		t.Fatalf("expected DAILY_LIMIT_EXCEEDED, got %v", err)
	}
}

func TestSpotService_Create_BlocksWhenTooCloseToExistingSpot(t *testing.T) {
	repo := &mockSpotRepo{
		duplicateCheckFn: func(ctx context.Context, lat, lng, radiusMeters float64, maxResults int) ([]domain.DuplicateCandidate, error) {
			return []domain.DuplicateCandidate{
				{Spot: &domain.Spot{ID: "existing-1", Name: "Rio Existente"}, DistanceMeters: 12},
			}, nil
		},
	}
	svc := application.NewSpotService(repo, &mockSpotSpeciesRepo{}, cache.New(time.Minute, time.Minute), 3, &mockDailyLimitProvider{}, 50, 5, domain.ReputationConfig{})

	_, err := svc.Create(context.Background(), domain.CreateSpotRequest{Name: "Test"}, "u1", "u1@test.com", "User One")
	if err == nil {
		t.Fatal("expected a duplicate-candidates error, got nil")
	}
	dupErr, ok := err.(*domain.DuplicateCandidatesError)
	if !ok {
		t.Fatalf("expected *domain.DuplicateCandidatesError, got %T", err)
	}
	if len(dupErr.Candidates) != 1 || dupErr.Candidates[0].Spot.ID != "existing-1" {
		t.Errorf("unexpected candidates: %+v", dupErr.Candidates)
	}
}

func TestSpotService_Create_AllowsWhenNoSpotWithinMinDistance(t *testing.T) {
	repo := &mockSpotRepo{
		duplicateCheckFn: func(ctx context.Context, lat, lng, radiusMeters float64, maxResults int) ([]domain.DuplicateCandidate, error) {
			return nil, nil
		},
	}
	svc := application.NewSpotService(repo, &mockSpotSpeciesRepo{}, cache.New(time.Minute, time.Minute), 3, &mockDailyLimitProvider{}, 50, 5, domain.ReputationConfig{})

	spot, err := svc.Create(context.Background(), domain.CreateSpotRequest{Name: "Test"}, "u1", "u1@test.com", "User One")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spot.ID != "generated-id" {
		t.Errorf("expected spot to be created, got %+v", spot)
	}
}

func TestSpotService_Create_FailsClosedWhenDistanceCheckErrors(t *testing.T) {
	repo := &mockSpotRepo{
		duplicateCheckFn: func(ctx context.Context, lat, lng, radiusMeters float64, maxResults int) ([]domain.DuplicateCandidate, error) {
			return nil, errors.New("firestore unavailable")
		},
	}
	svc := application.NewSpotService(repo, &mockSpotSpeciesRepo{}, cache.New(time.Minute, time.Minute), 3, &mockDailyLimitProvider{}, 50, 5, domain.ReputationConfig{})

	_, err := svc.Create(context.Background(), domain.CreateSpotRequest{Name: "Test"}, "u1", "u1@test.com", "User One")
	if err == nil {
		t.Fatal("expected creation to fail closed when the minimum-distance check itself errors")
	}
	if _, ok := err.(*domain.DuplicateCandidatesError); ok {
		t.Error("a check failure should not be reported as a duplicate-candidates error")
	}
}

func TestSpotService_FindDuplicateCandidates_ClampsExcessiveRadiusOverride(t *testing.T) {
	var receivedRadius float64
	repo := &mockSpotRepo{
		duplicateCheckFn: func(ctx context.Context, lat, lng, radiusMeters float64, maxResults int) ([]domain.DuplicateCandidate, error) {
			receivedRadius = radiusMeters
			return nil, nil
		},
	}
	svc := application.NewSpotService(repo, &mockSpotSpeciesRepo{}, cache.New(time.Minute, time.Minute), 3, &mockDailyLimitProvider{}, 50, 5, domain.ReputationConfig{})

	hugeRadius := 100000.0
	_, err := svc.FindDuplicateCandidates(context.Background(), -33.0, -70.0, &hugeRadius)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	const maxAllowed = 50 * 5
	if receivedRadius != maxAllowed {
		t.Errorf("expected radius clamped to %v, got %v", maxAllowed, receivedRadius)
	}
}

type fakeReputationRecorder struct {
	calls []reputationCall
}

type reputationCall struct {
	userID, eventType string
	delta             int
}

func (f *fakeReputationRecorder) RecordReputationEvent(ctx context.Context, userID, eventType string, delta int, relatedSpotID, reason string) error {
	f.calls = append(f.calls, reputationCall{userID: userID, eventType: eventType, delta: delta})
	return nil
}

func TestSpotService_UpdateStatusByAdmin_VerifyingAPendingSpotRewardsOwner(t *testing.T) {
	recorder := &fakeReputationRecorder{}
	repo := &mockSpotRepo{
		findByIDFn: func(ctx context.Context, id string) (*domain.Spot, error) {
			return &domain.Spot{ID: id, CreatedByUserID: "owner-1", Status: string(domain.SpotStatusPending)}, nil
		},
		updateStatusFn: func(ctx context.Context, id string, status domain.SpotStatus) (*domain.Spot, error) {
			return &domain.Spot{ID: id, CreatedByUserID: "owner-1", Status: string(status)}, nil
		},
	}
	svc := application.NewSpotService(repo, &mockSpotSpeciesRepo{}, cache.New(time.Minute, time.Minute), 3, &mockDailyLimitProvider{}, 50, 5, domain.ReputationConfig{
		Recorder: recorder, DeltaVerified: 10, DeltaHidden: -15, DeltaDeleted: -20,
	})

	_, err := svc.UpdateStatusByAdmin(context.Background(), "spot-1", domain.SpotStatusVerified, "looks good")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recorder.calls) != 1 || recorder.calls[0].eventType != "SPOT_VERIFIED" || recorder.calls[0].delta != 10 {
		t.Errorf("expected a SPOT_VERIFIED +10 event, got %+v", recorder.calls)
	}
}

func TestSpotService_UpdateStatusByAdmin_RestoringFromHiddenDoesNotRewardOwner(t *testing.T) {
	recorder := &fakeReputationRecorder{}
	repo := &mockSpotRepo{
		findByIDFn: func(ctx context.Context, id string) (*domain.Spot, error) {
			return &domain.Spot{ID: id, CreatedByUserID: "owner-1", Status: string(domain.SpotStatusHidden)}, nil
		},
		updateStatusFn: func(ctx context.Context, id string, status domain.SpotStatus) (*domain.Spot, error) {
			return &domain.Spot{ID: id, CreatedByUserID: "owner-1", Status: string(status)}, nil
		},
	}
	svc := application.NewSpotService(repo, &mockSpotSpeciesRepo{}, cache.New(time.Minute, time.Minute), 3, &mockDailyLimitProvider{}, 50, 5, domain.ReputationConfig{
		Recorder: recorder, DeltaVerified: 10, DeltaHidden: -15, DeltaDeleted: -20,
	})

	_, err := svc.UpdateStatusByAdmin(context.Background(), "spot-1", domain.SpotStatusVerified, "restoring")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recorder.calls) != 0 {
		t.Errorf("expected restoring a HIDDEN spot to not fire a reputation event, got %+v", recorder.calls)
	}
}

func TestSpotService_UpdateStatusByAdmin_HidingPenalizesOwner(t *testing.T) {
	recorder := &fakeReputationRecorder{}
	repo := &mockSpotRepo{
		findByIDFn: func(ctx context.Context, id string) (*domain.Spot, error) {
			return &domain.Spot{ID: id, CreatedByUserID: "owner-1", Status: string(domain.SpotStatusVerified)}, nil
		},
		updateStatusFn: func(ctx context.Context, id string, status domain.SpotStatus) (*domain.Spot, error) {
			return &domain.Spot{ID: id, CreatedByUserID: "owner-1", Status: string(status)}, nil
		},
	}
	svc := application.NewSpotService(repo, &mockSpotSpeciesRepo{}, cache.New(time.Minute, time.Minute), 3, &mockDailyLimitProvider{}, 50, 5, domain.ReputationConfig{
		Recorder: recorder, DeltaVerified: 10, DeltaHidden: -15, DeltaDeleted: -20,
	})

	_, err := svc.UpdateStatusByAdmin(context.Background(), "spot-1", domain.SpotStatusHidden, "spam")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recorder.calls) != 1 || recorder.calls[0].eventType != "SPOT_HIDDEN" || recorder.calls[0].delta != -15 {
		t.Errorf("expected a SPOT_HIDDEN -15 event, got %+v", recorder.calls)
	}
}
