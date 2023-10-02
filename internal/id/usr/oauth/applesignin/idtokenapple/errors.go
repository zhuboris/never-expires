package idtokenapple

import "errors"

var (
	ErrKeyNotFound          = errors.New("key for given kid not found")
	ErrInvalidRequiredField = errors.New("invalid required field")
	ErrWrongIssuer          = errors.New("expected other issuer")
	ErrWrongAudience        = errors.New("expected other audience")
	ErrExpired              = errors.New("token already expired")
	ErrWrongKey             = errors.New("wrong signing key")
)
