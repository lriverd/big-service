package infrastructure

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/lriverd/big-service/internal/pescaapp/spots/domain"
	"google.golang.org/api/iterator"
)

type SpotSpeciesFirestoreRepository struct {
	client     *firestore.Client
	collection string
}

func NewSpotSpeciesRepository(client *firestore.Client) *SpotSpeciesFirestoreRepository {
	return &SpotSpeciesFirestoreRepository{client: client, collection: "spot_species"}
}

func (r *SpotSpeciesFirestoreRepository) ListBySpot(ctx context.Context, spotID string) ([]domain.SpotSpecies, error) {
	iter := r.client.Collection(r.collection).Where("spotId", "==", spotID).Documents(ctx)
	var species []domain.SpotSpecies
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var sp domain.SpotSpecies
		if err := doc.DataTo(&sp); err != nil {
			continue
		}
		sp.ID = doc.Ref.ID
		species = append(species, sp)
	}
	return species, nil
}

func (r *SpotSpeciesFirestoreRepository) SetForSpot(ctx context.Context, spotID string, species []domain.SpotSpecies) error {
	// Delete existing
	if err := r.DeleteBySpot(ctx, spotID); err != nil {
		return err
	}
	// Add new
	for _, sp := range species {
		_, _, err := r.client.Collection(r.collection).Add(ctx, map[string]interface{}{
			"spotId":           spotID,
			"speciesId":        sp.SpeciesID,
			"recommendedBaits": sp.RecommendedBaits,
			"recommendedRod":   sp.RecommendedRod,
			"recommendedLine":  sp.RecommendedLine,
			"recommendedHook":  sp.RecommendedHook,
			"bestSeason":       sp.BestSeason,
			"difficulty":       sp.Difficulty,
			"notes":            sp.Notes,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *SpotSpeciesFirestoreRepository) DeleteBySpot(ctx context.Context, spotID string) error {
	iter := r.client.Collection(r.collection).Where("spotId", "==", spotID).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		if _, err := doc.Ref.Delete(ctx); err != nil {
			return err
		}
	}
	return nil
}

