package request

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/zhuboris/never-expires/internal/id/authservice"
	"github.com/zhuboris/never-expires/internal/id/usr"
	"github.com/zhuboris/never-expires/internal/shared/httpmux"
)

type ResetPasswordRequest struct {
	authService AuthService
}

func NewResetPasswordRequest(authService AuthService) *ResetPasswordRequest {
	return &ResetPasswordRequest{
		authService: authService,
	}
}

func (req ResetPasswordRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	token := r.URL.
		Query().
		Get(tokenQueryName)

	if token == "" {
		return ErrMissingRequiredField
	}

	var (
		ctx     = r.Context()
		handler = func() authservice.ResetPasswordResult {
			return req.authService.ResetPassword(ctx, token)
		}
	)

	result, ctxError := httpmux.HandleWithTimeout(ctx, handler)
	resultError := result.Error()
	if resultError != nil || ctxError != nil {
		resultError = errors.Join(resultError, ctxError)
	}

	domain := "." + httpmux.RemoveSubdomain(r.Host)
	deleteAllAuthCookies(domain, w)

	req.redirectToStatusPage(w, r, resultError)
	if resultError != nil {
		return resultError
	}

	req.sendNewPassword(result.Address(), result.NewPassword(), r)
	return nil
}

func (req ResetPasswordRequest) redirectToStatusPage(w http.ResponseWriter, r *http.Request, resultError error) {
	status, err := req.setStatus(resultError)
	if err != nil {
		return
	}

	pageWithResult, err := req.urlToPasswordResetResultPage(r.Host, status)
	if err != nil {
		return
	}

	http.Redirect(w, r, pageWithResult, http.StatusFound)
}

func (req ResetPasswordRequest) sendNewPassword(sendTo, password string, r *http.Request) {
	var (
		sendingCtx, cancel = ctxWithTimeoutToSendMail()
		msg                = emailSender.newPasswordMessage(sendTo, password)
	)

	go emailSender.addToQueue(sendingCtx, cancel, r, sendTo, msg)
}

func (req ResetPasswordRequest) setStatus(err error) (string, error) {
	const (
		success = "success"
		failure = "failure"
	)

	switch {
	case err == nil:
		return success, nil
	case errors.Is(err, usr.ErrPasswordResetRefused):
		return failure, nil
	default:
		return "", err
	}
}

func (req ResetPasswordRequest) urlToPasswordResetResultPage(host, status string) (string, error) {
	const (
		route          = "/password-reset-status"
		statusQueryKey = "status"
	)

	urlRaw := "https://" + httpmux.RemoveSubdomain(host) + route
	redirectURL, err := url.Parse(urlRaw)
	if err != nil {
		return "", err
	}

	query := redirectURL.Query()
	query.Set(statusQueryKey, status)
	redirectURL.RawQuery = query.Encode()

	return redirectURL.String(), nil
}
