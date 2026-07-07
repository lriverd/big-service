package domain

import "context"

// ReportRepository owns the atomic report → count → maybe-auto-hide sequence.
// Creating a report and (conditionally) transitioning the spot to HIDDEN must
// happen in a single transaction so concurrent reports on the same spot can
// never double-hide it or apply its side effects twice.
type ReportRepository interface {
	// RunReportTransaction creates the given report (rejecting a duplicate
	// report from the same user, via ErrAlreadyReported) and, based on the
	// decide callback, may transition the spot to HIDDEN in the same
	// transaction. decide receives the new report count (including this
	// report) and the spot's current raw status string, and returns whether
	// the spot should be auto-hidden.
	RunReportTransaction(ctx context.Context, report *SpotReport, decide func(newReportCount int, currentSpotStatus string) bool) (created *SpotReport, autoHidden bool, err error)
	ListBySpot(ctx context.Context, spotID string, limit, offset int) ([]*SpotReport, int, error)
}

// SpotCacheInvalidator is a small, consumer-defined port that lets the
// moderation module tell the spots module to drop its cached copy of a spot
// after an auto-hide, instead of waiting for the cache to expire.
type SpotCacheInvalidator interface {
	InvalidateSpotCache(id string)
}

// SpotOwnerLookup is a small, consumer-defined port that lets the
// moderation module find who owns a spot, so an auto-hide can be recorded
// as a reputation event against the right user.
type SpotOwnerLookup interface {
	GetSpotOwnerID(ctx context.Context, spotID string) (string, error)
}

// ReputationRecorder is a small, consumer-defined port that lets the
// moderation module log a reputation event without depending on the
// reputation module's domain types.
type ReputationRecorder interface {
	RecordReputationEvent(ctx context.Context, userID, eventType string, delta int, relatedSpotID, reason string) error
}
