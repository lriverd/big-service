package application_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/lriverd/big-service/internal/pescaapp/moderation/application"
	"github.com/lriverd/big-service/internal/pescaapp/moderation/domain"
	apperrors "github.com/lriverd/big-service/internal/shared/errors"
)

// fakeReportRepo mimics the atomicity contract RunReportTransaction relies
// on (read count + status, decide, write) using a mutex instead of a real
// Firestore transaction, so the application-level composition (in
// particular: never auto-hiding twice under concurrent reports) can be
// exercised without a Firestore emulator.
type fakeReportRepo struct {
	mu          sync.Mutex
	status      string
	reportCount int
	reporters   map[string]bool
	reports     []*domain.SpotReport
}

func newFakeReportRepo(initialStatus string) *fakeReportRepo {
	return &fakeReportRepo{status: initialStatus, reporters: map[string]bool{}}
}

func (f *fakeReportRepo) RunReportTransaction(ctx context.Context, report *domain.SpotReport, decide func(int, string) bool) (*domain.SpotReport, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.reporters[report.ReporterUserID] {
		return nil, false, domain.ErrAlreadyReported
	}

	newCount := f.reportCount + 1
	shouldHide := decide(newCount, f.status)

	created := &domain.SpotReport{
		ID:             fmt.Sprintf("report-%d", len(f.reports)+1),
		SpotID:         report.SpotID,
		ReporterUserID: report.ReporterUserID,
		Reason:         report.Reason,
		Details:        report.Details,
		Status:         domain.ReportStatusPendingReview,
		CreatedAt:      time.Now(),
	}

	f.reporters[report.ReporterUserID] = true
	f.reportCount = newCount
	f.reports = append(f.reports, created)
	if shouldHide {
		f.status = "HIDDEN"
	}
	return created, shouldHide, nil
}

func (f *fakeReportRepo) ListBySpot(ctx context.Context, spotID string, limit, offset int) ([]*domain.SpotReport, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.reports, len(f.reports), nil
}

type fakeInvalidator struct {
	mu          sync.Mutex
	invalidated []string
}

func (f *fakeInvalidator) InvalidateSpotCache(id string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.invalidated = append(f.invalidated, id)
}

func (f *fakeInvalidator) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.invalidated)
}

type fakeOwnerLookup struct{}

func (f *fakeOwnerLookup) GetSpotOwnerID(ctx context.Context, spotID string) (string, error) {
	return "owner-of-" + spotID, nil
}

type fakeReputationRecorder struct {
	mu    sync.Mutex
	calls []reputationCall
}

type reputationCall struct {
	userID, eventType string
	delta             int
}

func (f *fakeReputationRecorder) RecordReputationEvent(ctx context.Context, userID, eventType string, delta int, relatedSpotID, reason string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, reputationCall{userID: userID, eventType: eventType, delta: delta})
	return nil
}

func (f *fakeReputationRecorder) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

