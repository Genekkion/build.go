package slog

import (
	"context"
	"log/slog"
)

type ctxKeyLogger struct{}

func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKeyLogger{}, l)
}

func LoggerFromCtx(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKeyLogger{}).(*slog.Logger); ok {
		return l
	}
	return defaultLogger
}
