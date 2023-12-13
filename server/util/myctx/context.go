package myctx

import (
	"context"
	"log/slog"
)

type ctxString string

const logKey ctxString = "logger"

func LoggerFromCtx(ctx context.Context) *slog.Logger {
	logger := ctx.Value(logKey)
	switch l := logger.(type) {
	case *slog.Logger:
		return l
	default:
		panic("Logger not existing, giving up")
	}
}

func WithLogger(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, logKey, log)
}
