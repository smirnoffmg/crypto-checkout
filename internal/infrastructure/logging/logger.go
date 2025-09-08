package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger provides a structured logger for the application.
type Logger struct {
	*zap.Logger
}

// NewLogger creates a new logger instance.
func NewLogger() *Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()
	if err != nil {
		// Fallback to development config if production config fails
		logger, _ = zap.NewDevelopment()
	}

	return &Logger{Logger: logger}
}

// NewDevelopmentLogger creates a logger for development.
func NewDevelopmentLogger() *Logger {
	logger, _ := zap.NewDevelopment()
	return &Logger{Logger: logger}
}
