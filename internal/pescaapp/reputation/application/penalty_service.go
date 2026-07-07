package application

import (
	"context"
	"fmt"
	"time"

	"github.com/lriverd/big-service/internal/pescaapp/reputation/domain"
	log "github.com/sirupsen/logrus"
)

// PenaltyEvaluator watches a user's rejected-content rate and, once it
// crosses a configured threshold (with a minimum sample size, so a single
// rejected spot doesn't punish a brand-new user), applies a temporary
// penalty: a reputation score decrease plus a temporary reduction of their
// daily spot-creation limit. It intentionally does not depend on
// ReputationService — it shares the same two low-level ports
// (ReputationRepository, UserScoreWriter) instead — so ReputationService can
// call it after recording a rejection event without a circular dependency.
type PenaltyEvaluator struct {
	reputationRepo     domain.ReputationRepository
	scoreWriter        domain.UserScoreWriter
	penaltyRepo        domain.PenaltyRepository
	userPenaltyApplier domain.UserPenaltyApplier

	rateThreshold          float64
	minSampleSize          int
	reputationDeltaPenalty int
	dailyLimitReducedValue int
	dailyLimitDuration     time.Duration
}

func NewPenaltyEvaluator(
	reputationRepo domain.ReputationRepository,
	scoreWriter domain.UserScoreWriter,
	penaltyRepo domain.PenaltyRepository,
	userPenaltyApplier domain.UserPenaltyApplier,
	rateThreshold float64,
	minSampleSize int,
	reputationDeltaPenalty int,
	dailyLimitReducedValue int,
	dailyLimitDuration time.Duration,
) *PenaltyEvaluator {
	return &PenaltyEvaluator{
		reputationRepo:         reputationRepo,
		scoreWriter:            scoreWriter,
		penaltyRepo:            penaltyRepo,
		userPenaltyApplier:     userPenaltyApplier,
		rateThreshold:          rateThreshold,
		minSampleSize:          minSampleSize,
		reputationDeltaPenalty: reputationDeltaPenalty,
		dailyLimitReducedValue: dailyLimitReducedValue,
		dailyLimitDuration:     dailyLimitDuration,
	}
}

// EvaluateAfterRejection recomputes userID's all-time rejected-content rate
// (rejected spots ÷ judged spots, from their reputation history) and, if it
// meets or exceeds the threshold and no equivalent penalty is already
// active, applies a new one. Applying it is idempotent per active window:
// a user already under an active DAILY_LIMIT_REDUCTION penalty is not
// penalized again until that one expires or is revoked.
func (e *PenaltyEvaluator) EvaluateAfterRejection(ctx context.Context, userID string) error {
	// A user's judged-content volume is inherently small (bounded by how many
	// spots one person creates), so loading their full history is cheap and
	// avoids the complexity of a separate rolling-window aggregate.
	events, _, err := e.reputationRepo.ListByUser(ctx, userID, 1000, 0)
	if err != nil {
		return fmt.Errorf("loading reputation history for penalty evaluation: %w", err)
	}

	judged, rejected := 0, 0
	for _, ev := range events {
		switch domain.ReputationEventType(ev.EventType) {
		case domain.EventSpotVerified:
			judged++
		case domain.EventSpotHidden, domain.EventSpotDeleted:
			judged++
			rejected++
		}
	}
	if judged < e.minSampleSize {
		return nil
	}
	rate := float64(rejected) / float64(judged)
	if rate < e.rateThreshold {
		return nil
	}

	now := time.Now().UTC()
	active, err := e.penaltyRepo.HasActivePenaltyOfType(ctx, userID, string(domain.PenaltyTypeDailyLimitReduction), now)
	if err != nil {
		return fmt.Errorf("checking for an active penalty: %w", err)
	}
	if active {
		return nil
	}

	expiresAt := now.Add(e.dailyLimitDuration)
	reason := fmt.Sprintf("Tasa de contenido rechazado del %.0f%% supera el umbral configurado", rate*100)
	penalty := &domain.Penalty{
		UserID:    userID,
		Type:      string(domain.PenaltyTypeDailyLimitReduction),
		Value:     e.dailyLimitReducedValue,
		Reason:    reason,
		AppliedAt: now,
		ExpiresAt: &expiresAt,
	}
	if _, err := e.penaltyRepo.Create(ctx, penalty); err != nil {
		return fmt.Errorf("creating penalty: %w", err)
	}

	if err := e.userPenaltyApplier.SetDailySpotLimitOverride(ctx, userID, e.dailyLimitReducedValue, expiresAt); err != nil {
		log.WithError(err).WithField("userId", userID).Warn("Failed to apply daily spot limit override for penalty")
	}

	// Record the penalty itself as a further reputation event/audit entry,
	// same best-effort convention as ReputationService.RecordEvent.
	event := &domain.ReputationEvent{
		UserID: userID, EventType: string(domain.EventRejectedContentPenalty),
		Delta: e.reputationDeltaPenalty, Reason: reason, CreatedAt: now,
	}
	if _, err := e.reputationRepo.Create(ctx, event); err != nil {
		log.WithError(err).WithField("userId", userID).Warn("Failed to record penalty reputation event")
	} else if err := e.scoreWriter.IncrementReputationScore(ctx, userID, e.reputationDeltaPenalty); err != nil {
		log.WithError(err).WithField("userId", userID).Warn("Failed to apply penalty reputation score delta")
	}

	log.WithFields(log.Fields{"userId": userID, "rate": rate}).Info("Applied rejected-content penalty")
	return nil
}

func (e *PenaltyEvaluator) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Penalty, int, error) {
	return e.penaltyRepo.ListByUser(ctx, userID, limit, offset)
}
