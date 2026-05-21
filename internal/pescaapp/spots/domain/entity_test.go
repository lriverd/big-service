package domain_test

import (
	"testing"
	"time"

	"github.com/lriverd/big-service/internal/pescaapp/spots/domain"
)

func TestSpot(t *testing.T) {
	spot := domain.Spot{
		ID: "s1", Name: "Rio Maipo", Latitude: -33.89, Longitude: -70.55,
		Region: "RM", WaterType: "river", AverageRating: 4.3, TotalRatings: 12,
		CreatedAt: time.Now(),
	}
	if spot.Name != "Rio Maipo" || spot.AverageRating != 4.3 {
		t.Error("unexpected spot fields")
	}
}

func TestSpotSpecies(t *testing.T) {
	ss := domain.SpotSpecies{
		SpeciesID: "sp1", RecommendedBaits: []string{"mosca"}, Difficulty: "medium",
	}
	if ss.Difficulty != "medium" || len(ss.RecommendedBaits) != 1 {
		t.Error("unexpected spot species fields")
	}
}

func TestCreateSpotRequest(t *testing.T) {
	req := domain.CreateSpotRequest{
		Name: "Test", Description: "desc", Latitude: -33.0, Longitude: -70.0,
		Region: "RM", WaterType: "lake",
	}
	if req.Region != "RM" {
		t.Error("unexpected region")
	}
}

func TestUpdateSpotRequest(t *testing.T) {
	name := "Updated"
	req := domain.UpdateSpotRequest{Name: &name}
	if *req.Name != "Updated" {
		t.Error("unexpected name")
	}
}

func TestSpotFilter(t *testing.T) {
	lat := -33.0
	radius := 10.0
	br := true
	f := domain.SpotFilter{
		Region: "RM", WaterType: "river", BoatRequired: &br,
		Latitude: &lat, RadiusKm: &radius, SortBy: "rating",
	}
	if *f.BoatRequired != true || *f.RadiusKm != 10.0 {
		t.Error("unexpected filter fields")
	}
}

