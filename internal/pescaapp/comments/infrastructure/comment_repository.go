package infrastructure

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/lriverd/big-service/internal/pescaapp/comments/domain"
	"google.golang.org/api/iterator"
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
	query := r.client.Collection(r.collection).
		Where("spotId", "==", spotID).
		Where("deletedAt", "==", nil)

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
	total := len(all)

	iter := query.Offset(offset).Limit(limit).Documents(ctx)
	var comments []*domain.Comment
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, 0, err
		}
		var c domain.Comment
		if err := doc.DataTo(&c); err != nil {
			continue
		}
		c.ID = doc.Ref.ID
		comments = append(comments, &c)
	}
	return comments, total, nil
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

