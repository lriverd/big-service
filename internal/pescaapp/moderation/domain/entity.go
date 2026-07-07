package domain

import (
	"errors"
	"time"
)

// ReportReason is the set of reasons a user can give when reporting a spot.
type ReportReason string

const (
	ReportReasonDoesNotExist  ReportReason = "no_existe"
	ReportReasonWrongLocation ReportReason = "ubicacion_incorrecta"
	ReportReasonFalseInfo     ReportReason = "informacion_falsa"
	ReportReasonDuplicate     ReportReason = "duplicado"
	ReportReasonOther         ReportReason = "otro"
)

func (r ReportReason) IsValid() bool {
	switch r {
	case ReportReasonDoesNotExist, ReportReasonWrongLocation, ReportReasonFalseInfo, ReportReasonDuplicate, ReportReasonOther:
		return true
	default:
		return false
	}
}

// ReportReviewStatus tracks whether a report has been through moderator
// review yet. It starts at PENDING_REVIEW; VALID/REJECTED are set by a
// future manual review action and are not yet reachable via any endpoint.
type ReportReviewStatus string

const (
	ReportStatusPendingReview ReportReviewStatus = "PENDING_REVIEW"
	ReportStatusValid         ReportReviewStatus = "VALID"
	ReportStatusRejected      ReportReviewStatus = "REJECTED"
)

type SpotReport struct {
	ID               string             `json:"id" firestore:"-"`
	SpotID           string             `json:"spotId" firestore:"spotId"`
	ReporterUserID   string             `json:"reporterUserId" firestore:"reporterUserId"`
	Reason           ReportReason       `json:"reason" firestore:"reason"`
	Details          string             `json:"details,omitempty" firestore:"details,omitempty"`
	Status           ReportReviewStatus `json:"status" firestore:"status"`
	CreatedAt        time.Time          `json:"createdAt" firestore:"createdAt"`
	ReviewedAt       *time.Time         `json:"reviewedAt,omitempty" firestore:"reviewedAt,omitempty"`
	ReviewedByUserID *string            `json:"reviewedByUserId,omitempty" firestore:"reviewedByUserId,omitempty"`
}

type CreateReportRequest struct {
	Reason  ReportReason `json:"reason" binding:"required"`
	Details string       `json:"details"`
}

// ErrAlreadyReported is returned when a user tries to report the same spot
// more than once.
var ErrAlreadyReported = errors.New("user has already reported this spot")
