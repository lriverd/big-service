package infrastructure

import (
	"context"
	"math"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/lriverd/big-service/internal/pescaapp/spots/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SpotFirestoreRepository struct {
	client     *firestore.Client
	collection string
}

func NewSpotRepository(client *firestore.Client) *SpotFirestoreRepository {
	return &SpotFirestoreRepository{client: client, collection: "fishing_spots"}
}

func (r *SpotFirestoreRepository) FindByID(ctx context.Context, id string) (*domain.Spot, error) {
	doc, err := r.client.Collection(r.collection).Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	var spot domain.Spot
	if err := doc.DataTo(&spot); err != nil {
		return nil, err
	}
	spot.ID = doc.Ref.ID
	return &spot, nil
}

func (r *SpotFirestoreRepository) List(ctx context.Context, limit, offset int, filter domain.SpotFilter) ([]*domain.Spot, int, error) {
	query := r.client.Collection(r.collection).Query

	if filter.Region != "" {
		query = query.Where("region", "==", filter.Region)
	}
	if filter.WaterType != "" {
		query = query.Where("waterType", "==", filter.WaterType)
	}
	if filter.BoatRequired != nil {
		query = query.Where("boatRequired", "==", *filter.BoatRequired)
	}

	switch filter.SortBy {
	case "rating":
		query = query.OrderBy("averageRating", firestore.Desc)
	case "recent":
		query = query.OrderBy("createdAt", firestore.Desc)
	default:
		query = query.OrderBy("createdAt", firestore.Desc)
	}

	all, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, 0, err
	}

	// Firestore can't express "status not in (HIDDEN, DELETED)" alongside the
	// other filters/legacy-empty-status case in a single query, so visibility
	// and (optional) geo filtering are both applied in memory here.
	var visible []*domain.Spot
	for _, doc := range all {
		var s domain.Spot
		if err := doc.DataTo(&s); err != nil {
			continue
		}
		s.ID = doc.Ref.ID
		if !s.IsVisible() {
			continue
		}
		visible = append(visible, &s)
	}

	if filter.Latitude != nil && filter.Longitude != nil && filter.RadiusKm != nil {
		var geoFiltered []*domain.Spot
		for _, s := range visible {
			dist := haversine(s.Latitude, s.Longitude, *filter.Latitude, *filter.Longitude)
			if dist <= *filter.RadiusKm {
				geoFiltered = append(geoFiltered, s)
			}
		}
		visible = geoFiltered
	}

	total := len(visible)
	end := offset + limit
	if end > len(visible) {
		end = len(visible)
	}
	if offset >= len(visible) {
		return []*domain.Spot{}, total, nil
	}
	return visible[offset:end], total, nil
}

func (r *SpotFirestoreRepository) Create(ctx context.Context, spot *domain.Spot) (*domain.Spot, error) {
	now := time.Now().UTC()
	spot.CreatedAt = now
	spot.UpdatedAt = now
	spot.Status = string(domain.SpotStatusPending)

	ref, _, err := r.client.Collection(r.collection).Add(ctx, map[string]interface{}{
		"name":            spot.Name,
		"description":     spot.Description,
		"latitude":        spot.Latitude,
		"longitude":       spot.Longitude,
		"region":          spot.Region,
		"waterType":       spot.WaterType,
		"boatAllowed":     spot.BoatAllowed,
		"boatRequired":    spot.BoatRequired,
		"access":          spot.Access,
		"createdByUserId": spot.CreatedByUserID,
		"views":           0,
		"averageRating":   0,
		"totalRatings":    0,
		"totalComments":   0,
		"status":          spot.Status,
		"reportCount":     0,
		"createdAt":       spot.CreatedAt,
		"updatedAt":       spot.UpdatedAt,
	})
	if err != nil {
		return nil, err
	}
	spot.ID = ref.ID
	return spot, nil
}

func (r *SpotFirestoreRepository) Update(ctx context.Context, id string, req domain.UpdateSpotRequest) (*domain.Spot, error) {
	updates := []firestore.Update{{Path: "updatedAt", Value: time.Now().UTC()}}

	if req.Name != nil {
		updates = append(updates, firestore.Update{Path: "name", Value: *req.Name})
	}
	if req.Description != nil {
		updates = append(updates, firestore.Update{Path: "description", Value: *req.Description})
	}
	if req.Latitude != nil {
		updates = append(updates, firestore.Update{Path: "latitude", Value: *req.Latitude})
	}
	if req.Longitude != nil {
		updates = append(updates, firestore.Update{Path: "longitude", Value: *req.Longitude})
	}
	if req.Region != nil {
		updates = append(updates, firestore.Update{Path: "region", Value: *req.Region})
	}
	if req.WaterType != nil {
		updates = append(updates, firestore.Update{Path: "waterType", Value: *req.WaterType})
	}
	if req.BoatAllowed != nil {
		updates = append(updates, firestore.Update{Path: "boatAllowed", Value: *req.BoatAllowed})
	}
	if req.BoatRequired != nil {
		updates = append(updates, firestore.Update{Path: "boatRequired", Value: *req.BoatRequired})
	}
	if req.Access != nil {
		updates = append(updates, firestore.Update{Path: "access", Value: *req.Access})
	}

	_, err := r.client.Collection(r.collection).Doc(id).Update(ctx, updates)
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *SpotFirestoreRepository) Delete(ctx context.Context, id string) error {
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

func (r *SpotFirestoreRepository) FindNearby(ctx context.Context, lat, lng, radiusKm float64, limit int) ([]*domain.Spot, error) {
	// Firestore doesn't support geo queries natively; fetch all and filter in memory
	iter := r.client.Collection(r.collection).Documents(ctx)
	var spots []*domain.Spot
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var s domain.Spot
		if err := doc.DataTo(&s); err != nil {
			continue
		}
		s.ID = doc.Ref.ID
		if !s.IsVisible() {
			continue
		}
		dist := haversine(s.Latitude, s.Longitude, lat, lng)
		if dist <= radiusKm && dist > 0 {
			spots = append(spots, &s)
		}
		if len(spots) >= limit {
			break
		}
	}
	return spots, nil
}

