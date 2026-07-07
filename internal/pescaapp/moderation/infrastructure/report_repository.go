package infrastructure

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/lriverd/big-service/internal/pescaapp/moderation/domain"
)

// ReportFirestoreRepository owns the atomic report-creation / auto-hide
// transaction. It necessarily knows about both the "spot_reports" and
// "fishing_spots" collections: Firestore transactions can only span
// collections reachable from a single client, and the domain ports for
// those two collections don't (and shouldn't) expose a shared transaction
// handle across module boundaries. That coupling is deliberately contained
// here in infrastructure — nothing outside this file knows the spots
// collection's name or schema.
type ReportFirestoreRepository struct {
	client            *firestore.Client
	reportsCollection string
	spotsCollection   string
}

func NewReportRepository(client *firestore.Client) *ReportFirestoreRepository {
	return &ReportFirestoreRepository{
		client:            client,
		reportsCollection: "spot_reports",
		spotsCollection:   "fishing_spots",
	}
}

func (r *ReportFirestoreRepository) RunReportTransaction(
	ctx context.Context,
	report *domain.SpotReport,
	decide func(newReportCount int, currentSpotStatus string) bool,
) (*domain.SpotReport, bool, error) {
	var created *domain.SpotReport
	var autoHidden bool

	err := r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		created = nil
		autoHidden = false

		existingQuery := r.client.Collection(r.reportsCollection).
			Where("spotId", "==", report.SpotID).
			Where("reporterUserId", "==", report.ReporterUserID).
			Limit(1)
		existing, err := tx.Documents(existingQuery).GetAll()
		if err != nil {
			return err
		}
		if len(existing) > 0 {
			return domain.ErrAlreadyReported
		}

		spotRef := r.client.Collection(r.spotsCollection).Doc(report.SpotID)
		spotSnap, err := tx.Get(spotRef)
		if err != nil {
			return err
		}
		var spotData struct {
			Status      string `firestore:"status"`
			ReportCount int    `firestore:"reportCount"`
		}
		if err := spotSnap.DataTo(&spotData); err != nil {
			return err
		}

		newReportCount := spotData.ReportCount + 1
		shouldHide := decide(newReportCount, spotData.Status)

		now := time.Now().UTC()
		reportRef := r.client.Collection(r.reportsCollection).NewDoc()
		if err := tx.Create(reportRef, map[string]interface{}{
			"spotId":         report.SpotID,
			"reporterUserId": report.ReporterUserID,
			"reason":         string(report.Reason),
			"details":        report.Details,
			"status":         string(domain.ReportStatusPendingReview),
			"createdAt":      now,
		}); err != nil {
			return err
		}

		spotUpdates := []firestore.Update{
			{Path: "reportCount", Value: firestore.Increment(1)},
		}
		if shouldHide {
			spotUpdates = append(spotUpdates,
				firestore.Update{Path: "status", Value: "HIDDEN"},
				firestore.Update{Path: "updatedAt", Value: now},
			)
		}
		if err := tx.Update(spotRef, spotUpdates); err != nil {
			return err
		}

		created = &domain.SpotReport{
			ID:             reportRef.ID,
			SpotID:         report.SpotID,
			ReporterUserID: report.ReporterUserID,
			Reason:         report.Reason,
			Details:        report.Details,
			Status:         domain.ReportStatusPendingReview,
			CreatedAt:      now,
		}
		autoHidden = shouldHide
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	return created, autoHidden, nil
}

func (r *ReportFirestoreRepository) ListBySpot(ctx context.Context, spotID string, limit, offset int) ([]*domain.SpotReport, int, error) {
	query := r.client.Collection(r.reportsCollection).
		Where("spotId", "==", spotID).
		OrderBy("createdAt", firestore.Desc)

	all, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, 0, err
	}

	var reports []*domain.SpotReport
	for _, doc := range all {
		var rep domain.SpotReport
		if err := doc.DataTo(&rep); err != nil {
			continue
		}
		rep.ID = doc.Ref.ID
		reports = append(reports, &rep)
	}

	total := len(reports)
	end := offset + limit
	if end > len(reports) {
		end = len(reports)
	}
	if offset >= len(reports) {
		return []*domain.SpotReport{}, total, nil
	}
	return reports[offset:end], total, nil
}
