package application

import (
	"context"
	"fmt"
	"time"

	"github.com/lriverd/big-service/internal/pescaapp/spots/domain"
	statsDomain "github.com/lriverd/big-service/internal/pescaapp/statistics/domain"
	usersDomain "github.com/lriverd/big-service/internal/pescaapp/users/domain"
	"github.com/lriverd/big-service/internal/platform/cache"
	apperrors "github.com/lriverd/big-service/internal/shared/errors"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type StatsService struct {
	fsClient *firestore.Client
	spotRepo domain.SpotRepository
	userRepo usersDomain.UserRepository
	cache    *cache.Cache
}

func NewStatsService(
	fsClient *firestore.Client,
	spotRepo domain.SpotRepository,
	userRepo usersDomain.UserRepository,
	c *cache.Cache,
) *StatsService {
	return &StatsService{fsClient: fsClient, spotRepo: spotRepo, userRepo: userRepo, cache: c}
}

func (s *StatsService) GetSpotStats(ctx context.Context, spotID string) (*statsDomain.SpotStats, error) {
	cacheKey := fmt.Sprintf("stats:spot:%s", spotID)
	if cached, found := s.cache.Get(cacheKey); found {
		return cached.(*statsDomain.SpotStats), nil
	}

	spot, err := s.spotRepo.FindByID(ctx, spotID)
	if err != nil {
		return nil, apperrors.NotFound("Spot")
	}

	// Count unique users from ratings and comments
	ratingDocs, _ := s.fsClient.Collection("ratings").Where("spotId", "==", spotID).Documents(ctx).GetAll()
	commentDocs, _ := s.fsClient.Collection("comments").Where("spotId", "==", spotID).Where("deletedAt", "==", nil).Documents(ctx).GetAll()

	userSet := make(map[string]bool)
	for _, d := range ratingDocs {
		if uid, ok := d.Data()["userId"].(string); ok {
			userSet[uid] = true
		}
	}
	for _, d := range commentDocs {
		if uid, ok := d.Data()["userId"].(string); ok {
			userSet[uid] = true
		}
	}

	// Last comment date
	var lastCommentDate *time.Time
	if len(commentDocs) > 0 {
		last := commentDocs[len(commentDocs)-1]
		if t, ok := last.Data()["createdAt"].(time.Time); ok {
			lastCommentDate = &t
		}
	}

	// Top species
	speciesDocs, _ := s.fsClient.Collection("spot_species").Where("spotId", "==", spotID).Documents(ctx).GetAll()
	var topSpecies []statsDomain.TopSpecies
	for _, d := range speciesDocs {
		spID, _ := d.Data()["speciesId"].(string)
		topSpecies = append(topSpecies, statsDomain.TopSpecies{
			SpeciesID: spID,
			Mentions:  1,
		})
	}

	stats := &statsDomain.SpotStats{
		SpotID:          spot.ID,
		Name:            spot.Name,
		Visits:          spot.Views,
		UniqueUsers:     len(userSet),
		AverageRating:   spot.AverageRating,
		TotalRatings:    spot.TotalRatings,
		TotalComments:   spot.TotalComments,
		TopSpecies:      topSpecies,
		LastCommentDate: lastCommentDate,
		CreatedAt:       spot.CreatedAt,
	}
	s.cache.Set(cacheKey, stats, 5*time.Minute)
	return stats, nil
}

func (s *StatsService) GetUserStats(ctx context.Context, userID string) (*statsDomain.UserDetailedStats, error) {
	cacheKey := fmt.Sprintf("stats:user:%s", userID)
	if cached, found := s.cache.Get(cacheKey); found {
		return cached.(*statsDomain.UserDetailedStats), nil
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperrors.NotFound("User")
	}

	spotDocs, _ := s.fsClient.Collection("fishing_spots").Where("createdByUserId", "==", userID).Documents(ctx).GetAll()
	commentDocs, _ := s.fsClient.Collection("comments").Where("userId", "==", userID).Where("deletedAt", "==", nil).Documents(ctx).GetAll()
	ratingDocs, _ := s.fsClient.Collection("ratings").Where("userId", "==", userID).Documents(ctx).GetAll()

	// Favorite regions
	regionCount := make(map[string]int)
	for _, d := range spotDocs {
		if r, ok := d.Data()["region"].(string); ok {
			regionCount[r]++
		}
	}
	var favRegions []string
	for r := range regionCount {
		favRegions = append(favRegions, r)
	}

	// Average rating given
	totalStars := 0
	for _, d := range ratingDocs {
		if stars, ok := d.Data()["stars"].(int64); ok {
			totalStars += int(stars)
		}
	}
	avgRating := 0.0
	if len(ratingDocs) > 0 {
		avgRating = float64(totalStars) / float64(len(ratingDocs))
	}

	stats := &statsDomain.UserDetailedStats{
		UserID:          user.ID,
		Name:            user.Name,
		SpotsCreated:    len(spotDocs),
		CommentsCount:   len(commentDocs),
		RatingsCount:    len(ratingDocs),
		AverageRating:   avgRating,
		FavoriteRegions: favRegions,
		FavoriteSpecies: []string{},
		JoiningDate:     user.CreatedAt.Format(time.RFC3339),
	}
	s.cache.Set(cacheKey, stats, 5*time.Minute)
	return stats, nil
}

func (s *StatsService) GetPopularSpots(ctx context.Context, limit int, orderBy string) ([]*statsDomain.PopularSpot, error) {
	cacheKey := fmt.Sprintf("stats:popular:%d:%s", limit, orderBy)
	if cached, found := s.cache.Get(cacheKey); found {
		return cached.([]*statsDomain.PopularSpot), nil
	}

	query := s.fsClient.Collection("fishing_spots").Query
	switch orderBy {
	case "visits":
		query = query.OrderBy("views", firestore.Desc)
	case "comments":
		query = query.OrderBy("totalComments", firestore.Desc)
	default: // rating
		query = query.OrderBy("averageRating", firestore.Desc)
	}

	iter := query.Limit(limit).Documents(ctx)
	var spots []*statsDomain.PopularSpot
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		data := doc.Data()
		spot := &statsDomain.PopularSpot{
			ID:   doc.Ref.ID,
			Name: getString(data, "name"),
			Region: getString(data, "region"),
		}
		if v, ok := data["averageRating"].(float64); ok {
			spot.AverageRating = v
		}
		if v, ok := data["totalRatings"].(int64); ok {
			spot.TotalRatings = int(v)
		}
		if v, ok := data["views"].(int64); ok {
			spot.Views = int(v)
		}
		if v, ok := data["totalComments"].(int64); ok {
			spot.TotalComments = int(v)
		}
		spots = append(spots, spot)
	}

	s.cache.Set(cacheKey, spots, 10*time.Minute)
	return spots, nil
}

func getString(data map[string]interface{}, key string) string {
	if v, ok := data[key].(string); ok {
		return v
	}
	return ""
}

