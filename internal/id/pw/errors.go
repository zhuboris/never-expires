package pw

import (
	"errors"
	"fmt"
)

type InsecurePasswordError struct {
	reason string
}

func (e InsecurePasswordError) Error() string {
	return fmt.Sprint("password is not strong enough: ", e.reason)
}

func (e InsecurePasswordError) Reason() string {
	return e.reason
}

var ErrWrongPassword = errors.New("wrong password")

var (
	errTooShort               = InsecurePasswordError{"the password must contain at least 8 symbols"}
	errContainsNotAllowedChar = InsecurePasswordError{"the password can only contain the following characters: 'A-Z', 'a-z', '0-9', '~`!@#$%^&*()_-+={[}]|\\:;\"'<,>.?/'"}
	errNoUppers               = InsecurePasswordError{"the password must contain uppercase letter"}
	errNoLowers               = InsecurePasswordError{"the password must contain lowercase letter"}
	errNoNumbers              = InsecurePasswordError{"the password must contain numbers"}
)
