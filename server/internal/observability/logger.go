package observability

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a new structured logger with JSON output format.
// The logger uses zap as the backend and is configured for production use.
func NewLogger() logr.Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)

	zapLogger, err := config.Build()
	if err != nil {
		// Fallback to a basic logger if config fails
		zapLogger, _ = zap.NewProduction()
	}

	return zapr.NewLogger(zapLogger)
}

// NewLoggerFromZap creates a logr.Logger from an existing zap.Logger.
// This is useful for testing with custom zap configurations.
func NewLoggerFromZap(zapLogger *zap.Logger) logr.Logger {
	return zapr.NewLogger(zapLogger)
}

