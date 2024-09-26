package log

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var BaseLogger *zap.SugaredLogger

type ctxkey string

var (
	keyLogger ctxkey = "github.com/taxfyle/go-httpkit/log"
)

type Logger struct {
	*zap.SugaredLogger

	ID string
}

func NewBaseLogger(ctx context.Context) (context.Context, *Logger) {
	logger := &Logger{
		ID:            "base-logger",
		SugaredLogger: BaseLogger.Desugar().With(zap.String("log.id", "base-logger")).Sugar(),
	}

	return context.WithValue(ctx, keyLogger, logger), logger
}

func NewContext(ctx context.Context, logger *Logger) (context.Context, *Logger) {
	if logger == nil {
		id := uuid.New().String()

		logger = &Logger{
			SugaredLogger: BaseLogger.Desugar().With(zap.String("log.id", id)).Sugar(),
			ID:            id,
		}
	}

	return context.WithValue(ctx, keyLogger, logger), logger
}

func FromContext(ctx context.Context) *Logger {
	logger, ok := ctx.Value(keyLogger).(*Logger)
	if !ok {
		return &Logger{
			SugaredLogger: BaseLogger.Desugar().With(zap.String("log.id", "UNSET")).Sugar(),
			ID:            "UNSET",
		}
	}

	return logger
}

func (l *Logger) WithField(key string, value interface{}) *Logger {
	l.SugaredLogger = l.SugaredLogger.With(key, value)

	return l
}

func (l *Logger) WithFields(args ...interface{}) *Logger {
	l.SugaredLogger = l.SugaredLogger.With(args...)

	return l
}
