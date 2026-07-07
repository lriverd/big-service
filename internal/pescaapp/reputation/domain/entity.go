package domain

import "time"

// ReputationEventType is the set of things that move a user's reputation
// score. It also doubles as the audit trail's event vocabulary — every
// score change must be recorded as one of these, never applied silently.
type ReputationEventType string

const (
	EventSpotVerified           ReputationEventType = "SPOT_VERIFIED"
	EventSpotHidden             ReputationEventType = "SPOT_HIDDEN"
	EventSpotDeleted            ReputationEventType = "SPOT_DELETED"
	EventGoodRatingReceived     ReputationEventType = "GOOD_RATING_RECEIVED"
	EventRejectedContentPenalty ReputationEventType = "REJECTED_CONTENT_PENALTY"
)

// ReputationEvent is one entry in a user's reputation history. The full set
// of events for a user is also its audit trail — there is no separate audit
// log, this collection already is one.
type ReputationEvent struct {
	ID              string    `json:"id" firestore:"-"`
	UserID          string    `json:"userId" firestore:"userId"`
	EventType       string    `json:"eventType" firestore:"eventType"`
	Delta           int       `json:"delta" firestore:"delta"`
	RelatedSpotID   *string   `json:"relatedSpotId,omitempty" firestore:"relatedSpotId,omitempty"`
	RelatedReportID *string   `json:"relatedReportId,omitempty" firestore:"relatedReportId,omitempty"`
	Reason          string    `json:"reason" firestore:"reason"`
	CreatedAt       time.Time `json:"createdAt" firestore:"createdAt"`
}

// PenaltyType is the set of penalties that can be applied to a user.
// Deliberately a plain string, not a closed set enforced elsewhere: adding a
// new penalty type is meant to be just a new constant plus its
// interpretation in PenaltyEvaluator, not a schema or API change.
type PenaltyType string

const (
	PenaltyTypeDailyLimitReduction PenaltyType = "DAILY_LIMIT_REDUCTION"
)

// Penalty is a temporary, time-boxed restriction applied to a user as a
// consequence of a high rate of rejected content. "Active" is computed from
// AppliedAt/ExpiresAt/RevokedAt at read time (see IsActive) rather than
// stored as a mutable flag, so nothing needs a background job to turn a
// penalty off when it expires.
type Penalty struct {
	ID              string     `json:"id" firestore:"-"`
	UserID          string     `json:"userId" firestore:"userId"`
	Type            string     `json:"type" firestore:"type"`
	Value           int        `json:"value" firestore:"value"`
	Reason          string     `json:"reason" firestore:"reason"`
	AppliedAt       time.Time  `json:"appliedAt" firestore:"appliedAt"`
	ExpiresAt       *time.Time `json:"expiresAt,omitempty" firestore:"expiresAt,omitempty"`
	RevokedAt       *time.Time `json:"revokedAt,omitempty" firestore:"revokedAt,omitempty"`
	RevokedByUserID *string    `json:"revokedByUserId,omitempty" firestore:"revokedByUserId,omitempty"`
}

// IsActive reports whether the penalty is currently in effect: not revoked,
// and either permanent (no ExpiresAt) or not yet expired.
func (p *Penalty) IsActive(now time.Time) bool {
	if p.RevokedAt != nil {
		return false
	}
	if p.ExpiresAt != nil && now.After(*p.ExpiresAt) {
		return false
	}
	return true
}
