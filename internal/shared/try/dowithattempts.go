package try

import (
	"context"
	"errors"
	"time"
)

func DoWithAttempts(cancelCtx context.Context, f func() error, delay time.Duration) error {
	suitableFunc := func() (struct{}, error) {
		return struct{}{}, f()
	}

	_, err := GetWithAttempts(cancelCtx, suitableFunc, delay)
	return err
}

func GetWithAttempts[T any](cancelCtx context.Context, f func() (T, error), delay time.Duration) (T, error) {
	result, err := f()
	if err == nil { // if No error
		return result, err
	}

	ticker := time.NewTicker(delay)
	defer ticker.Stop()

attemptingLoop:
	for {
		select {
		case <-ticker.C:
			if result, err = f(); err == nil { // if NO error
				break attemptingLoop
			}
		case <-cancelCtx.Done():
			err = errors.Join(cancelCtx.Err(), err)
			break attemptingLoop
		}
	}

	return result, err
}
