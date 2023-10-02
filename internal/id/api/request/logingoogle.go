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

type LoginWithGoogleRequest struct {
	authService AuthService
}

func NewLoginWithGoogleRequest(authService AuthService) *LoginWithGoogleRequest {
	return &LoginWithGoogleRequest{
		authService: authService,
	}
}

func (req LoginWithGoogleRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	startTime := time.Now()

	var body oauth.Token
	if err := reqbody.Decode(&body, r.Body); err != nil {
		return errors.Join(ErrInvalidBody, err)
	}

	if body.IsMissingIDToken() {
		return ErrMissingRequiredField
	}

	ctx := r.Context()
	googleUser, err := req.authService.UserFromGoogleIDToken(ctx, body)
	if err != nil {
		return err
	}

	googleLoginData := authservice.NewLoginWithOAuthData(googleUser, device.Info(r))
	handler := func() authservice.LoginResult {
		return req.authService.LoginWithOAuth(ctx, googleLoginData, req.authService.WithGoogle())
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

	return sendEmailOnOAuthLogin(result.OAuthResultType(), loginData, oauth.GoogleAccount)
}
