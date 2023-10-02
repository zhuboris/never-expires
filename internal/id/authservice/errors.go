package authservice

import "errors"

var (
	ErrWrongLoginData      = errors.New("email or password is incorrect")
	ErrAlreadyRegistered   = errors.New("email already registered")
	ErrMissingEmailAddress = errors.New("email address does not belong to any user")
	errMissingUserDevice   = errors.New("login data must contain any information about user's device")
)
