package utils

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

func IntoContext(ctx context.Context, logger logr.Logger) context.Context {
	return logr.NewContext(ctx, logger)
}

func FromContext(ctx context.Context, keysAndValues ...interface{}) logr.Logger {
	logger, err := logr.FromContext(ctx)
	if err != nil {
		zaplogger, _ := zap.NewDevelopment()
		logger = zapr.NewLogger(zaplogger)
	}
	return logger.WithValues(keysAndValues...)
}
