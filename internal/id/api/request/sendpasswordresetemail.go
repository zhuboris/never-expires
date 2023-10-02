package request

import (
	"context"
	"errors"
	"net/http"

	"github.com/zhuboris/never-expires/internal/id/api/endpoint"
	"github.com/zhuboris/never-expires/internal/id/api/response"
	"github.com/zhuboris/never-expires/internal/id/authservice"
	"github.com/zhuboris/never-expires/internal/shared/httpmux"
	"github.com/zhuboris/never-expires/internal/shared/reqbody"
)

type SendResetPasswordEmailRequest struct {
	authService AuthService
}

var ErrMustConfirmEmail = errors.New("to restore password email must be confirmed")

func NewSendResetPasswordEmailRequest(authService AuthService) *SendResetPasswordEmailRequest {
	return &SendResetPasswordEmailRequest{
		authService: authService,
	}
}

func (req SendResetPasswordEmailRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	input := new(authservice.SendRestorePasswordEmailData)
	if err := reqbody.Decode(input, r.Body); err != nil {
		return errors.Join(ErrInvalidBody, err)
	}

	if input.IsMissingRequiredField() {
		return ErrMissingRequiredField
	}

	var (
		ctx     = r.Context()
		handler = func() error {
			err := req.authService.CheckIfUserExists(ctx, input.Email)
			if err != nil {
				return err
			}

			isConfirmed, err := req.authService.IsConfirmed(ctx, input.Email)
			if isConfirmed || err != nil {
				return err
			}

			return ErrMustConfirmEmail
		}
	)

	err := httpmux.HandleErrorFuncWithTimeout(ctx, handler)
	if err != nil {
		return err
	}

	response.WriteMessage(w, http.StatusAccepted, "request to send email accepted")
	return req.sendEmail(ctx, r, input.Email)
}

func (req SendResetPasswordEmailRequest) sendEmail(ctx context.Context, r *http.Request, sendTo string) error {
	token, err := req.authService.AddPasswordResetToken(ctx, sendTo)
	if err != nil {
		return err
	}

	url, err := urlWithToken(r, endpoint.ResetPassword, token)
	if err != nil {
		return err
	}

	sendingCtx, cancel := ctxWithTimeoutToSendMail()
	msg := emailSender.resetPasswordMessage(sendTo, url)
	go emailSender.addToQueue(sendingCtx, cancel, r, sendTo, msg)
	return nil
}
