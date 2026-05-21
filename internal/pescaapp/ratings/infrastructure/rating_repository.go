package infrastructure

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/lriverd/big-service/internal/pescaapp/ratings/domain"
	"google.golang.org/api/iterator"
)

type RatingFirestoreRepository struct {
	client     *firestore.Client
	collection string
}

func NewRatingRepository(client *firestore.Client) *RatingFirestoreRepository {
	return &RatingFirestoreRepository{client: client, collection: "ratings"}
}

func (r *RatingFirestoreRepository) FindBySpotAndUser(ctx context.Context, spotID, userID string) (*domain.Rating, error) {
	iter := r.client.Collection(r.collection).
		Where("spotId", "==", spotID).
		Where("userId", "==", userID).
		Limit(1).Documents(ctx)

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("rating not found")
	}
	if err != nil {
		return nil, err
	}
	var rating domain.Rating
	if err := doc.DataTo(&rating); err != nil {
		return nil, err
	}
	rating.ID = doc.Ref.ID
	return &rating, nil
}

func (r *RatingFirestoreRepository) ListBySpot(ctx context.Context, spotID string, limit, offset int) ([]*domain.Rating, int, error) {
	query := r.client.Collection(r.collection).
		Where("spotId", "==", spotID).
		OrderBy("createdAt", firestore.Desc)

	all, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, 0, err
	}
	total := len(all)

	iter := query.Offset(offset).Limit(limit).Documents(ctx)
	var ratings []*domain.Rating
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, 0, err
		}
		var rat domain.Rating
		if err := doc.DataTo(&rat); err != nil {
			continue
		}
		rat.ID = doc.Ref.ID
		ratings = append(ratings, &rat)
	}
	return ratings, total, nil
}

func (r *RatingFirestoreRepository) Create(ctx context.Context, rating *domain.Rating) (*domain.Rating, error) {
	ref, _, err := r.client.Collection(r.collection).Add(ctx, map[string]interface{}{
		"spotId":    rating.SpotID,
		"userId":    rating.UserID,
		"stars":     rating.Stars,
		"createdAt": rating.CreatedAt,
		"updatedAt": rating.UpdatedAt,
	})
	if err != nil {
		return nil, err
	}
	rating.ID = ref.ID
	return rating, nil
}

func (r *RatingFirestoreRepository) Update(ctx context.Context, id string, stars int) (*domain.Rating, error) {
	now := time.Now().UTC()
	_, err := r.client.Collection(r.collection).Doc(id).Update(ctx, []firestore.Update{
		{Path: "stars", Value: stars},
		{Path: "updatedAt", Value: now},
	})
	if err != nil {
		return nil, err
	}
	doc, err := r.client.Collection(r.collection).Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	var rating domain.Rating
	if err := doc.DataTo(&rating); err != nil {
		return nil, err
	}
	rating.ID = doc.Ref.ID
	return &rating, nil
}

func (r *RatingFirestoreRepository) Delete(ctx context.Context, spotID, userID string) error {
	iter := r.client.Collection(r.collection).
		Where("spotId", "==", spotID).
		Where("userId", "==", userID).
		Limit(1).Documents(ctx)

	doc, err := iter.Next()
	if err == iterator.Done {
		return fmt.Errorf("rating not found")
	}
	if err != nil {
		return err
	}
	_, err = doc.Ref.Delete(ctx)
	return err
}

func (r *RatingFirestoreRepository) GetStats(ctx context.Context, spotID string) (*domain.RatingStats, error) {
	iter := r.client.Collection(r.collection).
		Where("spotId", "==", spotID).Documents(ctx)

	distribution := map[string]int{"1": 0, "2": 0, "3": 0, "4": 0, "5": 0}
	totalStars := 0
	count := 0

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var rat domain.Rating
		if err := doc.DataTo(&rat); err != nil {
			continue
		}
		distribution[fmt.Sprintf("%d", rat.Stars)]++
		totalStars += rat.Stars
		count++
	}

	avg := 0.0
	if count > 0 {
		avg = float64(totalStars) / float64(count)
	}

	return &domain.RatingStats{
		AverageRating: avg,
		TotalRatings:  count,
		Distribution:  distribution,
	}, nil
}

func (r *RatingFirestoreRepository) CountByUser(ctx context.Context, userID string) (int, error) {
	docs, err := r.client.Collection(r.collection).
		Where("userId", "==", userID).Documents(ctx).GetAll()
	if err != nil {
		return 0, err
	}
	return len(docs), nil
}

