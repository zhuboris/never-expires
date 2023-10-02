package oauth

import "errors"

var (
	ErrInvalidToken      = errors.New("idToken is invalid")
	errWrongTypeCasting  = errors.New("claims fields wrong type casting")
	errNoUserProfileData = errors.New("missing required user profile fields")
)
