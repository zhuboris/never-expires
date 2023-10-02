package tkn

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	AuthorizationCookie = "access-jwt"
	RefreshCookie       = "refresh-jwt"
)

func containsInvalidJWT(err error) bool {
	isMissingCookie := errors.Is(err, http.ErrNoCookie)
	return err != nil && !isMissingCookie
}

func belongToDifferentUsers(firstOwner, secondOwner pgtype.UUID, err error) bool {
	isMissingCookie := errors.Is(err, http.ErrNoCookie)
	return firstOwner != secondOwner && !isMissingCookie
}

func parseIfValid(signedToken string) (pgtype.UUID, error) {
	jwtSecretKey, ok := os.LookupEnv(secretEnvKey)
	if !ok {
		return pgtype.UUID{}, errEnvMissingKey
	}

	claims := new(UserClaims)
	token, err := jwt.ParseWithClaims(signedToken, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: wrong signing method", ErrUnauthorized)
		}

		return []byte(jwtSecretKey), nil
	})

	if err != nil || !token.Valid {
		return claims.UserID, fmt.Errorf("%w: invalid token: %s, err: %w", ErrUnauthorized, signedToken, err)
	}

	return claims.UserID, nil
}
