package errors_test

import (
	"testing"

	apperrors "github.com/lriverd/big-service/internal/shared/errors"
)

func TestAppError_Error(t *testing.T) {
	err := apperrors.New(400, "BAD_REQUEST", "test error")
	expected := "[BAD_REQUEST] test error"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestNotFound(t *testing.T) {
	err := apperrors.NotFound("User")
	if err.Status != 404 {
		t.Errorf("expected status 404, got %d", err.Status)
	}
	if err.Code != "NOT_FOUND" {
		t.Errorf("expected code NOT_FOUND, got %s", err.Code)
	}
	if err.Message != "User not found" {
		t.Errorf("expected 'User not found', got %q", err.Message)
	}
}

func TestBadRequest(t *testing.T) {
	err := apperrors.BadRequest("invalid input")
	if err.Status != 400 {
		t.Errorf("expected status 400, got %d", err.Status)
	}
}

func TestForbidden(t *testing.T) {
	err := apperrors.Forbidden("no access")
	if err.Status != 403 {
		t.Errorf("expected status 403, got %d", err.Status)
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		err    *apperrors.AppError
		status int
		code   string
	}{
		{apperrors.ErrNotFound, 404, "NOT_FOUND"},
		{apperrors.ErrUnauthorized, 401, "UNAUTHORIZED"},
		{apperrors.ErrForbidden, 403, "FORBIDDEN"},
		{apperrors.ErrBadRequest, 400, "BAD_REQUEST"},
		{apperrors.ErrConflict, 409, "CONFLICT"},
		{apperrors.ErrInternal, 500, "INTERNAL_ERROR"},
		{apperrors.ErrRateLimited, 429, "RATE_LIMIT_EXCEEDED"},
		{apperrors.ErrValidation, 400, "VALIDATION_ERROR"},
	}
	for _, tt := range tests {
		if tt.err.Status != tt.status {
			t.Errorf("expected status %d for %s, got %d", tt.status, tt.code, tt.err.Status)
		}
		if tt.err.Code != tt.code {
			t.Errorf("expected code %s, got %s", tt.code, tt.err.Code)
		}
	}
}

