package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lriverd/big-service/internal/platform/config"
	"github.com/lriverd/big-service/internal/platform/middleware"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupAuthRouter() (*gin.Engine, *middleware.AuthMiddleware) {
	cfg := &config.Config{JWTSecret: testSecret}
	authMw := middleware.NewAuthMiddleware(cfg)
	r := gin.New()
	return r, authMw
}

func TestRequireAuth_NoHeader(t *testing.T) {
	r, authMw := setupAuthRouter()
	r.GET("/test", authMw.RequireAuth(), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuth_InvalidFormat(t *testing.T) {
	r, authMw := setupAuthRouter()
	r.GET("/test", authMw.RequireAuth(), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	r, authMw := setupAuthRouter()
	r.GET("/test", authMw.RequireAuth(), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuth_ValidToken(t *testing.T) {
	r, authMw := setupAuthRouter()
	r.GET("/test", authMw.RequireAuth(), func(c *gin.Context) {
		userID, _ := c.Get("userID")
		role, _ := c.Get("userRole")
		c.JSON(200, gin.H{"userID": userID, "role": role})
	})

	token, _ := middleware.GenerateJWT("user123", "test@test.com", "admin", testSecret, 60)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["userID"] != "user123" {
		t.Errorf("expected userID user123, got %v", body["userID"])
	}
	if body["role"] != "admin" {
		t.Errorf("expected role admin, got %v", body["role"])
	}
}

func TestRequireAdmin_Allowed(t *testing.T) {
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		c.Set("userRole", "admin")
		c.Next()
	}, middleware.RequireAdmin(), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRequireAdmin_Forbidden(t *testing.T) {
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		c.Set("userRole", "user")
		c.Next()
	}, middleware.RequireAdmin(), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestRequireAdmin_NoRole(t *testing.T) {
	r := gin.New()
	r.GET("/test", middleware.RequireAdmin(), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

