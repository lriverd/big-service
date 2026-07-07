package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/lriverd/big-service/internal/pescaapp/reputation/application"
	"github.com/lriverd/big-service/internal/pescaapp/reputation/domain"
)

type fakePenaltyRepo struct {
	penalties []*domain.Penalty
}

func (f *fakePenaltyRepo) Create(ctx context.Context, penalty *domain.Penalty) (*domain.Penalty, error) {
	penalty.ID = "penalty-1"
	f.penalties = append(f.penalties, penalty)
	return penalty, nil
}

func (f *fakePenaltyRepo) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Penalty, int, error) {
	var matched []*domain.Penalty
	for _, p := range f.penalties {
		if p.UserID == userID {
			matched = append(matched, p)
		}
	}
	return matched, len(matched), nil
}

func (f *fakePenaltyRepo) HasActivePenaltyOfType(ctx context.Context, userID, penaltyType string, now time.Time) (bool, error) {
	for _, p := range f.penalties {
		if p.UserID == userID && p.Type == penaltyType && p.IsActive(now) {
			return true, nil
		}
	}
	return false, nil
}

type fakeUserPenaltyApplier struct {
	calls int
	limit int
}

func (f *fakeUserPenaltyApplier) SetDailySpotLimitOverride(ctx context.Context, userID string, limit int, expiresAt time.Time) error {
	f.calls++
	f.limit = limit
	return nil
}

func seedJudgedEvents(repo *fakeReputationRepo, userID string, verifiedCount, rejectedCount int) {
	for i := 0; i < verifiedCount; i++ {
		repo.events = append(repo.events, &domain.ReputationEvent{UserID: userID, EventType: string(domain.EventSpotVerified)})
	}
	for i := 0; i < rejectedCount; i++ {
		repo.events = append(repo.events, &domain.ReputationEvent{UserID: userID, EventType: string(domain.EventSpotHidden)})
	}
}

func TestPenaltyEvaluator_SkipsBelowMinSampleSize(t *testing.T) {
	repo := &fakeReputationRepo{}
	seedJudgedEvents(repo, "u1", 0, 2) // 100% rejected, but only 2 judged spots
	penaltyRepo := &fakePenaltyRepo{}
	applier := &fakeUserPenaltyApplier{}
	eval := application.NewPenaltyEvaluator(repo, newFakeScoreWriter(), penaltyRepo, applier, 0.5, 3, -25, 1, 7*24*time.Hour)

	if err := eval.EvaluateAfterRejection(context.Background(), "u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(penaltyRepo.penalties) != 0 {
		t.Errorf("expected no penalty below min sample size, got %d", len(penaltyRepo.penalties))
	}
	if applier.calls != 0 {
		t.Errorf("expected no daily limit override below min sample size, got %d calls", applier.calls)
	}
}

func TestPenaltyEvaluator_SkipsBelowRateThreshold(t *testing.T) {
	repo := &fakeReputationRepo{}
	seedJudgedEvents(repo, "u1", 8, 2) // 20% rejected, below 50% threshold, but plenty of samples
	penaltyRepo := &fakePenaltyRepo{}
	eval := application.NewPenaltyEvaluator(repo, newFakeScoreWriter(), penaltyRepo, &fakeUserPenaltyApplier{}, 0.5, 3, -25, 1, 7*24*time.Hour)

	if err := eval.EvaluateAfterRejection(context.Background(), "u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(penaltyRepo.penalties) != 0 {
		t.Errorf("expected no penalty below rate threshold, got %d", len(penaltyRepo.penalties))
	}
}

func TestPenaltyEvaluator_AppliesPenaltyAtThreshold(t *testing.T) {
	repo := &fakeReputationRepo{}
	seedJudgedEvents(repo, "u1", 2, 2) // 50% rejected, meets threshold, 4 judged >= min sample 3
	penaltyRepo := &fakePenaltyRepo{}
	scoreWriter := newFakeScoreWriter()
	applier := &fakeUserPenaltyApplier{}
	eval := application.NewPenaltyEvaluator(repo, scoreWriter, penaltyRepo, applier, 0.5, 3, -25, 1, 7*24*time.Hour)

	if err := eval.EvaluateAfterRejection(context.Background(), "u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(penaltyRepo.penalties) != 1 {
		t.Fatalf("expected 1 penalty to be applied, got %d", len(penaltyRepo.penalties))
	}
	if penaltyRepo.penalties[0].Type != string(domain.PenaltyTypeDailyLimitReduction) || penaltyRepo.penalties[0].Value != 1 {
		t.Errorf("unexpected penalty: %+v", penaltyRepo.penalties[0])
	}
	if applier.calls != 1 || applier.limit != 1 {
		t.Errorf("expected the daily limit override to be applied once with value 1, got calls=%d limit=%d", applier.calls, applier.limit)
	}
	if scoreWriter.scores["u1"] != -25 {
		t.Errorf("expected a -25 reputation penalty delta, got %d", scoreWriter.scores["u1"])
	}
}

func TestPenaltyEvaluator_DoesNotStackWhilePenaltyIsActive(t *testing.T) {
	repo := &fakeReputationRepo{}
	seedJudgedEvents(repo, "u1", 0, 5) // 100% rejected, well above threshold
	penaltyRepo := &fakePenaltyRepo{}
	applier := &fakeUserPenaltyApplier{}
	eval := application.NewPenaltyEvaluator(repo, newFakeScoreWriter(), penaltyRepo, applier, 0.5, 3, -25, 1, 7*24*time.Hour)

	if err := eval.EvaluateAfterRejection(context.Background(), "u1"); err != nil {
		t.Fatalf("unexpected error on first evaluation: %v", err)
	}
	if err := eval.EvaluateAfterRejection(context.Background(), "u1"); err != nil {
		t.Fatalf("unexpected error on second evaluation: %v", err)
	}

	if len(penaltyRepo.penalties) != 1 {
		t.Errorf("expected exactly 1 penalty despite 2 evaluations while it's active, got %d", len(penaltyRepo.penalties))
	}
	if applier.calls != 1 {
		t.Errorf("expected exactly 1 daily limit override call, got %d", applier.calls)
	}
}

func TestPenaltyEvaluator_ReappliesAfterPreviousPenaltyExpired(t *testing.T) {
	repo := &fakeReputationRepo{}
	seedJudgedEvents(repo, "u1", 0, 5)
	past := time.Now().UTC().Add(-time.Hour)
	penaltyRepo := &fakePenaltyRepo{penalties: []*domain.Penalty{
		{UserID: "u1", Type: string(domain.PenaltyTypeDailyLimitReduction), AppliedAt: past.Add(-time.Hour), ExpiresAt: &past},
	}}
	applier := &fakeUserPenaltyApplier{}
	eval := application.NewPenaltyEvaluator(repo, newFakeScoreWriter(), penaltyRepo, applier, 0.5, 3, -25, 1, 7*24*time.Hour)

	if err := eval.EvaluateAfterRejection(context.Background(), "u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(penaltyRepo.penalties) != 2 {
		t.Errorf("expected a new penalty once the previous one expired, got %d total", len(penaltyRepo.penalties))
	}
}
