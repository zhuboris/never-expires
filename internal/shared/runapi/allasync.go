package runapi

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type Runner interface {
	RunWithCtx(ctx context.Context) error
}

func AllAsync(ctx context.Context, cancel context.CancelFunc, runners map[string]Runner) error {
	var (
		count   = len(runners)
		errChan = make(chan error, count)
		err     error
	)

	for key := range runners {
		name := key

		go func(ctx context.Context) {
			err := runners[name].RunWithCtx(ctx)
			errChan <- shutdownError(name, err)
			cancel()
		}(ctx)
	}

	for i := 0; i < count; i++ {
		err = errors.Join(<-errChan, err)
	}

	close(errChan)
	return err
}

func RunnersList(runners map[string]Runner) string {
	names := make([]string, 0, len(runners))
	for name := range runners {
		names = append(names, name)
	}

	return strings.Join(names, ", ")
}

func shutdownError(name string, err error) error {
	return fmt.Errorf("%s is shutdown: %w", name, err)
}
