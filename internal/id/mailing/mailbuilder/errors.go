package mailbuilder

import (
	"errors"
	"fmt"
)

type InvalidTemplateError struct {
	path string
}

func (e InvalidTemplateError) Error() string {
	return fmt.Sprintf("failed to read or execute template by path %q", e.path)
}

var (
	ErrTranslationsFileNotFound = errors.New("cannot open files with translations")
	ErrMissingDefaultLocale     = errors.New("locales dictionary missing default key")
	ErrFailedExecuteTemplate    = errors.New("failed to execute template, probably executing data is wrong")
)
