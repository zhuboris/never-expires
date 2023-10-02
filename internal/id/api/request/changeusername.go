package request

import (
	"errors"
	"net/http"

	"github.com/zhuboris/never-expires/internal/id/api/response"
	"github.com/zhuboris/never-expires/internal/id/authservice"
	"github.com/zhuboris/never-expires/internal/shared/httpmux"
	"github.com/zhuboris/never-expires/internal/shared/reqbody"
)

type ChangeUsernameRequest struct {
	authService AuthService
}

func NewChangeUsernameRequest(authService AuthService) *ChangeUsernameRequest {
	return &ChangeUsernameRequest{
		authService: authService,
	}
}

func (req ChangeUsernameRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	input := new(authservice.ChangeUsernameData)
	if err := reqbody.Decode(input, r.Body); err != nil {
		return errors.Join(ErrInvalidBody, err)
	}

	if input.IsMissingRequiredField() {
		return ErrMissingRequiredField
	}

	var (
		ctx     = r.Context()
		handler = func() error {
			return req.authService.ChangeUsername(ctx, *input)
		}
	)

	if err := httpmux.HandleErrorFuncWithTimeout(ctx, handler); err != nil {
		return err
	}

	response.WriteMessage(w, http.StatusOK, "username changed")
	return nil
}
