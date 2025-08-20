package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New builds a zap.Logger based on textual level (debug, info, warn, error).
func New(level string) (*zap.Logger, error) {
	var cfg zap.Config
	if strings.ToLower(level) == "debug" {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}
	if level != "" {
		if err := cfg.Level.UnmarshalText([]byte(level)); err != nil {
			_ = err // ignore invalid and keep default
		}
	}
	return cfg.Build()
}


