package usr

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
)

type ConfirmationToken struct {
	Value          string    `json:"token"`
	ExpirationTime time.Time `json:"expiration_time"`
	IsUsed         bool      `json:"is_used"`
}

const tokenLength = 128

func newConfirmationToken(lifetime time.Duration) (ConfirmationToken, error) {
	generatedToken, err := makeToken(tokenLength)
	if err != nil {
		return ConfirmationToken{}, err
	}

	return ConfirmationToken{
		Value:          generatedToken,
		ExpirationTime: expiration(lifetime),
	}, nil
}

func (t ConfirmationToken) InvalidityReason() error {
	switch {
	case t.IsUsed:
		return errTokenAlreadyUsed
	case time.Now().After(t.ExpirationTime):
		return errTokenExpired
	default:
		return InvalidTokenError{}
	}
}

func makeToken(len int) (string, error) {
	b := make([]byte, len)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("unexpected error occurred while making random token: %w", err)
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

func expiration(lifetime time.Duration) time.Time {
	return time.Now().
		Add(lifetime).
		UTC()
}
