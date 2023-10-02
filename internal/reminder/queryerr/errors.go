package queryerr

import "errors"

var ErrStorageNotExists = errors.New("user do not own storage with given id")
