package httpmux

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	errorHandledFunc func(w http.ResponseWriter, r *http.Request) error
	Middleware       func(errorHandledFunc) errorHandledFunc
)

func applyMiddlewares(handler errorHandledFunc, middlewares ...Middleware) errorHandledFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler
}

func makeScopedLogger(logger *zap.Logger) Middleware {
	return func(f errorHandledFunc) errorHandledFunc {
		return loggerMiddleware(logger, f)
	}
}

func checkHttpMethod(allowedMethodsList ...string) Middleware {
	return func(f errorHandledFunc) errorHandledFunc {
		return httpMethodCheckingMiddleware(allowedMethodsList, f)
	}
}

func Authorize() Middleware {
	return func(f errorHandledFunc) errorHandledFunc {
		return authMiddleware(f)
	}
}

func SetTimeout(value time.Duration) Middleware {
	return func(f errorHandledFunc) errorHandledFunc {
		return timeoutMiddleware(value, f)
	}
}
