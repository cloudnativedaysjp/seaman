package log

import (
	"context"
	"os"

	"golang.org/x/exp/slog"
)

type contextKey struct{}

func IntoContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

func FromContext(ctx context.Context, keysAndValues ...any) *slog.Logger {
	logger, ok := ctx.Value(contextKey{}).(*slog.Logger)
	if !ok {
		// if failed to get from context, get logger with default setting
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	}
	return logger
}
