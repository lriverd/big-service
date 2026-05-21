package interfaces_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	authDomain "github.com/lriverd/big-service/internal/pescaapp/auth/domain"
	authIface "github.com/lriverd/big-service/internal/pescaapp/auth/interfaces"
	"github.com/lriverd/big-service/internal/platform/config"
	"github.com/lriverd/big-service/internal/platform/middleware"
)

func init() { gin.SetMode(gin.TestMode) }

type mockAuthService struct {
	loginFn  func(ctx context.Context, req authDomain.LoginRequest) (*authDomain.AuthResponse, error)
	logoutFn func(ctx context.Context, userID string) error
}

func (m *mockAuthService) Login(ctx context.Context, req authDomain.LoginRequest) (*authDomain.AuthResponse, error) {
	return m.loginFn(ctx, req)
}

func (m *mockAuthService) Logout(ctx context.Context, userID string) error {
	if m.logoutFn != nil {
		return m.logoutFn(ctx, userID)
	}
	return nil
}

func setupAuthRouter(svc authDomain.AuthService) *gin.Engine {
	cfg := &config.Config{JWTSecret: "test-secret"}
	authMw := middleware.NewAuthMiddleware(cfg)
	handler := authIface.NewAuthHandler(svc)
	r := gin.New()
	v1 := r.Group("/v1")
	authIface.RegisterRoutes(v1, handler, authMw)
	return r
}

func TestLogin_Success(t *testing.T) {
	svc := &mockAuthService{
		loginFn: func(ctx context.Context, req authDomain.LoginRequest) (*authDomain.AuthResponse, error) {
			return &authDomain.AuthResponse{
				User:         authDomain.AuthUser{ID: "u1", Name: "Test", Email: req.Email},
				AccessToken:  "token123",
				RefreshToken: "refresh123",
			}, nil
		},
	}

	r := setupAuthRouter(svc)
	body, _ := json.Marshal(map[string]string{"idToken": "google-token", "email": "test@test.com"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestLogin_BadRequest(t *testing.T) {
	svc := &mockAuthService{}
	r := setupAuthRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/auth/login", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLogin_AuthFailure(t *testing.T) {
	svc := &mockAuthService{
		loginFn: func(ctx context.Context, req authDomain.LoginRequest) (*authDomain.AuthResponse, error) {
			return nil, fmt.Errorf("invalid token")
		},
	}
	r := setupAuthRouter(svc)

	body, _ := json.Marshal(map[string]string{"idToken": "bad", "email": "test@test.com"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestLogout_Success(t *testing.T) {
	svc := &mockAuthService{}
	r := setupAuthRouter(svc)

	token, _ := middleware.GenerateJWT("user1", "test@test.com", "user", "test-secret", 60)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestLogout_NoAuth(t *testing.T) {
	svc := &mockAuthService{}
	r := setupAuthRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/auth/logout", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

