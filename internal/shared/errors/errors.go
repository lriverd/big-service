package errors

import "fmt"

type AppError struct {
	Code    string
	Message string
	Status  int
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func New(status int, code, message string) *AppError {
	return &AppError{Status: status, Code: code, Message: message}
}

var (
	ErrNotFound       = &AppError{Status: 404, Code: "NOT_FOUND", Message: "Resource not found"}
	ErrUnauthorized   = &AppError{Status: 401, Code: "UNAUTHORIZED", Message: "Unauthorized"}
	ErrForbidden      = &AppError{Status: 403, Code: "FORBIDDEN", Message: "Permission denied"}
	ErrBadRequest     = &AppError{Status: 400, Code: "BAD_REQUEST", Message: "Bad request"}
	ErrConflict       = &AppError{Status: 409, Code: "CONFLICT", Message: "Resource already exists"}
	ErrInternal       = &AppError{Status: 500, Code: "INTERNAL_ERROR", Message: "Internal server error"}
	ErrRateLimited    = &AppError{Status: 429, Code: "RATE_LIMIT_EXCEEDED", Message: "Too many requests"}
	ErrValidation     = &AppError{Status: 400, Code: "VALIDATION_ERROR", Message: "Validation failed"}
)

func NotFound(resource string) *AppError {
	return &AppError{Status: 404, Code: "NOT_FOUND", Message: fmt.Sprintf("%s not found", resource)}
}

func BadRequest(message string) *AppError {
	return &AppError{Status: 400, Code: "BAD_REQUEST", Message: message}
}

func Forbidden(message string) *AppError {
	return &AppError{Status: 403, Code: "FORBIDDEN", Message: message}
}

func Unauthorized(message string) *AppError {
	return &AppError{Status: 401, Code: "UNAUTHORIZED", Message: message}
}

func Conflict(message string) *AppError {
	return &AppError{Status: 409, Code: "CONFLICT", Message: message}
}

