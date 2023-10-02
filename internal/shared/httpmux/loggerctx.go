package httpmux

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

func withLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func withLoggingUserId(ctx context.Context, id pgtype.UUID) (context.Context, error) {
	logger, ok := ctx.Value(loggerKey).(*zap.Logger)
	if !ok || logger == nil {
		return ctx, errMissingValidLogger
	}

	idBytes, err := id.MarshalJSON()
	if err != nil {
		logger.Error("marshalling error or added user ID is invalid",
			zap.Any(rawIDLogKey, id),
			zap.Error(err),
		)
		return ctx, err
	}

	newLogger := logger.With(zap.String(userIDLogKey, string(idBytes)))
	return withLogger(ctx, newLogger), nil
}
