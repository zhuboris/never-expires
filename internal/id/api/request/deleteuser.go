package request

import (
	"net/http"

	"github.com/zhuboris/never-expires/internal/shared/httpmux"
)

type DeleteUserRequest struct {
	authService AuthService
}

func NewDeleteUserRequest(authService AuthService) *DeleteUserRequest {
	return &DeleteUserRequest{
		authService: authService,
	}
}

func (req DeleteUserRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	var (
		ctx     = r.Context()
		handler = func() error {
			return req.authService.DeleteUser(ctx)
		}
	)

	if err := httpmux.HandleErrorFuncWithTimeout(ctx, handler); err != nil {
		return err
	}

	domain := "." + httpmux.RemoveSubdomain(r.Host)
	deleteAllAuthCookies(domain, w)

	w.WriteHeader(http.StatusNoContent)
	return nil
}
