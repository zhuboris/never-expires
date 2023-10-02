package usr

import "errors"

var (
	ErrNotFound                   = errors.New("provided user not found")
	ErrInvalidEmail               = errors.New("invalid email format")
	ErrMissingConfirmToken        = errors.New("token for confirmation email is required")
	ErrValidationRefused          = errors.New("validation is refused, inputted email cannot be confirmed with provided token")
	ErrPasswordResetRefused       = errors.New("password reset is refused")
	ErrNotConfirmedOrChangedEmail = errors.New("email is not confirmed or was changed")
)
