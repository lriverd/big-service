package logger_test

import (
	"testing"

	"github.com/lriverd/big-service/internal/platform/logger"
)

func TestSetup_Development(t *testing.T) {
	// Should not panic
	logger.Setup("debug", "development")
}

func TestSetup_Production(t *testing.T) {
	logger.Setup("info", "production")
}

func TestSetup_InvalidLevel(t *testing.T) {
	// Should fallback to info without panic
	logger.Setup("invalid", "development")
}

