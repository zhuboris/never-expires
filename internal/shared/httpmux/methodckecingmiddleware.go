package httpmux

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/zhuboris/never-expires/internal/shared/str"
)

const methodCheckServiceName = "httpMethodCheckingMiddleware"

var validMethods = [...]string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodConnect,
	http.MethodOptions,
	http.MethodTrace,
}

var (
	ErrInvalidMethod    = errors.New("invalid HTTP method")
	ErrMethodNotAllowed = errors.New("method not allowed")
)

type allowedMethods []string

func (m allowedMethods) String() string {
	return strings.Join(m, ", ")
}

func (m allowedMethods) containsInvalid() error {
	for _, method := range m {
		if isInvalid(method) {
			return fmt.Errorf("allowed method %q is invalid", method)
		}
	}

	return nil
}

func (m allowedMethods) containsAskedMethod(askedMethod string) bool {
	for _, method := range m {
		if askedMethod == method {
			return true
		}
	}

	return false
}

func httpMethodCheckingMiddleware(allowed allowedMethods, next errorHandledFunc) errorHandledFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var err error

		defer func(startTime time.Time) {
			loggingFunc := func(logger *zap.Logger, elapsedTime time.Duration) {
				logMethodCheck(logger, err, allowed, r.Method, elapsedTime)
			}

			logMiddlewareResult(r.Context(), methodCheckServiceName, startTime, loggingFunc)
		}(time.Now())

		if len(allowed) == 0 {
			return errors.New("no http methods allowed")
		}

		if err = allowed.containsInvalid(); err != nil {
			return err
		}

		if isInvalid(r.Method) {
			return ErrInvalidMethod
		}

		if isAllowed := allowed.containsAskedMethod(r.Method); !isAllowed {
			w.Header().Add(allowHeader, allowed.String())
			return ErrMethodNotAllowed
		}

		return next(w, r)
	}
}

func isInvalid(method string) bool {
	for _, validMethod := range validMethods {
		if method == validMethod {
			return false
		}
	}

	return true
}

func logMethodCheck(logger *zap.Logger, err error, allowed allowedMethods, requestedMethod string, elapsedTime time.Duration) {
	var (
		msg    = "Success, method allowed"
		logLvl = zapcore.InfoLevel
	)

	if err != nil {
		logLvl = zapcore.ErrorLevel
		msg = err.Error()
	}

	logger.Log(logLvl, str.Capitalized(msg),
		zap.String(requestedMethodLogKey, requestedMethod),
		zap.String(allowedMethodsLogKey, allowed.String()),
		zap.Duration(elapsedTimeLogKey, elapsedTime),
	)
}
