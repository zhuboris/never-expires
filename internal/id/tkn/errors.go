package tkn

import (
	"errors"
	"fmt"
)

var (
	ErrUnauthorized  = errors.New("access denied")
	ErrForbidden     = errors.New("access forbidden")
	errEnvMissingKey = fmt.Errorf("key %q not exists in env", secretEnvKey)
)
