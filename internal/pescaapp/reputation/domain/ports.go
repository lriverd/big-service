package domain

import (
	"context"
	"time"
)

type ReputationRepository interface {
	Create(ctx context.Context, event *ReputationEvent) (*ReputationEvent, error)
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]*ReputationEvent, int, error)
}

type PenaltyRepository interface {
	Create(ctx context.Context, penalty *Penalty) (*Penalty, error)
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]*Penalty, int, error)
	HasActivePenaltyOfType(ctx context.Context, userID, penaltyType string, now time.Time) (bool, error)
}

// UserScoreWriter is a small, consumer-defined port that lets the
// reputation module apply a score delta to a user without depending on the
// full users module.
type UserScoreWriter interface {
	IncrementReputationScore(ctx context.Context, userID string, delta int) error
}

// UserScoreReader is a small, consumer-defined port that lets the
// reputation module read a user's current score for the summary endpoint.
type UserScoreReader interface {
	GetReputationScore(ctx context.Context, userID string) (int, error)
}

// UserPenaltyApplier is a small, consumer-defined port that lets a penalty
// apply a temporary daily-spot-limit reduction to a user, without the
// reputation module depending on the full users module.
type UserPenaltyApplier interface {
	SetDailySpotLimitOverride(ctx context.Context, userID string, limit int, expiresAt time.Time) error
}

// PenaltyTrigger is implemented by PenaltyEvaluator and called by
// ReputationService right after recording a rejection event (SPOT_HIDDEN or
// SPOT_DELETED), so penalty evaluation stays a self-contained concern that
// consumer modules (spots, moderation) never need to know about.
type PenaltyTrigger interface {
	EvaluateAfterRejection(ctx context.Context, userID string) error
}
