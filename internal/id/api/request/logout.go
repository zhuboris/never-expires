package request

import (
	"errors"
	"net/http"

	"github.com/zhuboris/never-expires/internal/id/api/response"
	"github.com/zhuboris/never-expires/internal/id/authservice"
	"github.com/zhuboris/never-expires/internal/shared/httpmux"
)

type LogoutRequest struct {
	authService AuthService
}

var ErrMissingSessionOnLogout = errors.New("cookie do not contain session id on logout")

func NewLogoutRequest(authService AuthService) *LogoutRequest {
	return &LogoutRequest{
		authService: authService,
	}
}

func (req LogoutRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	domain := "." + httpmux.RemoveSubdomain(r.Host)
	deleteAllAuthCookies(domain, w)
	response.WriteMessage(w, http.StatusOK, "logout")

	return req.tryDeactivateSession(r)
}

func (req LogoutRequest) tryDeactivateSession(r *http.Request) error {
	sessionID, err := sessionID("", r)
	if err != nil {
		return errors.Join(ErrMissingSessionOnLogout, err)
	}

	var (
		ctx     = r.Context()
		data    = authservice.NewLogoutData(sessionID)
		handler = func() error {
			return req.authService.Logout(ctx, data)
		}
	)

	return httpmux.HandleErrorFuncWithTimeout(ctx, handler)
}
