package logger

import (
	"context"

	"go.uber.org/zap"
)

type loggerKey struct{}

func InjectInContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

func New(ctx context.Context) *zap.Logger {
	val := ctx.Value(loggerKey{})
	if logger, ok := val.(*zap.Logger); ok && logger != nil {
		return logger
	}

	return zap.NewNop()
}
