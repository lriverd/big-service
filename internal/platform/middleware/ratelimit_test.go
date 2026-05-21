package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lriverd/big-service/internal/platform/cache"
	"github.com/lriverd/big-service/internal/platform/middleware"
)

func TestRateLimiter_AllowsRequests(t *testing.T) {
	c := cache.New(time.Minute, time.Minute)
	rl := middleware.NewRateLimiter(c, 10)

	r := gin.New()
	r.Use(rl.Limit())
	r.GET("/test", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Header().Get("X-RateLimit-Limit") != "10" {
		t.Errorf("expected X-RateLimit-Limit=10, got %s", w.Header().Get("X-RateLimit-Limit"))
	}
}

func TestRateLimiter_BlocksAfterLimit(t *testing.T) {
	appCache := cache.New(time.Minute, time.Minute)
	rl := middleware.NewRateLimiter(appCache, 3)

	r := gin.New()
	r.Use(rl.Limit())
	r.GET("/test", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"ok": true})
	})

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Errorf("request %d: expected 200, got %d", i, w.Code)
		}
	}

	// 4th request should be blocked
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", w.Code)
	}
}

