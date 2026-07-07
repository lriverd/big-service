package domain_test

import (
	"testing"

	"github.com/lriverd/big-service/internal/pescaapp/reputation/domain"
)

func TestReputationEvent(t *testing.T) {
	spotID := "s1"
	event := domain.ReputationEvent{
		ID: "e1", UserID: "u1", EventType: string(domain.EventSpotVerified),
		Delta: 10, RelatedSpotID: &spotID, Reason: "verified",
	}
	if event.EventType != "SPOT_VERIFIED" || event.Delta != 10 || *event.RelatedSpotID != "s1" {
		t.Error("unexpected reputation event fields")
	}
}
