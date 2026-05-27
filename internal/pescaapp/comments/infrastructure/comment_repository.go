package infrastructure

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/lriverd/big-service/internal/pescaapp/comments/domain"
)

type CommentFirestoreRepository struct {
	client     *firestore.Client
	collection string
}

func NewCommentRepository(client *firestore.Client) *CommentFirestoreRepository {
	return &CommentFirestoreRepository{client: client, collection: "comments"}
}

func (r *CommentFirestoreRepository) FindByID(ctx context.Context, id string) (*domain.Comment, error) {
	doc, err := r.client.Collection(r.collection).Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	var c domain.Comment
	if err := doc.DataTo(&c); err != nil {
		return nil, err
	}
	c.ID = doc.Ref.ID
	return &c, nil
}

func (r *CommentFirestoreRepository) ListBySpot(ctx context.Context, spotID string, limit, offset int, sortBy string) ([]*domain.Comment, int, error) {
	// Query only by spotId + order — no deletedAt filter to avoid needing a composite index.
	// Soft-deleted comments are filtered in memory.
	query := r.client.Collection(r.collection).
		Where("spotId", "==", spotID)

	switch sortBy {
	case "oldest":
		query = query.OrderBy("createdAt", firestore.Asc)
	case "helpful":
		query = query.OrderBy("likes", firestore.Desc)
	default: // recent
		query = query.OrderBy("createdAt", firestore.Desc)
	}

	all, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, 0, err
	}

	// Decode and filter deleted in one pass
	var active []*domain.Comment
	for _, doc := range all {
		var c domain.Comment
		if err := doc.DataTo(&c); err != nil {
			continue
		}
		if c.DeletedAt != nil {
			continue
		}
		c.ID = doc.Ref.ID
		active = append(active, &c)
	}
	total := len(active)

	end := offset + limit
	if end > total {
		end = total
	}
	var page []*domain.Comment
	if offset < total {
		page = active[offset:end]
	}
	return page, total, nil
}

func (r *CommentFirestoreRepository) Create(ctx context.Context, comment *domain.Comment) (*domain.Comment, error) {
	ref, _, err := r.client.Collection(r.collection).Add(ctx, map[string]interface{}{
		"spotId":    comment.SpotID,
		"userId":    comment.UserID,
		"text":      comment.Text,
		"likes":     0,
		"createdAt": comment.CreatedAt,
		"updatedAt": nil,
		"deletedAt": nil,
	})
	if err != nil {
		return nil, err
	}
	comment.ID = ref.ID
	return comment, nil
}

func (r *CommentFirestoreRepository) Update(ctx context.Context, id string, text string) (*domain.Comment, error) {
	now := time.Now().UTC()
	_, err := r.client.Collection(r.collection).Doc(id).Update(ctx, []firestore.Update{
		{Path: "text", Value: text},
		{Path: "updatedAt", Value: now},
	})
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *CommentFirestoreRepository) SoftDelete(ctx context.Context, id string) error {
	now := time.Now().UTC()
	_, err := r.client.Collection(r.collection).Doc(id).Update(ctx, []firestore.Update{
		{Path: "deletedAt", Value: now},
	})
	return err
}

func (r *CommentFirestoreRepository) IncrementLikes(ctx context.Context, id string, delta int) error {
	_, err := r.client.Collection(r.collection).Doc(id).Update(ctx, []firestore.Update{
		{Path: "likes", Value: firestore.Increment(delta)},
	})
	return err
}

