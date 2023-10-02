package usr

import (
	"net/mail"
	"strings"
)

func parseAddress(address string) (*mail.Address, error) {
	result, err := mail.ParseAddress(address)
	if err != nil {
		return result, err
	}

	if result.Name != "" {
		return result, nil
	}

	name, err := parseName(result.Address)
	result.Name = name

	return result, err
}

func parseName(email string) (string, error) {
	const atSing = "@"
	const notExist = -1

	atIndex := strings.Index(email, atSing)
	if atIndex == notExist {
		return "", ErrInvalidEmail
	}

	return email[:atIndex], nil
}
