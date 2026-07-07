package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/lriverd/big-service/internal/pescaapp/moderation/domain"
	spotsDomain "github.com/lriverd/big-service/internal/pescaapp/spots/domain"
	apperrors "github.com/lriverd/big-service/internal/shared/errors"
	log "github.com/sirupsen/logrus"
)

type ReportService struct {
	repo                  domain.ReportRepository
	invalidator           domain.SpotCacheInvalidator
	ownerLookup           domain.SpotOwnerLookup
	reputationRecorder    domain.ReputationRecorder
	hideThreshold         int
	reputationDeltaHidden int
}

func NewReportService(
	repo domain.ReportRepository,
	invalidator domain.SpotCacheInvalidator,
	ownerLookup domain.SpotOwnerLookup,
	reputationRecorder domain.ReputationRecorder,
	hideThreshold int,
	reputationDeltaHidden int,
) *ReportService {
	return &ReportService{
		repo:                  repo,
		invalidator:           invalidator,
		ownerLookup:           ownerLookup,
		reputationRecorder:    reputationRecorder,
		hideThreshold:         hideThreshold,
		reputationDeltaHidden: reputationDeltaHidden,
	}
}

// Report records a user's report against a spot. If this report pushes the
// spot's valid report count to the configured threshold, the spot is
// atomically transitioned to HIDDEN as part of the same operation.
func (s *ReportService) Report(ctx context.Context, spotID, reporterUserID string, reason domain.ReportReason, details string) (*domain.SpotReport, bool, error) {
	if !reason.IsValid() {
		return nil, false, apperrors.BadRequest("Invalid report reason")
	}
	if reason == domain.ReportReasonOther && strings.TrimSpace(details) == "" {
		return nil, false, apperrors.BadRequest("Details are required when reason is 'otro'")
	}

	report := &domain.SpotReport{
		SpotID:         spotID,
		ReporterUserID: reporterUserID,
		Reason:         reason,
		Details:        details,
	}

	var finalReportCount int
	created, autoHidden, err := s.repo.RunReportTransaction(ctx, report, func(newReportCount int, currentSpotStatus string) bool {
		finalReportCount = newReportCount
		switch spotsDomain.SpotStatus(currentSpotStatus) {
		case spotsDomain.SpotStatusHidden, spotsDomain.SpotStatusDeleted:
			return false // already out of public view; don't re-trigger
		}
		return newReportCount >= s.hideThreshold
	})
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyReported) {
			return nil, false, apperrors.Conflict("You have already reported this spot")
		}
		return nil, false, fmt.Errorf("creating spot report: %w", err)
	}

	if autoHidden {
		s.invalidator.InvalidateSpotCache(spotID)
		if ownerID, err := s.ownerLookup.GetSpotOwnerID(ctx, spotID); err != nil {
			log.WithError(err).WithField("spotId", spotID).Warn("Failed to look up spot owner for reputation event after auto-hide")
		} else {
			reason := fmt.Sprintf("Spot ocultado automáticamente tras alcanzar %d reportes", finalReportCount)
			if err := s.reputationRecorder.RecordReputationEvent(ctx, ownerID, "SPOT_HIDDEN", s.reputationDeltaHidden, spotID, reason); err != nil {
				log.WithError(err).WithField("spotId", spotID).Warn("Failed to record reputation event after auto-hide")
			}
		}
		log.WithField("spotId", spotID).Info("Spot auto-hidden after reaching the report threshold")
	}

	return created, autoHidden, nil
}

func (s *ReportService) ListBySpot(ctx context.Context, spotID string, limit, offset int) ([]*domain.SpotReport, int, error) {
	return s.repo.ListBySpot(ctx, spotID, limit, offset)
}
