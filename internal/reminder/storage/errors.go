package storage

import (
	"errors"
)

var (
	ErrStorageNameNotUnique = errors.New("cannot add storage with unique name")
	ErrDeletingNotAllowed   = errors.New("default storage is forbidden to delete")
)
