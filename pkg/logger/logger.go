package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(env string) (*zap.Logger, error) {
	if env == "production" {
		cfg := zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		return cfg.Build()
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return cfg.Build()
}
