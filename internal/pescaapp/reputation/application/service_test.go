package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/lriverd/big-service/internal/pescaapp/reputation/application"
	"github.com/lriverd/big-service/internal/pescaapp/reputation/domain"
)

type fakeReputationRepo struct {
	events []*domain.ReputationEvent
}

func (f *fakeReputationRepo) Create(ctx context.Context, event *domain.ReputationEvent) (*domain.ReputationEvent, error) {
	event.ID = "event-1"
	f.events = append(f.events, event)
	return event, nil
}

func (f *fakeReputationRepo) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.ReputationEvent, int, error) {
	var matched []*domain.ReputationEvent
	for _, e := range f.events {
		if e.UserID == userID {
			matched = append(matched, e)
		}
	}
	total := len(matched)
	end := offset + limit
	if end > len(matched) {
		end = len(matched)
	}
	if offset >= len(matched) {
		return []*domain.ReputationEvent{}, total, nil
	}
	return matched[offset:end], total, nil
}

type fakeScoreWriter struct {
	scores  map[string]int
	failErr error
}

func newFakeScoreWriter() *fakeScoreWriter {
	return &fakeScoreWriter{scores: map[string]int{}}
}

func (f *fakeScoreWriter) IncrementReputationScore(ctx context.Context, userID string, delta int) error {
	if f.failErr != nil {
		return f.failErr
	}
	f.scores[userID] += delta
	return nil
}

type fakeScoreReader struct {
	scores map[string]int
}

func (f *fakeScoreReader) GetReputationScore(ctx context.Context, userID string) (int, error) {
	return f.scores[userID], nil
}

func TestReputationService_RecordEvent_CreatesEventAndAppliesDelta(t *testing.T) {
	repo := &fakeReputationRepo{}
	writer := newFakeScoreWriter()
	svc := application.NewReputationService(repo, writer, &fakeScoreReader{}, nil)

	spotID := "spot-1"
	event, err := svc.RecordEvent(context.Background(), "u1", domain.EventSpotVerified, 10, &spotID, nil, "verified")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.EventType != "SPOT_VERIFIED" || event.Delta != 10 {
		t.Errorf("unexpected event: %+v", event)
	}
	if writer.scores["u1"] != 10 {
		t.Errorf("expected score delta of 10 to be applied, got %d", writer.scores["u1"])
	}
}

func TestReputationService_RecordEvent_StillReturnsEventWhenScoreUpdateFails(t *testing.T) {
	repo := &fakeReputationRepo{}
	writer := newFakeScoreWriter()
	writer.failErr = errors.New("firestore unavailable")
	svc := application.NewReputationService(repo, writer, &fakeScoreReader{}, nil)

	event, err := svc.RecordEvent(context.Background(), "u1", domain.EventSpotHidden, -15, nil, nil, "hidden")
	if err != nil {
		t.Fatalf("expected the audit event to still be recorded despite the score-update failure, got error: %v", err)
	}
	if event == nil {
		t.Fatal("expected a non-nil event")
	}
	if len(repo.events) != 1 {
		t.Errorf("expected the event to be persisted even though the score update failed, got %d events", len(repo.events))
	}
}

func TestReputationService_RecordReputationEvent_AdapterConvertsEmptySpotIDToNil(t *testing.T) {
	repo := &fakeReputationRepo{}
	svc := application.NewReputationService(repo, newFakeScoreWriter(), &fakeScoreReader{}, nil)

	if err := svc.RecordReputationEvent(context.Background(), "u1", "SPOT_DELETED", -20, "", "no spot"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.events[0].RelatedSpotID != nil {
		t.Error("expected an empty relatedSpotID string to be stored as nil")
	}
}

func TestReputationService_GetSummary_ReturnsScoreAndRecentEvents(t *testing.T) {
	repo := &fakeReputationRepo{}
	reader := &fakeScoreReader{scores: map[string]int{"u1": 42}}
	svc := application.NewReputationService(repo, newFakeScoreWriter(), reader, nil)

	_, _ = svc.RecordEvent(context.Background(), "u1", domain.EventGoodRatingReceived, 2, nil, nil, "good rating")

	score, events, err := svc.GetSummary(context.Background(), "u1", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score != 42 {
		t.Errorf("expected score 42, got %d", score)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 recent event, got %d", len(events))
	}
}

type fakePenaltyTrigger struct {
	evaluatedUserIDs []string
}

func (f *fakePenaltyTrigger) EvaluateAfterRejection(ctx context.Context, userID string) error {
	f.evaluatedUserIDs = append(f.evaluatedUserIDs, userID)
	return nil
}

func TestReputationService_RecordEvent_TriggersPenaltyEvaluationOnlyForRejections(t *testing.T) {
	trigger := &fakePenaltyTrigger{}
	svc := application.NewReputationService(&fakeReputationRepo{}, newFakeScoreWriter(), &fakeScoreReader{}, trigger)

	_, _ = svc.RecordEvent(context.Background(), "u1", domain.EventSpotVerified, 10, nil, nil, "verified")
	_, _ = svc.RecordEvent(context.Background(), "u1", domain.EventGoodRatingReceived, 2, nil, nil, "good rating")
	if len(trigger.evaluatedUserIDs) != 0 {
		t.Errorf("expected no penalty evaluation for non-rejection events, got %v", trigger.evaluatedUserIDs)
	}

	_, _ = svc.RecordEvent(context.Background(), "u2", domain.EventSpotHidden, -15, nil, nil, "hidden")
	_, _ = svc.RecordEvent(context.Background(), "u3", domain.EventSpotDeleted, -20, nil, nil, "deleted")
	if len(trigger.evaluatedUserIDs) != 2 || trigger.evaluatedUserIDs[0] != "u2" || trigger.evaluatedUserIDs[1] != "u3" {
		t.Errorf("expected penalty evaluation for both SPOT_HIDDEN and SPOT_DELETED, got %v", trigger.evaluatedUserIDs)
	}
}
