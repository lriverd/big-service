package domain_test

import (
	"testing"
	"time"

	"github.com/lriverd/big-service/internal/pescaapp/ratings/domain"
)

func TestRating(t *testing.T) {
	r := domain.Rating{
		ID: "r1", SpotID: "s1", UserID: "u1", Stars: 4,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if r.Stars != 4 {
		t.Errorf("expected 4 stars, got %d", r.Stars)
	}
}

func TestRatingWithUser(t *testing.T) {
	r := domain.Rating{
		ID: "r1", User: &domain.UserInfo{ID: "u1", Name: "Test"},
	}
	if r.User.Name != "Test" {
		t.Error("unexpected user name")
	}
}

func TestRatingStats(t *testing.T) {
	stats := domain.RatingStats{
		AverageRating: 4.3, TotalRatings: 12,
		Distribution: map[string]int{"5": 5, "4": 4, "3": 2, "2": 1, "1": 0},
	}
	if stats.Distribution["5"] != 5 {
		t.Errorf("expected 5 five-star ratings, got %d", stats.Distribution["5"])
	}
	if stats.TotalRatings != 12 {
		t.Errorf("expected 12 total, got %d", stats.TotalRatings)
	}
}

func TestCreateRatingRequest(t *testing.T) {
	req := domain.CreateRatingRequest{Stars: 5}
	if req.Stars != 5 {
		t.Errorf("expected 5 stars, got %d", req.Stars)
	}
}

