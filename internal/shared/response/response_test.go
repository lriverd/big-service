package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lriverd/big-service/internal/shared/response"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.Success(c, gin.H{"key": "value"})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["key"] != "value" {
		t.Errorf("expected key=value, got %v", body)
	}
}

func TestCreated(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.Created(c, gin.H{"id": "123"})

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestNoContent(t *testing.T) {
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		response.NoContent(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestPaginated(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := []string{"a", "b"}
	response.Paginated(c, data, 100, 20, 0)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)

	pag := body["pagination"].(map[string]interface{})
	if pag["total"].(float64) != 100 {
		t.Errorf("expected total 100")
	}
	if pag["hasMore"].(bool) != true {
		t.Errorf("expected hasMore true")
	}
}

func TestPaginated_NoMore(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.Paginated(c, []string{}, 5, 20, 0)

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	pag := body["pagination"].(map[string]interface{})
	if pag["hasMore"].(bool) != false {
		t.Errorf("expected hasMore false")
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "test error")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	errObj := body["error"].(map[string]interface{})
	if errObj["code"] != "BAD_REQUEST" {
		t.Errorf("expected BAD_REQUEST code")
	}
	if body["timestamp"] == nil {
		t.Error("expected timestamp")
	}
}

func TestErrorWithDetails(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.ErrorWithDetails(c, 400, "BAD_REQUEST", "msg", "details here")

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	errObj := body["error"].(map[string]interface{})
	if errObj["details"] != "details here" {
		t.Errorf("expected details")
	}
}



