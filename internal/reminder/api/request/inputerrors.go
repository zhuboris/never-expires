package request

import (
	"errors"
	"fmt"

	"github.com/zhuboris/never-expires/internal/shared/postgresql"
)

type InputError struct {
	msg string
}

func (e InputError) Error() string {
	return fmt.Sprintf("invalid request input: %q", e.msg)
}

var (
	ErrMissingParam         = InputError{"missing required parameter"}
	ErrInvalidUUIDInBody    = InputError{"invalid uuid in body"}
	ErrInvalidTimeFormat    = InputError{"invalid time format"}
	ErrInvalidQuery         = InputError{"invalid query data"}
	ErrOptionNotExists      = InputError{"option is not exist"}
	ErrInvalidBody          = InputError{"invalid body"}
	ErrMissingRequiredField = InputError{"body is missing at least one required field"}
	ErrNewUUIDNotUnique     = InputError{"given uuid is not unique"}
)

func InvalidTimeFormatError(inputtedTime string) error {
	const formatExample = "2023-07-11T15:04:05Z"
	return fmt.Errorf("%w: got %q, expected in ISO 8601 format %q", ErrInvalidTimeFormat, inputtedTime, formatExample)
}

func warpAddingError(err error) error {
	if errors.Is(err, postgresql.ErrAddedDuplicateOfUnique) {
		err = errors.Join(ErrNewUUIDNotUnique, err)
	}

	return err
}
