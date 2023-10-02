package request

import (
	"errors"
	"net/http"

	"github.com/zhuboris/never-expires/internal/id/api/response"
	"github.com/zhuboris/never-expires/internal/id/authservice"
	"github.com/zhuboris/never-expires/internal/shared/httpmux"
)

type GetUserRequest struct {
	authService AuthService
}

func NewGetUserRequest(authService AuthService) *GetUserRequest {
	return &GetUserRequest{
		authService: authService,
	}
}

func (req GetUserRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	var (
		ctx     = r.Context()
		handler = func() authservice.GettingUserResult {
			return req.authService.AuthorizedUser(ctx)
		}
	)

	result, ctxError := httpmux.HandleWithTimeout(ctx, handler)
	if resultError := result.Error(); resultError != nil || ctxError != nil {
		return errors.Join(resultError, ctxError)
	}

	return response.WriteJSONData(w, http.StatusOK, result.UserData())
}
