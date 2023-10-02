package pw

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 10

var errBcrypting = errors.New("unexpected bcrypting error")

func Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", errors.Join(errBcrypting, err)
	}

	return string(bytes), nil
}

func Check(input, storedHash string) error {
	if input == "" && storedHash == "" {
		return nil
	}

	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(input))

	switch {
	case err == nil:
		return nil
	case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
		return ErrWrongPassword
	default:
		return errors.Join(errBcrypting, err)
	}
}
