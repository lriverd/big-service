package domain_test

import (
	"testing"

	"github.com/lriverd/big-service/internal/pescaapp/species/domain"
)

func TestSpecies(t *testing.T) {
	sp := domain.Species{
		ID: "sp1", CommonName: "Trucha", ScientificName: "Oncorhynchus mykiss",
		AverageSizeCm: 35, MaxWeightKg: 9.0, FishingTips: []string{"tip1"},
	}
	if sp.CommonName != "Trucha" || len(sp.FishingTips) != 1 {
		t.Error("unexpected species fields")
	}
}

func TestCreateSpeciesRequest(t *testing.T) {
	req := domain.CreateSpeciesRequest{
		CommonName: "Trucha", ScientificName: "O. mykiss", Description: "desc",
		AverageSizeCm: 35, FishingTips: []string{"t1", "t2"},
	}
	if len(req.FishingTips) != 2 {
		t.Errorf("expected 2 tips, got %d", len(req.FishingTips))
	}
}

func TestUpdateSpeciesRequest(t *testing.T) {
	name := "Updated"
	req := domain.UpdateSpeciesRequest{CommonName: &name}
	if *req.CommonName != "Updated" {
		t.Error("unexpected name")
	}
}

