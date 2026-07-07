package infrastructure

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/lriverd/big-service/internal/pescaapp/reputation/domain"
)

type PenaltyFirestoreRepository struct {
	client     *firestore.Client
	collection string
}

func NewPenaltyRepository(client *firestore.Client) *PenaltyFirestoreRepository {
	return &PenaltyFirestoreRepository{client: client, collection: "user_penalties"}
}

func (r *PenaltyFirestoreRepository) Create(ctx context.Context, penalty *domain.Penalty) (*domain.Penalty, error) {
	data := map[string]interface{}{
		"userId":    penalty.UserID,
		"type":      penalty.Type,
		"value":     penalty.Value,
		"reason":    penalty.Reason,
		"appliedAt": penalty.AppliedAt,
	}
	if penalty.ExpiresAt != nil {
		data["expiresAt"] = *penalty.ExpiresAt
	}

	ref, _, err := r.client.Collection(r.collection).Add(ctx, data)
	if err != nil {
		return nil, err
	}
	penalty.ID = ref.ID
	return penalty, nil
}

func (r *PenaltyFirestoreRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Penalty, int, error) {
	query := r.client.Collection(r.collection).
		Where("userId", "==", userID).
		OrderBy("appliedAt", firestore.Desc)

	all, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, 0, err
	}

	var penalties []*domain.Penalty
	for _, doc := range all {
		var p domain.Penalty
		if err := doc.DataTo(&p); err != nil {
			continue
		}
		p.ID = doc.Ref.ID
		penalties = append(penalties, &p)
	}

	total := len(penalties)
	end := offset + limit
	if end > len(penalties) {
		end = len(penalties)
	}
	if offset >= len(penalties) {
		return []*domain.Penalty{}, total, nil
	}
	return penalties[offset:end], total, nil
}

func (r *PenaltyFirestoreRepository) HasActivePenaltyOfType(ctx context.Context, userID, penaltyType string, now time.Time) (bool, error) {
	docs, err := r.client.Collection(r.collection).
		Where("userId", "==", userID).
		Where("type", "==", penaltyType).
		Documents(ctx).GetAll()
	if err != nil {
		return false, err
	}
	for _, doc := range docs {
		var p domain.Penalty
		if err := doc.DataTo(&p); err != nil {
			continue
		}
		if p.IsActive(now) {
			return true, nil
		}
	}
	return false, nil
}
