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

type ChangeMailRequest struct {
	authService AuthService
}

func NewChangeMailRequest(authService AuthService) *ChangeMailRequest {
	return &ChangeMailRequest{
		authService: authService,
	}
}

func (req ChangeMailRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	input := new(authservice.ChangeMailData)
	if err := reqbody.Decode(input, r.Body); err != nil {
		return errors.Join(ErrInvalidBody, err)
	}

	if input.IsMissingRequiredField() {
		return ErrMissingRequiredField
	}

	var (
		ctx     = r.Context()
		handler = func() error {
			return req.authService.UpdateMail(ctx, *input)
		}
	)

	if err := httpmux.HandleErrorFuncWithTimeout(ctx, handler); err != nil {
		return err
	}

	response.WriteMessage(w, http.StatusOK, "email changed")
	
	return req.sendNotificationEmail(ctx, r, input.NewEmail)
}

func (req ChangeMailRequest) sendNotificationEmail(ctx context.Context, r *http.Request, sendTo string) error {
	url, err := confirmEmailURL(ctx, r, req.authService, sendTo)
	if err != nil {
		return err
	}

	sendingCtx, cancel := ctxWithTimeoutToSendMail()
	msg := emailSender.confirmEmailOnChangeMessage(sendTo, url)
	go emailSender.addToQueue(sendingCtx, cancel, r, sendTo, msg)
	return nil
}
