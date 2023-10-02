package request

import (
	"context"
	"errors"
	"net/http"

	"github.com/zhuboris/never-expires/internal/id/api/response"
	"github.com/zhuboris/never-expires/internal/id/authservice"
	"github.com/zhuboris/never-expires/internal/shared/httpmux"
	"github.com/zhuboris/never-expires/internal/shared/reqbody"
)

type RegisterRequest struct {
	authService AuthService
}

func NewRegisterRequest(authService AuthService) *RegisterRequest {
	return &RegisterRequest{
		authService: authService,
	}
}

func (req RegisterRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	input := new(authservice.RegisterData)
	if err := reqbody.Decode(input, r.Body); err != nil {
		return errors.Join(ErrInvalidBody, err)
	}

	if input.IsMissingRequiredField() {
		return ErrMissingRequiredField
	}

	var (
		ctx     = r.Context()
		handler = func() authservice.RegisterResult {
			return req.authService.Register(ctx, *input)
		}
	)

	result, ctxError := httpmux.HandleWithTimeout(ctx, handler)
	if resultError := result.Error(); resultError != nil || ctxError != nil {
		return errors.Join(resultError, ctxError)
	}

	response.WriteMessage(w, http.StatusCreated, "new account registered")

	err2 := req.sendNotifyEmail(ctx, r, result.User().Email)
	if err2 != nil {
		return err2
	}
	return nil
}

func (req RegisterRequest) sendNotifyEmail(ctx context.Context, r *http.Request, sendTo string) error {
	url, err := confirmEmailURL(ctx, r, req.authService, sendTo)
	if err != nil {
		return err
	}

	var (
		sendingCtx, cancel = ctxWithTimeoutToSendMail()
		msg                = emailSender.registerMessage(sendTo, emailSender.registerWithConfirmationButton(url))
	)

	go emailSender.addToQueue(sendingCtx, cancel, r, sendTo, msg)
	return nil
}
