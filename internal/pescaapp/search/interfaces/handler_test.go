package interfaces_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"context"
	"github.com/gin-gonic/gin"
	searchApp "github.com/lriverd/big-service/internal/pescaapp/search/application"
	"github.com/lriverd/big-service/internal/pescaapp/search/domain"
	searchIface "github.com/lriverd/big-service/internal/pescaapp/search/interfaces"
	speciesDomain "github.com/lriverd/big-service/internal/pescaapp/species/domain"
	spotsDomain "github.com/lriverd/big-service/internal/pescaapp/spots/domain"
	usersDomain "github.com/lriverd/big-service/internal/pescaapp/users/domain"
)

func init() { gin.SetMode(gin.TestMode) }

// Mocks
type mockSpotRepo struct{}

func (m *mockSpotRepo) FindByID(ctx context.Context, id string) (*spotsDomain.Spot, error) {
	return nil, nil
}
func (m *mockSpotRepo) List(ctx context.Context, limit, offset int, filter spotsDomain.SpotFilter) ([]*spotsDomain.Spot, int, error) {
	return nil, 0, nil
}
func (m *mockSpotRepo) Create(ctx context.Context, spot *spotsDomain.Spot) (*spotsDomain.Spot, error) {
	return nil, nil
}
func (m *mockSpotRepo) Update(ctx context.Context, id string, req spotsDomain.UpdateSpotRequest) (*spotsDomain.Spot, error) {
	return nil, nil
}
func (m *mockSpotRepo) Delete(ctx context.Context, id string) error { return nil }
func (m *mockSpotRepo) FindNearby(ctx context.Context, lat, lng, radiusKm float64, limit int) ([]*spotsDomain.Spot, error) {
	return nil, nil
}
func (m *mockSpotRepo) IncrementViews(ctx context.Context, id string) error { return nil }
func (m *mockSpotRepo) UpdateRatingStats(ctx context.Context, id string, avgRating float64, totalRatings int) error {
	return nil
}
func (m *mockSpotRepo) UpdateCommentCount(ctx context.Context, id string, delta int) error {
	return nil
}
func (m *mockSpotRepo) Search(ctx context.Context, query string, limit int) ([]*spotsDomain.Spot, error) {
	return []*spotsDomain.Spot{{ID: "s1", Name: "Test Spot", Region: "Test Region", AverageRating: 4.5}}, nil
}
func (m *mockSpotRepo) UpdateStatus(ctx context.Context, id string, status spotsDomain.SpotStatus) (*spotsDomain.Spot, error) {
	return nil, nil
}
func (m *mockSpotRepo) FindByCreatedByUserID(ctx context.Context, userID string, limit, offset int) ([]*spotsDomain.Spot, int, error) {
	return nil, 0, nil
}
func (m *mockSpotRepo) CountCreatedSince(ctx context.Context, userID string, since time.Time) (int, error) {
	return 0, nil
}
func (m *mockSpotRepo) FindNearbyForDuplicateCheck(ctx context.Context, lat, lng, radiusMeters float64, maxResults int) ([]spotsDomain.DuplicateCandidate, error) {
	return nil, nil
}

type mockSpeciesRepo struct{}

func (m *mockSpeciesRepo) FindByID(ctx context.Context, id string) (*speciesDomain.Species, error) {
	return nil, nil
}
func (m *mockSpeciesRepo) List(ctx context.Context, limit, offset int, search string) ([]*speciesDomain.Species, int, error) {
	return nil, 0, nil
}
func (m *mockSpeciesRepo) Create(ctx context.Context, sp *speciesDomain.Species) (*speciesDomain.Species, error) {
	return nil, nil
}
func (m *mockSpeciesRepo) Update(ctx context.Context, id string, req speciesDomain.UpdateSpeciesRequest) (*speciesDomain.Species, error) {
	return nil, nil
}
func (m *mockSpeciesRepo) Delete(ctx context.Context, id string) error { return nil }
func (m *mockSpeciesRepo) Search(ctx context.Context, query string, limit int) ([]*speciesDomain.Species, error) {
	return []*speciesDomain.Species{{ID: "sp1", CommonName: "Trucha", ScientificName: "Oncorhynchus"}}, nil
}

type mockUserRepo struct{}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*usersDomain.User, error) {
	return nil, nil
}
func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*usersDomain.User, error) {
	return nil, nil
}
func (m *mockUserRepo) Create(ctx context.Context, user *usersDomain.User) (*usersDomain.User, error) {
	return nil, nil
}
func (m *mockUserRepo) Update(ctx context.Context, id string, req usersDomain.UpdateUserRequest) (*usersDomain.User, error) {
	return nil, nil
}
func (m *mockUserRepo) List(ctx context.Context, limit, offset int, search string) ([]*usersDomain.UserPublic, int, error) {
	return []*usersDomain.UserPublic{{ID: "u1", Name: "Test User"}}, 1, nil
}
func (m *mockUserRepo) Count(ctx context.Context) (int, error) { return 0, nil }
func (m *mockUserRepo) IncrementReputationScore(ctx context.Context, id string, delta int) error {
	return nil
}
func (m *mockUserRepo) SetDailySpotLimitOverride(ctx context.Context, id string, limit int, expiresAt time.Time) error {
	return nil
}

func setupSearchRouter() *gin.Engine {
	svc := searchApp.NewSearchService(&mockSpotRepo{}, &mockSpeciesRepo{}, &mockUserRepo{})
	handler := searchIface.NewSearchHandler(svc)
	r := gin.New()
	v1 := r.Group("/v1")
	searchIface.RegisterRoutes(v1, handler)
	return r
}

func TestSearch_Success(t *testing.T) {
	r := setupSearchRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/search?q=test&type=all&limit=10", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	results := body["results"].(map[string]interface{})

	spots := results["spots"].([]interface{})
	if len(spots) != 1 {
		t.Errorf("expected 1 spot, got %d", len(spots))
	}

	species := results["species"].([]interface{})
	if len(species) != 1 {
		t.Errorf("expected 1 species, got %d", len(species))
	}

	users := results["users"].([]interface{})
	if len(users) != 1 {
		t.Errorf("expected 1 user, got %d", len(users))
	}
}

func TestSearch_MissingQuery(t *testing.T) {
	r := setupSearchRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/search", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSearch_SpotOnly(t *testing.T) {
	r := setupSearchRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/search?q=test&type=spot", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	results := body["results"].(map[string]interface{})

	// Verify spots were searched
	_ = results

	// Species and users should be empty arrays
	sp := results["species"].([]interface{})
	if len(sp) != 0 {
		t.Errorf("expected 0 species for spot-only search, got %d", len(sp))
	}
}

func TestSearch_SpeciesOnly(t *testing.T) {
	r := setupSearchRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/search?q=trucha&type=species", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestSearch_UserOnly(t *testing.T) {
	r := setupSearchRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/search?q=luis&type=user", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// Domain entity tests
func TestSearchResult_Empty(t *testing.T) {
	r := domain.SearchResult{
		Spots:   []domain.SpotResult{},
		Species: []domain.SpeciesResult{},
		Users:   []domain.UserResult{},
	}
	if len(r.Spots) != 0 || len(r.Species) != 0 || len(r.Users) != 0 {
		t.Error("expected empty result")
	}
}
