// Package logger provides a production-structured zap logger.
//
// Convention: never pass secrets (API keys, tokens, passwords, signed URLs,
// cookies, auth headers) as log fields. Use request IDs for correlation.
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New builds a zap logger. In "debug" mode it uses a human-readable console
// encoder at DebugLevel; otherwise a JSON encoder at InfoLevel.
func New(mode string) (*zap.Logger, error) {
	if mode == "debug" || mode == "test" {
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		return cfg.Build(zap.AddStacktrace(zapcore.ErrorLevel))
	}
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return cfg.Build(zap.AddStacktrace(zapcore.ErrorLevel))
}
