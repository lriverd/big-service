package domain_test

import (
	"testing"

	"github.com/lriverd/big-service/internal/pescaapp/moderation/domain"
)

func TestReportReason_IsValid(t *testing.T) {
	valid := []domain.ReportReason{
		domain.ReportReasonDoesNotExist,
		domain.ReportReasonWrongLocation,
		domain.ReportReasonFalseInfo,
		domain.ReportReasonDuplicate,
		domain.ReportReasonOther,
	}
	for _, r := range valid {
		if !r.IsValid() {
			t.Errorf("expected %s to be valid", r)
		}
	}
	if domain.ReportReason("not_a_reason").IsValid() {
		t.Error("expected unknown reason to be invalid")
	}
}

func TestSpotReport(t *testing.T) {
	report := domain.SpotReport{
		ID: "r1", SpotID: "s1", ReporterUserID: "u1",
		Reason: domain.ReportReasonDuplicate, Status: domain.ReportStatusPendingReview,
	}
	if report.Reason != domain.ReportReasonDuplicate || report.Status != domain.ReportStatusPendingReview {
		t.Error("unexpected report fields")
	}
}
