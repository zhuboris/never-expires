package runapi

import (
	"context"
	"errors"
)

func WithContext(ctx context.Context, runFunc func() error, shutdownFunc func() error) error {
	var err error
	errChan := make(chan error)
	go func() {
		errChan <- runFunc()
	}()

loop:
	for {
		select {
		case <-ctx.Done():
			err = shutdownFunc()
			break loop
		case err = <-errChan:
			break loop
		default:
			continue
		}
	}

	err = errors.Join(err, ctx.Err())
	return err

}
