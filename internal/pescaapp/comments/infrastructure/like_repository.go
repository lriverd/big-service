package infrastructure

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/lriverd/big-service/internal/pescaapp/comments/domain"
	"google.golang.org/api/iterator"
)

type CommentLikeFirestoreRepository struct {
	client     *firestore.Client
	collection string
}

func NewCommentLikeRepository(client *firestore.Client) *CommentLikeFirestoreRepository {
	return &CommentLikeFirestoreRepository{client: client, collection: "comment_likes"}
}

func (r *CommentLikeFirestoreRepository) Exists(ctx context.Context, commentID, userID string) (bool, error) {
	iter := r.client.Collection(r.collection).
		Where("commentId", "==", commentID).
		Where("userId", "==", userID).
		Limit(1).Documents(ctx)

	_, err := iter.Next()
	if err == iterator.Done {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *CommentLikeFirestoreRepository) Create(ctx context.Context, like *domain.CommentLike) error {
	_, _, err := r.client.Collection(r.collection).Add(ctx, map[string]interface{}{
		"commentId": like.CommentID,
		"userId":    like.UserID,
		"createdAt": time.Now().UTC(),
	})
	return err
}

func (r *CommentLikeFirestoreRepository) Delete(ctx context.Context, commentID, userID string) error {
	iter := r.client.Collection(r.collection).
		Where("commentId", "==", commentID).
		Where("userId", "==", userID).
		Limit(1).Documents(ctx)

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil
	}
	if err != nil {
		return err
	}
	_, err = doc.Ref.Delete(ctx)
	return err
}

