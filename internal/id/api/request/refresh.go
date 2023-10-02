package request

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zhuboris/never-expires/internal/id/api/response"
	"github.com/zhuboris/never-expires/internal/id/authservice"
	"github.com/zhuboris/never-expires/internal/id/session"
	"github.com/zhuboris/never-expires/internal/id/tkn"
	"github.com/zhuboris/never-expires/internal/shared/httpmux"
	"github.com/zhuboris/never-expires/internal/shared/reqbody"
)

type RefreshRequest struct {
	authService AuthService
}

func NewRefreshRequest(authService AuthService) *RefreshRequest {
	return &RefreshRequest{
		authService: authService,
	}
}

func (req RefreshRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	input := new(authservice.RefreshJWTData)
	if err := reqbody.Decode(input, r.Body); err != nil {
		return err
	}

	domain := "." + httpmux.RemoveSubdomain(r.Host)
	userID, err := req.validateRefreshToken(input, r)
	if err != nil {
		tryDeleteCookies(err, domain, w)
		return err
	}

	token, err := tkn.CreateJWT(userID, authservice.AuthLifetime)
	if err != nil {
		tryDeleteCookies(err, domain, w)

		return err
	}

	authTokenCookie := jwtCookie(token, tkn.AuthorizationCookie, domain, authservice.AuthLifetime)
	http.SetCookie(w, authTokenCookie)
	return response.WriteJSONData(w, http.StatusOK, req.responseBody(token))
}

func (req RefreshRequest) validateRefreshToken(body *authservice.RefreshJWTData, r *http.Request) (pgtype.UUID, error) {
	validRefreshToken, userID, err := tkn.VerifyRefreshJWT(body.RefreshToken, r)
	if err != nil {
		return userID, err
	}

	if err := req.checkIsSessionValid(body.SessionID, userID, validRefreshToken, r); err != nil {
		return userID, err
	}

	return userID, nil
}

func (req RefreshRequest) checkIsSessionValid(idFromBody string, userID pgtype.UUID, token string, r *http.Request) error {
	sessionID, err := sessionID(idFromBody, r)
	if err != nil {
		return errors.Join(tkn.ErrUnauthorized, err)
	}

	currentSession := session.Session{
		ID:         sessionID,
		UserID:     userID,
		RefreshJWT: token,
	}

	return req.authService.AllowRefreshingJWT(r.Context(), currentSession)
}

func (req RefreshRequest) responseBody(token string) any {
	return struct {
		AccessToken string `json:"access_token"`
	}{token}
}

func tryDeleteCookies(err error, domain string, w http.ResponseWriter) {
	if errors.Is(err, tkn.ErrUnauthorized) {
		deleteAllAuthCookies(domain, w)
	}
}
