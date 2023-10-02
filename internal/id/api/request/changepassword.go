package request

import (
	"context"
	"errors"
	"net/http"

	"github.com/zhuboris/never-expires/internal/id/api/request/device"
	"github.com/zhuboris/never-expires/internal/id/api/response"
	"github.com/zhuboris/never-expires/internal/id/authservice"
	"github.com/zhuboris/never-expires/internal/id/usr"
	"github.com/zhuboris/never-expires/internal/shared/httpmux"
	"github.com/zhuboris/never-expires/internal/shared/reqbody"
)

type (
	ChangePasswordRequest struct {
		authService AuthService
	}
	sessionCreationResult struct {
		authData authservice.AuthData
		err      error
	}
)

func NewChangePasswordRequest(authService AuthService) *ChangePasswordRequest {
	return &ChangePasswordRequest{
		authService: authService,
	}
}

func (req ChangePasswordRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	input := new(authservice.ChangePasswordData)
	if err := reqbody.Decode(input, r.Body); err != nil {
		return errors.Join(ErrInvalidBody, err)
	}

	if input.IsMissingRequiredField() {
		return ErrMissingRequiredField
	}

	var (
		ctx             = r.Context()
		changePwHandler = func() error {
			return req.authService.ChangePassword(ctx, *input)
		}
	)

	if err := httpmux.HandleErrorFuncWithTimeout(ctx, changePwHandler); err != nil {
		return err
	}

	domain := "." + httpmux.RemoveSubdomain(r.Host)
	deleteAllAuthCookies(domain, w)

	authData, err := req.createNewAuthData(ctx, r)
	if err != nil {
		return err
	}

	if err := setAuthCookie(ctx, r, authData, req.authService, w); err != nil {
		return err
	}

	respondingError := response.WriteJSONData(w, http.StatusOK, authData.ToResponse())
	sendingError := req.notifyAboutChangeWithEmail(ctx, r)

	return errors.Join(respondingError, sendingError)
}

func (req ChangePasswordRequest) createNewAuthData(ctx context.Context, r *http.Request) (authservice.AuthData, error) {
	userID, err := usr.ID(ctx)
	if err != nil {
		return authservice.AuthData{}, err
	}

	createSessionHandler := func() sessionCreationResult {
		data, err := req.authService.CreateSession(ctx, userID, device.Info(r))
		return sessionCreationResult{
			authData: data,
			err:      err,
		}
	}

	result, err := httpmux.HandleWithTimeout(ctx, createSessionHandler)
	if err != nil || result.err != nil {
		return authservice.AuthData{}, errors.Join(result.err, err)
	}

	return result.authData, nil
}

func (req ChangePasswordRequest) notifyAboutChangeWithEmail(ctx context.Context, r *http.Request) error {
	sendTo, err := req.userEmail(ctx)
	if err != nil {
		return err
	}

	sendingCtx, cancel := ctxWithTimeoutToSendMail()
	msg := emailSender.passwordIsChangedMessage(sendTo)

	go emailSender.addToQueue(sendingCtx, cancel, r, sendTo, msg)
	return nil
}

func (req ChangePasswordRequest) userEmail(ctx context.Context) (string, error) {
	getUserResult := req.authService.AuthorizedUser(ctx)
	if err := getUserResult.Error(); err != nil {
		return "", err
	}

	return getUserResult.UserData().Email, nil
}
