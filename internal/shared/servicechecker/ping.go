package servicechecker

import (
	"context"
	"errors"
)

type (
	StatusDisplay interface {
		Set(error)
	}
	pinger interface {
		Ping(ctx context.Context) error
	}
)

func Ping(ctx context.Context, service pinger, statusDisplay StatusDisplay, serviceName string) error {
	err := service.Ping(ctx)
	if err != nil {
		err = errors.Join(IsUnavailableError{serviceName}, err)
	}

	statusDisplay.Set(err)
	return err
}
