package log

import (
	"context"

	"go.uber.org/zap"
)

var BaseLogger *zap.Logger

type ctxkey string

var (
	keyLogger ctxkey = "github.com/taxfyle/go-httpkit/log:logger"
)

func WithContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, keyLogger, logger)
}

func FromContext(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(keyLogger).(*zap.Logger)
	if !ok {
		return BaseLogger.With() // return a copy of the base logger, not the original
	}

	return logger
}
