package infrastructure

import (
	"context"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/lriverd/big-service/internal/pescaapp/species/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SpeciesFirestoreRepository struct {
	client     *firestore.Client
	collection string
}

func NewSpeciesRepository(client *firestore.Client) *SpeciesFirestoreRepository {
	return &SpeciesFirestoreRepository{client: client, collection: "species"}
}

func (r *SpeciesFirestoreRepository) FindByID(ctx context.Context, id string) (*domain.Species, error) {
	doc, err := r.client.Collection(r.collection).Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	var sp domain.Species
	if err := doc.DataTo(&sp); err != nil {
		return nil, err
	}
	sp.ID = doc.Ref.ID
	return &sp, nil
}

func (r *SpeciesFirestoreRepository) List(ctx context.Context, limit, offset int, search string) ([]*domain.Species, int, error) {
	query := r.client.Collection(r.collection).OrderBy("commonName", firestore.Asc)

	if search != "" {
		lower := strings.ToLower(search)
		query = r.client.Collection(r.collection).
			Where("commonName", ">=", lower).
			Where("commonName", "<=", lower+"\uf8ff")
	}

	all, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, 0, err
	}
	total := len(all)

	iter := query.Offset(offset).Limit(limit).Documents(ctx)
	var species []*domain.Species
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, 0, err
		}
		var sp domain.Species
		if err := doc.DataTo(&sp); err != nil {
			continue
		}
		sp.ID = doc.Ref.ID
		species = append(species, &sp)
	}
	return species, total, nil
}

func (r *SpeciesFirestoreRepository) Create(ctx context.Context, sp *domain.Species) (*domain.Species, error) {
	now := time.Now().UTC()
	sp.CreatedAt = now
	sp.UpdatedAt = now

	ref, _, err := r.client.Collection(r.collection).Add(ctx, map[string]interface{}{
		"commonName":      sp.CommonName,
		"scientificName":  sp.ScientificName,
		"description":     sp.Description,
		"imageUrl":        sp.ImageURL,
		"habitat":         sp.Habitat,
		"diet":            sp.Diet,
		"averageSizeCm":   sp.AverageSizeCm,
		"averageWeightKg": sp.AverageWeightKg,
		"maxSizeCm":       sp.MaxSizeCm,
		"maxWeightKg":     sp.MaxWeightKg,
		"fishingTips":     sp.FishingTips,
		"createdAt":       sp.CreatedAt,
		"updatedAt":       sp.UpdatedAt,
	})
	if err != nil {
		return nil, err
	}
	sp.ID = ref.ID
	return sp, nil
}

func (r *SpeciesFirestoreRepository) Update(ctx context.Context, id string, req domain.UpdateSpeciesRequest) (*domain.Species, error) {
	updates := []firestore.Update{{Path: "updatedAt", Value: time.Now().UTC()}}

	if req.CommonName != nil {
		updates = append(updates, firestore.Update{Path: "commonName", Value: *req.CommonName})
	}
	if req.ScientificName != nil {
		updates = append(updates, firestore.Update{Path: "scientificName", Value: *req.ScientificName})
	}
	if req.Description != nil {
		updates = append(updates, firestore.Update{Path: "description", Value: *req.Description})
	}
	if req.ImageURL != nil {
		updates = append(updates, firestore.Update{Path: "imageUrl", Value: *req.ImageURL})
	}
	if req.Habitat != nil {
		updates = append(updates, firestore.Update{Path: "habitat", Value: *req.Habitat})
	}
	if req.Diet != nil {
		updates = append(updates, firestore.Update{Path: "diet", Value: *req.Diet})
	}
	if req.AverageSizeCm != nil {
		updates = append(updates, firestore.Update{Path: "averageSizeCm", Value: *req.AverageSizeCm})
	}
	if req.AverageWeightKg != nil {
		updates = append(updates, firestore.Update{Path: "averageWeightKg", Value: *req.AverageWeightKg})
	}
	if req.MaxSizeCm != nil {
		updates = append(updates, firestore.Update{Path: "maxSizeCm", Value: *req.MaxSizeCm})
	}
	if req.MaxWeightKg != nil {
		updates = append(updates, firestore.Update{Path: "maxWeightKg", Value: *req.MaxWeightKg})
	}
	if req.FishingTips != nil {
		updates = append(updates, firestore.Update{Path: "fishingTips", Value: req.FishingTips})
	}

	_, err := r.client.Collection(r.collection).Doc(id).Update(ctx, updates)
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *SpeciesFirestoreRepository) Delete(ctx context.Context, id string) error {
	_, err := r.client.Collection(r.collection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return err
		}
		return err
	}
	_, err = r.client.Collection(r.collection).Doc(id).Delete(ctx)
	return err
}

func (r *SpeciesFirestoreRepository) Search(ctx context.Context, query string, limit int) ([]*domain.Species, error) {
	lower := strings.ToLower(query)
	iter := r.client.Collection(r.collection).
		Where("commonName", ">=", lower).
		Where("commonName", "<=", lower+"\uf8ff").
		Limit(limit).Documents(ctx)

	var species []*domain.Species
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var sp domain.Species
		if err := doc.DataTo(&sp); err != nil {
			continue
		}
		sp.ID = doc.Ref.ID
		species = append(species, &sp)
	}
	return species, nil
}

