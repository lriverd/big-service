package domain_test

import (
	"testing"
	"time"

	"github.com/lriverd/big-service/internal/pescaapp/statistics/domain"
)

func TestSpotStats(t *testing.T) {
	now := time.Now()
	s := domain.SpotStats{
		SpotID: "s1", Name: "Test", Visits: 156, UniqueUsers: 42,
		AverageRating: 4.3, TotalRatings: 12, TotalComments: 3,
		TopSpecies:      []domain.TopSpecies{{SpeciesID: "sp1", Name: "Trucha", Mentions: 8}},
		LastCommentDate: &now, CreatedAt: now,
	}
	if s.Visits != 156 || len(s.TopSpecies) != 1 {
		t.Error("unexpected stats fields")
	}
}

func TestUserDetailedStats(t *testing.T) {
	u := domain.UserDetailedStats{
		UserID: "u1", Name: "Test", SpotsCreated: 5, CommentsCount: 12,
		RatingsCount: 8, AverageRating: 4.5,
		FavoriteRegions: []string{"RM"}, FavoriteSpecies: []string{"sp1"},
	}
	if u.SpotsCreated != 5 || len(u.FavoriteRegions) != 1 {
		t.Error("unexpected user stats fields")
	}
}

func TestPopularSpot(t *testing.T) {
	p := domain.PopularSpot{
		ID: "s1", Name: "Test", Region: "RM",
		AverageRating: 4.5, TotalRatings: 20, Views: 500, TotalComments: 15,
	}
	if p.Views != 500 {
		t.Errorf("expected 500 views, got %d", p.Views)
	}
}

func TestTopSpecies(t *testing.T) {
	ts := domain.TopSpecies{SpeciesID: "sp1", Name: "Trucha", Mentions: 8}
	if ts.Mentions != 8 {
		t.Errorf("expected 8 mentions, got %d", ts.Mentions)
	}
}