func TestReportService_Report_CreatesReport(t *testing.T) {
	repo := newFakeReportRepo("VERIFIED")
	svc := application.NewReportService(repo, &fakeInvalidator{}, &fakeOwnerLookup{}, &fakeReputationRecorder{}, 5, -15)

	report, autoHidden, err := svc.Report(context.Background(), "spot-1", "u1", domain.ReportReasonDuplicate, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if autoHidden {
		t.Error("expected a single report to not trigger auto-hide with threshold 5")
	}
	if report.SpotID != "spot-1" || report.ReporterUserID != "u1" {
		t.Errorf("unexpected report: %+v", report)
	}
}

func TestReportService_Report_RejectsDuplicateFromSameUser(t *testing.T) {
	repo := newFakeReportRepo("VERIFIED")
	svc := application.NewReportService(repo, &fakeInvalidator{}, &fakeOwnerLookup{}, &fakeReputationRecorder{}, 5, -15)

	if _, _, err := svc.Report(context.Background(), "spot-1", "u1", domain.ReportReasonDuplicate, ""); err != nil {
		t.Fatalf("unexpected error on first report: %v", err)
	}
	_, _, err := svc.Report(context.Background(), "spot-1", "u1", domain.ReportReasonOther, "otra vez")
	appErr, ok := err.(*apperrors.AppError)
	if !ok || appErr.Status != 409 {
		t.Fatalf("expected a 409 conflict for duplicate report, got %v", err)
	}
}

func TestReportService_Report_RejectsInvalidReason(t *testing.T) {
	repo := newFakeReportRepo("VERIFIED")
	svc := application.NewReportService(repo, &fakeInvalidator{}, &fakeOwnerLookup{}, &fakeReputationRecorder{}, 5, -15)

	_, _, err := svc.Report(context.Background(), "spot-1", "u1", domain.ReportReason("bogus"), "")
	appErr, ok := err.(*apperrors.AppError)
	if !ok || appErr.Status != 400 {
		t.Fatalf("expected a 400 for invalid reason, got %v", err)
	}
}

func TestReportService_Report_RequiresDetailsWhenReasonIsOther(t *testing.T) {
	repo := newFakeReportRepo("VERIFIED")
	svc := application.NewReportService(repo, &fakeInvalidator{}, &fakeOwnerLookup{}, &fakeReputationRecorder{}, 5, -15)

	_, _, err := svc.Report(context.Background(), "spot-1", "u1", domain.ReportReasonOther, "   ")
	appErr, ok := err.(*apperrors.AppError)
	if !ok || appErr.Status != 400 {
		t.Fatalf("expected a 400 when 'otro' has no details, got %v", err)
	}
}

func TestReportService_Report_AutoHidesAtThreshold(t *testing.T) {
	repo := newFakeReportRepo("VERIFIED")
	invalidator := &fakeInvalidator{}
	recorder := &fakeReputationRecorder{}
	svc := application.NewReportService(repo, invalidator, &fakeOwnerLookup{}, recorder, 3, -15)

	for i, user := range []string{"u1", "u2"} {
		_, autoHidden, err := svc.Report(context.Background(), "spot-1", user, domain.ReportReasonDuplicate, "")
		if err != nil {
			t.Fatalf("report %d: unexpected error: %v", i, err)
		}
		if autoHidden {
			t.Fatalf("report %d: did not expect auto-hide before reaching threshold", i)
		}
	}

	_, autoHidden, err := svc.Report(context.Background(), "spot-1", "u3", domain.ReportReasonDuplicate, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !autoHidden {
		t.Error("expected the 3rd report to trigger auto-hide at threshold=3")
	}
	if invalidator.count() != 1 {
		t.Errorf("expected exactly 1 cache invalidation, got %d", invalidator.count())
	}
	if recorder.count() != 1 || recorder.calls[0].eventType != "SPOT_HIDDEN" || recorder.calls[0].delta != -15 {
		t.Errorf("expected exactly 1 SPOT_HIDDEN -15 reputation event, got %+v", recorder.calls)
	}
}

func TestReportService_Report_DoesNotReTriggerOnAlreadyHiddenSpot(t *testing.T) {
	repo := newFakeReportRepo("HIDDEN")
	invalidator := &fakeInvalidator{}
	svc := application.NewReportService(repo, invalidator, &fakeOwnerLookup{}, &fakeReputationRecorder{}, 1, -15)

	_, autoHidden, err := svc.Report(context.Background(), "spot-1", "u1", domain.ReportReasonDuplicate, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if autoHidden {
		t.Error("expected no auto-hide transition for a spot that is already HIDDEN")
	}
	if invalidator.count() != 0 {
		t.Errorf("expected no cache invalidation for an already-hidden spot, got %d", invalidator.count())
	}
}

func TestReportService_Report_ConcurrentReportsHideExactlyOnce(t *testing.T) {
	repo := newFakeReportRepo("VERIFIED")
	invalidator := &fakeInvalidator{}
	recorder := &fakeReputationRecorder{}
	const threshold = 5
	const reporters = 20
	svc := application.NewReportService(repo, invalidator, &fakeOwnerLookup{}, recorder, threshold, -15)

	var wg sync.WaitGroup
	autoHiddenCount := int32(0)
	var mu sync.Mutex
	for i := 0; i < reporters; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_, autoHidden, err := svc.Report(context.Background(), "spot-1", fmt.Sprintf("user-%d", n), domain.ReportReasonDuplicate, "")
			if err != nil {
				t.Errorf("unexpected error from concurrent report: %v", err)
				return
			}
			if autoHidden {
				mu.Lock()
				autoHiddenCount++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()

	if autoHiddenCount != 1 {
		t.Errorf("expected exactly 1 auto-hide transition under concurrent reports, got %d", autoHiddenCount)
	}
	if invalidator.count() != 1 {
		t.Errorf("expected exactly 1 cache invalidation under concurrent reports, got %d", invalidator.count())
	}
	if recorder.count() != 1 {
		t.Errorf("expected exactly 1 reputation event under concurrent reports, got %d", recorder.count())
	}
}
