package pagination_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lriverd/big-service/internal/shared/pagination"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestParse_Defaults(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)

	p := pagination.Parse(c)
	if p.Limit != 20 {
		t.Errorf("expected limit 20, got %d", p.Limit)
	}
	if p.Offset != 0 {
		t.Errorf("expected offset 0, got %d", p.Offset)
	}
}

func TestParse_CustomValues(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test?limit=50&offset=10", nil)

	p := pagination.Parse(c)
	if p.Limit != 50 {
		t.Errorf("expected limit 50, got %d", p.Limit)
	}
	if p.Offset != 10 {
		t.Errorf("expected offset 10, got %d", p.Offset)
	}
}

func TestParse_MaxLimit(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test?limit=200", nil)

	p := pagination.Parse(c)
	if p.Limit != 100 {
		t.Errorf("expected max limit 100, got %d", p.Limit)
	}
}

func TestParse_InvalidValues(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test?limit=abc&offset=xyz", nil)

	p := pagination.Parse(c)
	if p.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", p.Limit)
	}
	if p.Offset != 0 {
		t.Errorf("expected default offset 0, got %d", p.Offset)
	}
}

func TestParse_NegativeValues(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test?limit=-5&offset=-10", nil)

	p := pagination.Parse(c)
	if p.Limit != 1 {
		t.Errorf("expected min limit 1, got %d", p.Limit)
	}
	if p.Offset != 0 {
		t.Errorf("expected min offset 0, got %d", p.Offset)
	}
}

