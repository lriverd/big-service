package application

import (
	"context"
	"fmt"
	"time"

	"github.com/lriverd/big-service/internal/pescaapp/reputation/domain"
	log "github.com/sirupsen/logrus"
)

type ReputationService struct {
	repo           domain.ReputationRepository
	scoreWriter    domain.UserScoreWriter
	scoreReader    domain.UserScoreReader
	penaltyTrigger domain.PenaltyTrigger // optional; nil disables penalty evaluation (e.g. in tests)
}

func NewReputationService(repo domain.ReputationRepository, scoreWriter domain.UserScoreWriter, scoreReader domain.UserScoreReader, penaltyTrigger domain.PenaltyTrigger) *ReputationService {
	return &ReputationService{repo: repo, scoreWriter: scoreWriter, scoreReader: scoreReader, penaltyTrigger: penaltyTrigger}
}

// RecordEvent appends an entry to the user's reputation history and applies
// its delta to their score. The event is always recorded first: if the
// score update then fails, the audit trail still reflects what should have
// happened, and the update is logged rather than propagated as a hard
// failure — consistent with how other cross-aggregate side effects in this
// codebase (e.g. spot comment/rating counters) are treated as best-effort.
func (s *ReputationService) RecordEvent(ctx context.Context, userID string, eventType domain.ReputationEventType, delta int, relatedSpotID, relatedReportID *string, reason string) (*domain.ReputationEvent, error) {
	event := &domain.ReputationEvent{
		UserID:          userID,
		EventType:       string(eventType),
		Delta:           delta,
		RelatedSpotID:   relatedSpotID,
		RelatedReportID: relatedReportID,
		Reason:          reason,
		CreatedAt:       time.Now().UTC(),
	}

	created, err := s.repo.Create(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("recording reputation event: %w", err)
	}

	if err := s.scoreWriter.IncrementReputationScore(ctx, userID, delta); err != nil {
		log.WithError(err).WithField("userId", userID).Warn("Failed to apply reputation score delta after recording event")
	}

	if s.penaltyTrigger != nil && isRejectionEvent(eventType) {
		if err := s.penaltyTrigger.EvaluateAfterRejection(ctx, userID); err != nil {
			log.WithError(err).WithField("userId", userID).Warn("Failed to evaluate rejected-content penalty")
		}
	}

	return created, nil
}

func isRejectionEvent(eventType domain.ReputationEventType) bool {
	return eventType == domain.EventSpotHidden || eventType == domain.EventSpotDeleted
}

// RecordReputationEvent is a primitive-typed adapter that satisfies the
// small ReputationRecorder ports defined by consumer modules (spots,
// moderation, ratings), so those modules don't need to depend on this
// module's domain types.
func (s *ReputationService) RecordReputationEvent(ctx context.Context, userID, eventType string, delta int, relatedSpotID, reason string) error {
	var relSpot *string
	if relatedSpotID != "" {
		relSpot = &relatedSpotID
	}
	_, err := s.RecordEvent(ctx, userID, domain.ReputationEventType(eventType), delta, relSpot, nil, reason)
	return err
}

func (s *ReputationService) GetSummary(ctx context.Context, userID string, recentEventsLimit int) (int, []*domain.ReputationEvent, error) {
	score, err := s.scoreReader.GetReputationScore(ctx, userID)
	if err != nil {
		return 0, nil, err
	}
	recent, _, err := s.repo.ListByUser(ctx, userID, recentEventsLimit, 0)
	if err != nil {
		return 0, nil, err
	}
	return score, recent, nil
}

func (s *ReputationService) ListHistory(ctx context.Context, userID string, limit, offset int) ([]*domain.ReputationEvent, int, error) {
	return s.repo.ListByUser(ctx, userID, limit, offset)
}
