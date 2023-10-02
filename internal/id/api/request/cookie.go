package request

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/zhuboris/never-expires/internal/id/authservice"
	"github.com/zhuboris/never-expires/internal/id/tkn"
	"github.com/zhuboris/never-expires/internal/shared/httpmux"
)

func setAuthCookie(ctx context.Context, r *http.Request, authData authservice.AuthData, authService AuthService, w http.ResponseWriter) error {
	domain := "." + httpmux.RemoveSubdomain(r.Host)
	authTokenCookie := jwtCookie(authData.AccessToken(), tkn.AuthorizationCookie, domain, authservice.AuthLifetime)
	refreshTokenCookie := jwtCookie(authData.RefreshToken(), tkn.RefreshCookie, domain, authservice.RefreshLifetime)
	sessionIDCookie, err := sessionCookie(authData.Session().ID, domain)
	if err != nil {
		deactivationError := authService.DeactivateSession(ctx, authData.Session().ID)
		return errors.Join(deactivationError, err)
	}

	http.SetCookie(w, authTokenCookie)
	http.SetCookie(w, refreshTokenCookie)
	http.SetCookie(w, sessionIDCookie)
	return nil
}

func jwtCookie(token, name, domain string, expires time.Duration) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    token,
		Expires:  time.Now().Add(expires),
		HttpOnly: true,
		Secure:   true,
		Domain:   domain,
	}
}

func deleteAllAuthCookies(domain string, w http.ResponseWriter) {
	http.SetCookie(w, deletingCookie(tkn.AuthorizationCookie, domain))
	http.SetCookie(w, deletingCookie(tkn.RefreshCookie, domain))
	http.SetCookie(w, deletingCookie(sessionIDCookie, domain))
}

func deletingCookie(name, domain string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		Domain:   domain,
	}
}
