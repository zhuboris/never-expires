package request

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zhuboris/never-expires/internal/id/tkn"
	"github.com/zhuboris/never-expires/internal/shared/reqbody"
	"github.com/zhuboris/never-expires/internal/shared/uuidformat"
)

type sessionIDData struct {
	SessionID string `json:"session_id"`
}

const (
	sessionIDCookie = "session-id"
)

var errMissingSessionID = errors.New("request does not contain session id")

func sessionCookie(id pgtype.UUID, domain string) (cookie *http.Cookie, err error) {
	idString := uuid.UUID(id.Bytes).String()

	return &http.Cookie{
		Name:     sessionIDCookie,
		Value:    idString,
		Expires:  time.Time{},
		HttpOnly: true,
		Secure:   true,
		Domain:   domain,
	}, nil
}

func sessionID(inputtedID string, r *http.Request) (pgtype.UUID, error) {
	idRaw, err := rawSessionID(inputtedID, r)
	if err != nil {
		return pgtype.UUID{}, err
	}

	if idRaw == "" {
		return pgtype.UUID{}, errMissingSessionID
	}

	return uuidformat.StrToPgtype(idRaw)
}

func rawSessionID(inputtedID string, r *http.Request) (string, error) {
	if inputtedID != "" {
		return inputtedID, nil
	}

	token, readingBodyErr := sessionIDFromBody(r)
	if readingBodyErr == nil { // if NO error
		return token, nil
	}

	token, readingCookieErr := sessionIDFromCookie(r)
	if readingCookieErr != nil {
		return "", errors.Join(readingBodyErr, readingCookieErr)
	}

	return token, nil
}

func sessionIDFromBody(r *http.Request) (string, error) {
	fromBody := new(sessionIDData)
	if err := reqbody.Decode(fromBody, r.Body); err != nil {
		return "", err
	}

	if fromBody.SessionID == "" {
		return "", errMissingSessionID
	}

	return fromBody.SessionID, nil
}

func sessionIDFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(sessionIDCookie)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", tkn.ErrUnauthorized
		}

		return "", err
	}
	return cookie.Value, nil
}
