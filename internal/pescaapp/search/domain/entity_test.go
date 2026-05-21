package domain_test

import (
	"testing"

	"github.com/lriverd/big-service/internal/pescaapp/search/domain"
)

func TestSearchResult(t *testing.T) {
	r := domain.SearchResult{
		Spots:   []domain.SpotResult{{ID: "s1", Name: "Spot", Region: "RM", Rating: 4.5}},
		Species: []domain.SpeciesResult{{ID: "sp1", CommonName: "Trucha", ScientificName: "O.m"}},
		Users:   []domain.UserResult{{ID: "u1", Name: "User"}},
	}
	if len(r.Spots) != 1 || len(r.Species) != 1 || len(r.Users) != 1 {
		t.Error("unexpected result counts")
	}
	if r.Spots[0].Rating != 4.5 {
		t.Errorf("expected rating 4.5")
	}
}

func TestSpotResult(t *testing.T) {
	sr := domain.SpotResult{ID: "s1", Name: "Spot", Region: "RM", Rating: 3.0}
	if sr.ID != "s1" {
		t.Error("unexpected id")
	}
}

func TestSpeciesResult(t *testing.T) {
	sr := domain.SpeciesResult{ID: "sp1", CommonName: "Trucha", ScientificName: "O.m"}
	if sr.CommonName != "Trucha" {
		t.Error("unexpected name")
	}
}

func TestUserResult(t *testing.T) {
	ur := domain.UserResult{ID: "u1", Name: "User"}
	if ur.Name != "User" {
		t.Error("unexpected name")
	}
}

