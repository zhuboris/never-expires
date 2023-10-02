package request

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/zhuboris/never-expires/internal/id/usr"
	"github.com/zhuboris/never-expires/internal/shared/httpmux"
)

type EmailConfirmationRequest struct {
	authService AuthService
}

func NewEmailConfirmationRequest(authService AuthService) *EmailConfirmationRequest {
	return &EmailConfirmationRequest{
		authService: authService,
	}
}

func (req EmailConfirmationRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	token := r.URL.
		Query().
		Get(tokenQueryName)

	var (
		ctx     = r.Context()
		handler = func() error {
			return req.authService.ConfirmEmail(r.Context(), token)
		}
	)

	validationError := httpmux.HandleErrorFuncWithTimeout(ctx, handler)
	redirectionErr := req.redirectToStatusPage(w, r, validationError)
	return errors.Join(validationError, redirectionErr)
}

func (req EmailConfirmationRequest) redirectToStatusPage(w http.ResponseWriter, r *http.Request, validationError error) error {
	status, err := req.setStatus(validationError)
	if err != nil {
		return err
	}

	pageWithResult, err := req.urlToConfirmationResultPage(r.Host, status)
	if err != nil {
		return err
	}

	http.Redirect(w, r, pageWithResult, http.StatusFound)
	return nil
}

func (req EmailConfirmationRequest) setStatus(err error) (string, error) {
	const (
		success          = "success"
		alreadyConfirmed = "already_confirmed"
		failure          = "failure"
	)

	switch {
	case err == nil:
		return success, nil
	case errors.Is(err, usr.ErrEmailAlreadyConfirmed):
		return alreadyConfirmed, nil
	case errors.Is(err, usr.ErrValidationRefused):
		return failure, nil
	default:
		return "", err
	}
}

func (req EmailConfirmationRequest) urlToConfirmationResultPage(host, status string) (string, error) {
	const (
		route          = "/confirmation-status"
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
