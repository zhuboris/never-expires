package httpmux

import (
	"context"
	"errors"

	"go.uber.org/zap"
)

const (
	serviceLogKey         = "service"
	requestIDLogKey       = "requestID"
	requestedMethodLogKey = "requestedMethod"
	allowedMethodsLogKey  = "allowedMethods"
	pathLogKey            = "path"
	responseLogKey        = "responseBody"
	elapsedTimeLogKey     = "elapsedTimeMS"
	statusCodeLogKey      = "statusCode"
	userIDLogKey          = "userID"
	rawIDLogKey           = "rawID"
)

type loggerKeyType struct{}

var loggerKey loggerKeyType

var errMissingValidLogger = errors.New("cannot get valid logger from context")

func NamedLogger(ctx context.Context, serviceName string) (*zap.Logger, error) {
	logger, ok := ctx.Value(loggerKey).(*zap.Logger)
	if ok && (logger != nil) {
		return addServiceName(logger, serviceName)
	}

	return logger, errMissingValidLogger
}

func addServiceName(logger *zap.Logger, name string) (*zap.Logger, error) {
	if err := checkName(logger, name); err != nil {
		return logger, err
	}

	return logger.With(zap.String(serviceLogKey, name)), nil
}

func checkName(logger *zap.Logger, name string) error {
	if name == "" {
		err := errors.New("service name is not provided")
		logger.Error("Cannot init logger", zap.Error(err))

		return err
	}

	return nil
}
