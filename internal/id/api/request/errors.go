package request

import (
	"errors"
)

var (
	ErrInvalidBody          = errors.New("invalid request body")
	ErrMissingRequiredField = errors.New("body is missing at least one required field")
)
