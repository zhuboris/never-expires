package request

import (
	"errors"
	"net/http"
	"time"

	"github.com/zhuboris/never-expires/internal/id/api/request/device"
	"github.com/zhuboris/never-expires/internal/id/authservice"
	"github.com/zhuboris/never-expires/internal/id/usr/oauth"
	"github.com/zhuboris/never-expires/internal/shared/reqbody"
)

type LoginWithAppleRequest struct {
	authService AuthService
}

func NewLoginWithAppleRequest(authService AuthService) *LoginWithAppleRequest {
	return &LoginWithAppleRequest{
		authService: authService,
	}
}

func (req LoginWithAppleRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	startTime := time.Now()

	var body oauth.Token
	if err := reqbody.Decode(&body, r.Body); err != nil {
		return errors.Join(ErrInvalidBody, err)
	}

	if body.IsMissingIDToken() || body.IsMissingAuthCode() {
		return ErrMissingRequiredField
	}

	ctx := r.Context()
	appleUser, err := req.authService.UserFromAppleTokenCode(ctx, body)
	if err != nil {
		return err
	}

	appleLoginData := authservice.NewLoginWithOAuthData(appleUser, device.Info(r))
	handler := func() authservice.LoginResult {
		return req.authService.LoginWithOAuth(ctx, appleLoginData, req.authService.WithApple())
	}

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

	return sendEmailOnOAuthLogin(result.OAuthResultType(), loginData, oauth.AppleID)
}
