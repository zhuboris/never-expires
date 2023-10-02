package httpmux

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type requestsCounter interface {
	Increment(route, method string, statusCode, internalErrorCode int)
}

type errorHandlingFunc func(error) RequestingResult

type Mux struct {
	zapLogger         *zap.Logger
	defaultTimeout    time.Duration
	errorHandlingFunc errorHandlingFunc
	requestCounter    requestsCounter

	http.ServeMux
}

func NewMux(errorHandlingFunc errorHandlingFunc) *Mux {
	return &Mux{
		errorHandlingFunc: errorHandlingFunc,
	}
}

func (m *Mux) SetLogger(zapLogger *zap.Logger) {
	m.zapLogger = zapLogger
}

func (m *Mux) SetDefaultTimeout(value time.Duration) {
	m.defaultTimeout = value
}

func (m *Mux) SetRequestCounter(counter requestsCounter) {
	m.requestCounter = counter
}

func (m *Mux) HandleGet(route string, f errorHandledFunc, middlewares ...Middleware) {
	m.HandleFuncWithMiddlewares(route, f, []string{http.MethodGet}, middlewares...)
}

func (m *Mux) HandleHead(route string, f errorHandledFunc, middlewares ...Middleware) {
	m.HandleFuncWithMiddlewares(route, f, []string{http.MethodHead}, middlewares...)
}

func (m *Mux) HandlePut(route string, f errorHandledFunc, middlewares ...Middleware) {
	m.HandleFuncWithMiddlewares(route, f, []string{http.MethodPut}, middlewares...)
}

func (m *Mux) HandlePatch(route string, f errorHandledFunc, middlewares ...Middleware) {
	m.HandleFuncWithMiddlewares(route, f, []string{http.MethodPatch}, middlewares...)
}

func (m *Mux) HandleDelete(route string, f errorHandledFunc, middlewares ...Middleware) {
	m.HandleFuncWithMiddlewares(route, f, []string{http.MethodDelete}, middlewares...)
}

func (m *Mux) HandleConnect(route string, f errorHandledFunc, middlewares ...Middleware) {
	m.HandleFuncWithMiddlewares(route, f, []string{http.MethodConnect}, middlewares...)
}
func (m *Mux) HandlePost(route string, f errorHandledFunc, middlewares ...Middleware) {
	m.HandleFuncWithMiddlewares(route, f, []string{http.MethodPost}, middlewares...)
}

func (m *Mux) HandleOptions(route string, f errorHandledFunc, middlewares ...Middleware) {
	m.HandleFuncWithMiddlewares(route, f, []string{http.MethodOptions}, middlewares...)
}

func (m *Mux) HandleTrace(route string, f errorHandledFunc, middlewares ...Middleware) {
	m.HandleFuncWithMiddlewares(route, f, []string{http.MethodTrace}, middlewares...)
}

func (m *Mux) HandleFuncWithMiddlewares(route string, f errorHandledFunc, httpMethods []string, middlewares ...Middleware) {
	middlewares = append(m.defaultMiddlewares(httpMethods), middlewares...)
	httpHandler := requestIDMiddleware(m.handleAPIErrors(applyMiddlewares(f, middlewares...)))
	httpHandler = m.metricsRegisterMiddleware(route, httpHandler)

	m.HandleFunc(route, httpHandler)
}

func (m *Mux) defaultMiddlewares(httpMethods []string) []Middleware {
	return []Middleware{m.tryAddDefaultTimeoutMiddleware(), makeScopedLogger(m.zapLogger), checkHttpMethod(httpMethods...)}
}

func (m *Mux) tryAddDefaultTimeoutMiddleware() Middleware {
	return func(f errorHandledFunc) errorHandledFunc {
		return timeoutMiddleware(m.defaultTimeout, f)
	}
}

func (m *Mux) handleAPIErrors(f errorHandledFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var handledResult RequestingResult

		defer func(startTime time.Time) {
			timeLeft := time.Since(startTime)
			if m.zapLogger == nil {
				return
			}

			logger := m.zapLogger
			if id, err := requestID(r.Context()); err == nil { // if NO error
				logger = logger.With(zap.Stringer(requestIDLogKey, id))
			}

			logResult(logger, handledResult, timeLeft)
		}(time.Now())

		err := f(w, r)
		handledResult = m.errorHandlingFunc(err)
		handledResult.AddToRespond(w)

		m.tryRegisterHandledRequest(handledResult, r)
	}
}

func (m *Mux) tryRegisterHandledRequest(result RequestingResult, r *http.Request) {
	if m.requestCounter == nil {
		return
	}

	ctx := r.Context()
	requestedEndpoint, ok := endpoint(ctx)
	if !ok {
		return
	}

	statusCode, errorCode := result.responseCodes()
	m.requestCounter.Increment(requestedEndpoint, r.Method, statusCode, errorCode)
}
