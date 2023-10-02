package httpmux

import (
	"context"
	"errors"
	"fmt"
)

type handlerFunc[T any] func() T

var (
	ErrTimeout  = errors.New("request timeout")
	ErrCanceled = errors.New("request canceled")
)

func HandleErrorFuncWithTimeout(ctx context.Context, errorHandlerFunc handlerFunc[error]) error {
	funcErr, ctxErr := HandleWithTimeout(ctx, errorHandlerFunc)
	return errors.Join(funcErr, ctxErr)
}

// HandleWithTimeout is processing handler func until it is finished or context is Done.
// Finishing successfully it returns panhandler's result, otherwise it returns handled
// context error.
//
// If your handler simply returns an error you should use HandleErrorFuncWithTimeout.
func HandleWithTimeout[T any](ctx context.Context, handler handlerFunc[T]) (funcResult T, ctxError error) {
	resultCh := make(chan T)
	go func() {
		resultCh <- handler()
	}()

	select {
	case <-ctx.Done():
		ctxError = handleContextError(ctx)
	case funcResult = <-resultCh:
		// empty because funcResult is simply returned to be handled
	}

	return funcResult, ctxError
}

func handleContextError(ctx context.Context) error {
	err := ctx.Err()

	switch {
	case errors.Is(err, context.Canceled):
		return errors.Join(ErrCanceled, err)
	case errors.Is(err, context.DeadlineExceeded):
		return errors.Join(ErrTimeout, err)
	default:
		return fmt.Errorf("unknown context error: %w", err)
	}
}
