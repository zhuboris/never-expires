package httpmux

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func loggerMiddleware(logger *zap.Logger, next errorHandledFunc) errorHandledFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		if logger == nil {
			return errors.New("expected valid logger but it is nil")
		}

		var (
			scopedLogger *zap.Logger
			err          error
		)

		defer func(startTime time.Time) {
			elapsedTime := time.Since(startTime)

			scopedLogger.Info("Logger was added to request", zap.Duration(elapsedTimeLogKey, elapsedTime), zap.Error(err))
		}(time.Now())

		ctx := r.Context()
		id, err := requestID(ctx)
		if err != nil {
			return fmt.Errorf("failed adding the logger to the context: %w", err)
		}

		scopedLogger = logger.With(zap.Stringer(requestIDLogKey, id), zap.String(pathLogKey, r.URL.Path))
		ctx = withLogger(ctx, scopedLogger)
		r = r.WithContext(ctx)

		return next(w, r)
	}
}