func (r *SpotFirestoreRepository) UpdateStatus(ctx context.Context, id string, status domain.SpotStatus) (*domain.Spot, error) {
	_, err := r.client.Collection(r.collection).Doc(id).Update(ctx, []firestore.Update{
		{Path: "status", Value: string(status)},
		{Path: "updatedAt", Value: time.Now().UTC()},
	})
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

// CountCreatedSince counts spots a user has created at or after the given
// time. Requires a composite index on (createdByUserId ASC, createdAt ASC)
// in Firestore — see FIRESTORE_COLLECTIONS.md.
func (r *SpotFirestoreRepository) CountCreatedSince(ctx context.Context, userID string, since time.Time) (int, error) {
	docs, err := r.client.Collection(r.collection).
		Where("createdByUserId", "==", userID).
		Where("createdAt", ">=", since).
		Documents(ctx).GetAll()
	if err != nil {
		return 0, err
	}
	return len(docs), nil
}

// FindNearbyForDuplicateCheck finds visible spots within radiusMeters of the
// given point, sorted by distance ascending, capped at maxResults. Used to
// warn a user creating a spot that a very similar one may already exist.
func (r *SpotFirestoreRepository) FindNearbyForDuplicateCheck(ctx context.Context, lat, lng, radiusMeters float64, maxResults int) ([]domain.DuplicateCandidate, error) {
	iter := r.client.Collection(r.collection).Documents(ctx)
	radiusKm := radiusMeters / 1000
	var candidates []domain.DuplicateCandidate
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var s domain.Spot
		if err := doc.DataTo(&s); err != nil {
			continue
		}
		s.ID = doc.Ref.ID
		if !s.IsVisible() {
			continue
		}
		distKm := haversine(s.Latitude, s.Longitude, lat, lng)
		if distKm <= radiusKm {
			candidates = append(candidates, domain.DuplicateCandidate{Spot: &s, DistanceMeters: distKm * 1000})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].DistanceMeters < candidates[j].DistanceMeters
	})
	if len(candidates) > maxResults {
		candidates = candidates[:maxResults]
	}
	return candidates, nil
}

func (r *SpotFirestoreRepository) FindByCreatedByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Spot, int, error) {
	query := r.client.Collection(r.collection).
		Where("createdByUserId", "==", userID).
		OrderBy("createdAt", firestore.Desc)

	all, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, 0, err
	}

	var spots []*domain.Spot
	for _, doc := range all {
		var s domain.Spot
		if err := doc.DataTo(&s); err != nil {
			continue
		}
		s.ID = doc.Ref.ID
		spots = append(spots, &s)
	}

	total := len(spots)
	end := offset + limit
	if end > len(spots) {
		end = len(spots)
	}
	if offset >= len(spots) {
		return []*domain.Spot{}, total, nil
	}
	return spots[offset:end], total, nil
}

func (r *SpotFirestoreRepository) IncrementViews(ctx context.Context, id string) error {
	_, err := r.client.Collection(r.collection).Doc(id).Update(ctx, []firestore.Update{
		{Path: "views", Value: firestore.Increment(1)},
	})
	return err
}

func (r *SpotFirestoreRepository) UpdateRatingStats(ctx context.Context, id string, avgRating float64, totalRatings int) error {
	_, err := r.client.Collection(r.collection).Doc(id).Update(ctx, []firestore.Update{
		{Path: "averageRating", Value: avgRating},
		{Path: "totalRatings", Value: totalRatings},
	})
	return err
}

func (r *SpotFirestoreRepository) UpdateCommentCount(ctx context.Context, id string, delta int) error {
	_, err := r.client.Collection(r.collection).Doc(id).Update(ctx, []firestore.Update{
		{Path: "totalComments", Value: firestore.Increment(delta)},
	})
	return err
}

func (r *SpotFirestoreRepository) Search(ctx context.Context, query string, limit int) ([]*domain.Spot, error) {
	lower := strings.ToLower(query)
	iter := r.client.Collection(r.collection).
		Where("name", ">=", lower).
		Where("name", "<=", lower+"\uf8ff").
		Limit(limit).Documents(ctx)

	var spots []*domain.Spot
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var s domain.Spot
		if err := doc.DataTo(&s); err != nil {
			continue
		}
		s.ID = doc.Ref.ID
		spots = append(spots, &s)
	}
	return spots, nil
}

// haversine calculates the distance between two points in km
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth radius in km
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
