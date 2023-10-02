package pw

import (
	"unicode"
)

const minValidLength = 8

func Validate(password string) error {
	var (
		containsUpper  = false
		containsLower  = false
		containsNumber = false
	)

	if len(password) < minValidLength {
		return errTooShort
	}

	for _, char := range password {
		switch {
		case isNotASCIIChar(char):
			return errContainsNotAllowedChar
		case unicode.IsUpper(char):
			containsUpper = true
		case unicode.IsLower(char):
			containsLower = true
		case unicode.IsNumber(char):
			containsNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			continue
		default:
			return errContainsNotAllowedChar
		}
	}

	switch {
	case !containsUpper:
		return errNoUppers
	case !containsLower:
		return errNoLowers
	case !containsNumber:
		return errNoNumbers
	default:
		return nil
	}
}

func isNotASCIIChar(c rune) bool {
	const (
		minASCIICode = 0
		maxASCIICode = 127
	)

	return c < minASCIICode || c > maxASCIICode
}
