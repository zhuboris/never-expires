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

type SendConfirmationEmailRequest struct {
	authService AuthService
}

func NewSendConfirmationEmailRequest(authService AuthService) *SendConfirmationEmailRequest {
	return &SendConfirmationEmailRequest{
		authService: authService,
	}
}

func (req SendConfirmationEmailRequest) Handle(w http.ResponseWriter, r *http.Request) error {
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

	currentUser := result.UserData()
	if currentUser.IsEmailConfirmed {
		return usr.ErrEmailAlreadyConfirmed
	}

	response.WriteMessage(w, http.StatusAccepted, "request to send email accepted")
	return req.sendEmail(ctx, r, currentUser.Email)
}

func (req SendConfirmationEmailRequest) sendEmail(ctx context.Context, r *http.Request, sendTo string) error {
	url, err := confirmEmailURL(ctx, r, req.authService, sendTo)
	if err != nil {
		return err
	}

	sendingCtx, cancel := ctxWithTimeoutToSendMail()
	msg := emailSender.confirmEmailMessage(sendTo, url)
	go emailSender.addToQueue(sendingCtx, cancel, r, sendTo, msg)
	return nil
}
