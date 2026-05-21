package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lriverd/big-service/internal/platform/middleware"
)

func TestCORSMiddleware(t *testing.T) {
	r := gin.New()
	r.Use(middleware.CORSMiddleware([]string{"*"}))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCORSMiddleware_Options(t *testing.T) {
	r := gin.New()
	r.Use(middleware.CORSMiddleware([]string{"*"}))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Errorf("expected 200 or 204 for preflight, got %d", w.Code)
	}
}

