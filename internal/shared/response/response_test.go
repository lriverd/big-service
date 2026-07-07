package response_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

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

func TestInternalError_LogsRealErrorAndRespondsGenerically(t *testing.T) {
	var logBuf bytes.Buffer
	origOut := log.StandardLogger().Out
	log.SetOutput(&logBuf)
	defer log.SetOutput(origOut)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/spots", nil)
	c.Set("requestId", "req-test-123")

	underlying := errors.New("firestore: deadline exceeded")
	response.InternalError(c, underlying, "Failed to create spot")

	// The client only ever sees the generic envelope, never the raw error.
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	errObj := body["error"].(map[string]interface{})
	if errObj["code"] != "INTERNAL_ERROR" {
		t.Errorf("expected INTERNAL_ERROR code, got %v", errObj["code"])
	}
	if strings.Contains(w.Body.String(), "deadline exceeded") {
		t.Error("the real error must not leak into the HTTP response")
	}

	// But the log must have everything needed to actually debug it.
	logged := logBuf.String()
	if !strings.Contains(logged, "deadline exceeded") {
		t.Errorf("expected the real error in the log, got: %s", logged)
	}
	if !strings.Contains(logged, "req-test-123") {
		t.Errorf("expected the request ID in the log for correlation, got: %s", logged)
	}
	if !strings.Contains(logged, "/v1/spots") {
		t.Errorf("expected the request path in the log, got: %s", logged)
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
