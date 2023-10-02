package usr

import (
	"fmt"
)

type InvalidTokenError struct {
	reason string
}

func (e InvalidTokenError) Error() string {
	return fmt.Sprintf("invalid token: %q", e.reason)
}

type InvalidTemplateError struct {
	path string
}

func (e InvalidTemplateError) Error() string {
	return fmt.Sprintf("failed to read or execute template by path %q", e.path)
}

type ExtractingSubjectError struct {
	reason string
}

func (e ExtractingSubjectError) Error() string {
	return fmt.Sprintf("failed to extract subject from HTML: %q", e.reason)
}

var (
	ErrEmailAlreadyConfirmed = InvalidTokenError{"email is already confirmed"}
	errTokenExpired          = InvalidTokenError{"expired"}
	errTokenNotExists        = InvalidTokenError{"not exists"}
	errTokenAlreadyUsed      = InvalidTokenError{"it has already been used"}
)
