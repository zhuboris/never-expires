package httpmux

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type loggingFunc func(logger *zap.Logger, elapsedTime time.Duration)

func logMiddlewareResult(ctx context.Context, middlewareName string, startTime time.Time, logFunc loggingFunc) {
	timeLeft := time.Since(startTime)

	logger, err := NamedLogger(ctx, middlewareName)
	if err != nil {
		return
	}

	logFunc(logger, timeLeft)
}
