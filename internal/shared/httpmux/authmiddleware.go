package httpmux

import (
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/zhuboris/never-expires/internal/id/tkn"
	"github.com/zhuboris/never-expires/internal/id/usr"
	"github.com/zhuboris/never-expires/internal/shared/str"
)

const authServiceName = "authMiddleware"

var ErrUnauthorized = errors.New("access denied")

func authMiddleware(next errorHandledFunc) errorHandledFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var err error

		defer func(startTime time.Time) {
			loggingFunc := func(logger *zap.Logger, elapsedTime time.Duration) {
				logAuthAttempt(logger, err, elapsedTime)
			}

			logMiddlewareResult(r.Context(), authServiceName, startTime, loggingFunc)
		}(time.Now())

		id, err := tkn.VerifyUserJWT(r)
		if err != nil {
			return ErrUnauthorized
		}

		r = addIDToContext(r, id)
		if ctx, err := withLoggingUserId(r.Context(), id); err != nil {
			r = r.WithContext(ctx)
		}

		return next(w, r)
	}
}

func addIDToContext(r *http.Request, userID pgtype.UUID) *http.Request {
	ctx := usr.WithUserID(r.Context(), userID)
	return r.WithContext(ctx)
}

func logAuthAttempt(logger *zap.Logger, err error, elapsedTime time.Duration) {
	msg, logLvl := handleAuthErrorForLog(err)

	logger.Log(logLvl, str.Capitalized(msg),
		zap.Duration(elapsedTimeLogKey, elapsedTime),
		zap.Error(err),
	)
}

func handleAuthErrorForLog(err error) (string, zapcore.Level) {
	if err != nil {
		return "Access denied: failed extracting cookie from request or token invalid", zapcore.ErrorLevel
	}

	return "Access granted", zapcore.InfoLevel
}
