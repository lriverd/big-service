package infrastructure

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/lriverd/big-service/internal/pescaapp/reputation/domain"
)

type ReputationFirestoreRepository struct {
	client     *firestore.Client
	collection string
}

func NewReputationRepository(client *firestore.Client) *ReputationFirestoreRepository {
	return &ReputationFirestoreRepository{client: client, collection: "reputation_events"}
}

func (r *ReputationFirestoreRepository) Create(ctx context.Context, event *domain.ReputationEvent) (*domain.ReputationEvent, error) {
	data := map[string]interface{}{
		"userId":    event.UserID,
		"eventType": event.EventType,
		"delta":     event.Delta,
		"reason":    event.Reason,
		"createdAt": event.CreatedAt,
	}
	if event.RelatedSpotID != nil {
		data["relatedSpotId"] = *event.RelatedSpotID
	}
	if event.RelatedReportID != nil {
		data["relatedReportId"] = *event.RelatedReportID
	}

	ref, _, err := r.client.Collection(r.collection).Add(ctx, data)
	if err != nil {
		return nil, err
	}
	event.ID = ref.ID
	return event, nil
}

func (r *ReputationFirestoreRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.ReputationEvent, int, error) {
	query := r.client.Collection(r.collection).
		Where("userId", "==", userID).
		OrderBy("createdAt", firestore.Desc)

	all, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, 0, err
	}

	var events []*domain.ReputationEvent
	for _, doc := range all {
		var e domain.ReputationEvent
		if err := doc.DataTo(&e); err != nil {
			continue
		}
		e.ID = doc.Ref.ID
		events = append(events, &e)
	}

	total := len(events)
	end := offset + limit
	if end > len(events) {
		end = len(events)
	}
	if offset >= len(events) {
		return []*domain.ReputationEvent{}, total, nil
	}
	return events[offset:end], total, nil
}
