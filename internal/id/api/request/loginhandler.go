package request

import (
	"context"
	"errors"
	"net/http"

	"github.com/zhuboris/never-expires/internal/id/api/response"
	"github.com/zhuboris/never-expires/internal/id/authservice"
	"github.com/zhuboris/never-expires/internal/id/usr"
	"github.com/zhuboris/never-expires/internal/shared/httpmux"
)

func handleLogin(ctx context.Context, handler func() authservice.LoginResult, authService AuthService, w http.ResponseWriter, r *http.Request) (authservice.LoginResult, error) {
	result, ctxError := httpmux.HandleWithTimeout(ctx, handler)
	if err := result.Error(); err != nil || ctxError != nil {
		return authservice.LoginResult{}, errors.Join(err, ctxError)
	}

	if err := deactivateSessionIfWasActive(authService, r); err != nil {
		return authservice.LoginResult{}, err
	}

	authData := result.AuthData()
	if err := setAuthCookie(ctx, r, authData, authService, w); err != nil {
		return authservice.LoginResult{}, err
	}

	user := result.UserData()
	body := loginResponseBody(user, authData.ToResponse())
	if err := response.WriteJSONData(w, http.StatusOK, body); err != nil {
		return authservice.LoginResult{}, err
	}

	return result, nil
}

func deactivateSessionIfWasActive(authService AuthService, r *http.Request) error {
	if id, err := sessionID("", r); err == nil { // if NO error
		return authService.DeactivateSession(r.Context(), id)
	}

	return nil
}

func loginResponseBody(user *usr.PublicData, authResponse authservice.AuthResponse) any {
	return struct {
		*usr.PublicData
		authservice.AuthResponse
	}{
		PublicData:   user,
		AuthResponse: authResponse,
	}
}
