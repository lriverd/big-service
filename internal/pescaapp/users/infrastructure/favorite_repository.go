package infrastructure

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/lriverd/big-service/internal/pescaapp/users/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FavoriteFirestoreRepository struct {
	client     *firestore.Client
	collection string
}

func NewFavoriteRepository(client *firestore.Client) *FavoriteFirestoreRepository {
	return &FavoriteFirestoreRepository{client: client, collection: "favorites"}
}

func (r *FavoriteFirestoreRepository) List(ctx context.Context, userID string, limit, offset int) ([]*domain.Favorite, int, error) {
	query := r.client.Collection(r.collection).
		Where("userId", "==", userID).
		OrderBy("createdAt", firestore.Desc)

	all, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, 0, err
	}
	total := len(all)

	iter := query.Offset(offset).Limit(limit).Documents(ctx)
	var favs []*domain.Favorite
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, 0, err
		}
		var f domain.Favorite
		if err := doc.DataTo(&f); err != nil {
			continue
		}
		f.ID = doc.Ref.ID
		favs = append(favs, &f)
	}
	return favs, total, nil
}

func (r *FavoriteFirestoreRepository) Add(ctx context.Context, fav *domain.Favorite) error {
	fav.CreatedAt = time.Now().UTC()
	_, _, err := r.client.Collection(r.collection).Add(ctx, map[string]interface{}{
		"userId":    fav.UserID,
		"spotId":    fav.SpotID,
		"createdAt": fav.CreatedAt,
	})
	return err
}

func (r *FavoriteFirestoreRepository) Remove(ctx context.Context, userID, spotID string) error {
	iter := r.client.Collection(r.collection).
		Where("userId", "==", userID).
		Where("spotId", "==", spotID).
		Limit(1).Documents(ctx)

	doc, err := iter.Next()
	if err == iterator.Done {
		return status.Error(codes.NotFound, "favorite not found")
	}
	if err != nil {
		return err
	}
	_, err = doc.Ref.Delete(ctx)
	return err
}

func (r *FavoriteFirestoreRepository) Exists(ctx context.Context, userID, spotID string) (bool, error) {
	iter := r.client.Collection(r.collection).
		Where("userId", "==", userID).
		Where("spotId", "==", spotID).
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

