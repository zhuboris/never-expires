package request

import (
	"errors"
	"net/http"
	"time"

	"github.com/zhuboris/never-expires/internal/id/api/request/device"
	"github.com/zhuboris/never-expires/internal/id/authservice"
	"github.com/zhuboris/never-expires/internal/shared/reqbody"
)

type SendingError struct {
	msg string
}

func (s SendingError) Error() string {
	return "failed to send email: " + s.msg
}

type LoginRequest struct {
	authService AuthService
}

func NewLoginRequest(authService AuthService) *LoginRequest {
	return &LoginRequest{
		authService: authService,
	}
}

func (req LoginRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	var (
		startTime  = time.Now()
		deviceInfo = device.Info(r)
		input      = authservice.NewLoginData(deviceInfo)
	)

	if err := reqbody.Decode(&input, r.Body); err != nil {
		return errors.Join(ErrInvalidBody, err)
	}

	if input.IsMissingRequiredField() {
		return ErrMissingRequiredField
	}

	var (
		ctx     = r.Context()
		handler = func() authservice.LoginResult {
			return req.authService.Login(ctx, input)
		}
	)

	result, err := handleLogin(ctx, handler, req.authService, w, r)
	if err != nil {
		return err
	}

	loginData := successLoginData{
		user:       result.UserData(),
		newSession: result.AuthData().Session(),
		loginTime:  startTime,
		request:    r,
		service:    req.authService,
	}

	return sendNewDeviceNotifyIfNeeded(loginData)
}
