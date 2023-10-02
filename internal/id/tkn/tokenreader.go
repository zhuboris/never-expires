package tkn

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

func VerifyRefreshJWT(jwtFromBody string, r *http.Request) (token string, userID pgtype.UUID, err error) {
	token, err = readRefreshToken(jwtFromBody, r)
	if err != nil {
		return "", userID, errors.Join(ErrUnauthorized, err)
	}

	idFromRefresh, err := parseIfValid(token)
	if err != nil {
		return "", userID, err
	}

	idFromAuth, err := VerifyUserJWT(r)
	if containsInvalidJWT(err) {
		return "", userID, ErrUnauthorized
	}

	if belongToDifferentUsers(idFromRefresh, idFromAuth, err) {
		return "", userID, ErrForbidden
	}

	return token, idFromRefresh, nil
}

func VerifyUserJWT(r *http.Request) (pgtype.UUID, error) {
	token, err := readAccessToken(r)
	if err != nil {
		return pgtype.UUID{}, errors.Join(ErrUnauthorized, err)
	}

	return parseIfValid(token)
}

func readRefreshToken(jwtFromBody string, r *http.Request) (string, error) {
	if jwtFromBody != "" {
		return jwtFromBody, nil
	}

	return cookieValue(RefreshCookie, r)
}

func readAccessToken(r *http.Request) (string, error) {
	token, readingHeaderError := readAuthorizationHeader(r)
	if readingHeaderError == nil { // if NO error
		return token, nil
	}

	token, readingCookieError := cookieValue(AuthorizationCookie, r)
	if readingCookieError != nil {
		return "", errors.Join(readingHeaderError, readingCookieError)
	}
	return token, nil
}

func cookieValue(cookieName string, r *http.Request) (string, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", errors.Join(fmt.Errorf("cookies do not contain %q", cookieName), err)
		}

		return "", err
	}

	return cookie.Value, nil
}

func readAuthorizationHeader(r *http.Request) (string, error) {
	const bearer = "Bearer "

	value, err := authorizationHeaderValue(r)
	if err != nil {
		return "", err
	}

	signedToken := strings.TrimPrefix(value, bearer)
	return signedToken, nil
}

func authorizationHeaderValue(r *http.Request) (string, error) {
	const name = "Authorization"

	value := r.Header.Get(name)
	if value == "" {
		return "", fmt.Errorf("header %q does not contain value", name)
	}

	return value, nil
}
