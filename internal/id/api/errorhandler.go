package api

import (
	"errors"
	"net/http"

	"github.com/zhuboris/never-expires/internal/id/api/request"
	"github.com/zhuboris/never-expires/internal/id/authservice"
	"github.com/zhuboris/never-expires/internal/id/pw"
	"github.com/zhuboris/never-expires/internal/id/tkn"
	"github.com/zhuboris/never-expires/internal/id/usr"
	"github.com/zhuboris/never-expires/internal/shared/httpmux"
	"github.com/zhuboris/never-expires/internal/shared/servicechecker"
)

const (
	StatusEmailAlreadyRegistered       httpmux.StatusCode = 2001
	StatusWrongLoginData               httpmux.StatusCode = 2002
	StatusWrongPassword                httpmux.StatusCode = 2003
	StatusEmailAlreadyConfirmed        httpmux.StatusCode = 2004
	StatusEmailIsNotConfirmed          httpmux.StatusCode = 2005
	StatusEmailIsChangedOrNotConfirmed httpmux.StatusCode = 2006
	StatusEmailIsNotBelongToAnyUser    httpmux.StatusCode = 2007
	StatusUserNotFound                 httpmux.StatusCode = 2008
	StatusInvalidEmail                 httpmux.StatusCode = 3001
	StatusInsecurePassword             httpmux.StatusCode = 3002
)

func handleResponseErrors(err error) httpmux.RequestingResult {
	const StatusClientClosedRequest = 499

	if err == nil {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Info).
			AddStatusCode(http.StatusOK).
			AddResponseMessage(httpmux.StatusCode(http.StatusOK).SuccessMessage("successfully completed request")).
			WithoutResponse().
			Build()
	}

	if errors.Is(err, httpmux.ErrTimeout) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusRequestTimeout).
			AddError(err).
			Build()
	}

	if errors.Is(err, httpmux.ErrCanceled) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(StatusClientClosedRequest).
			WithoutResponse().
			AddError(err).
			Build()
	}

	if errors.Is(err, tkn.ErrForbidden) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnauthorized).
			AddError(err).
			Build()
	}

	if errors.Is(err, tkn.ErrUnauthorized) || errors.Is(err, httpmux.ErrUnauthorized) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnauthorized).
			AddError(err).
			Build()
	}

	if errors.Is(err, httpmux.ErrInvalidMethod) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddInternalErrorCode(httpmux.StatusUnexistingHTTPMethod).
			AddResponseMessage(httpmux.StatusUnexistingHTTPMethod.ErrorMessage(httpmux.ErrInvalidMethod.Error())).
			AddError(err).
			Build()
	}

	if errors.Is(err, httpmux.ErrMethodNotAllowed) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusMethodNotAllowed).
			AddError(err).
			Build()
	}

	if errors.Is(err, authservice.ErrAlreadyRegistered) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddInternalErrorCode(StatusEmailAlreadyRegistered).
			AddResponseMessage(StatusEmailAlreadyRegistered.ErrorMessage(authservice.ErrAlreadyRegistered.Error())).
			AddError(err).
			Build()
	}

	if errors.Is(err, authservice.ErrWrongLoginData) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddInternalErrorCode(StatusWrongLoginData).
			AddResponseMessage(StatusWrongLoginData.ErrorMessage(authservice.ErrWrongLoginData.Error())).
			AddError(err).
			Build()
	}

	if errors.Is(err, pw.ErrWrongPassword) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddInternalErrorCode(StatusWrongPassword).
			AddResponseMessage(StatusWrongPassword.ErrorMessage(pw.ErrWrongPassword.Error())).
			AddError(err).
			Build()
	}

	if errors.Is(err, request.ErrInvalidBody) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddInternalErrorCode(httpmux.StatusInvalidJSONBody).
			AddResponseMessage(httpmux.StatusInvalidJSONBody.ErrorMessage(request.ErrInvalidBody.Error())).
			AddError(err).
			Build()
	}

	if errors.Is(err, request.ErrMissingRequiredField) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddInternalErrorCode(httpmux.StatusMissingParameter).
			AddResponseMessage(httpmux.StatusMissingParameter.ErrorMessage(request.ErrMissingRequiredField.Error())).
			AddError(err).
			Build()
	}

	if errors.Is(err, authservice.ErrMissingEmailAddress) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddInternalErrorCode(StatusEmailIsNotBelongToAnyUser).
			AddResponseMessage(StatusEmailIsNotBelongToAnyUser.ErrorMessage(authservice.ErrMissingEmailAddress.Error())).
			AddError(err).
			Build()
	}

	if errors.Is(err, usr.ErrNotFound) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddInternalErrorCode(StatusUserNotFound).
			AddResponseMessage(StatusUserNotFound.ErrorMessage(usr.ErrNotFound.Error())).
			AddError(err).
			Build()
	}

	if errors.Is(err, usr.ErrNotConfirmedOrChangedEmail) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddInternalErrorCode(StatusEmailIsChangedOrNotConfirmed).
			AddResponseMessage(StatusEmailIsChangedOrNotConfirmed.ErrorMessage(usr.ErrNotConfirmedOrChangedEmail.Error())).
			AddError(err).
			Build()
	}

	if errors.Is(err, usr.ErrInvalidEmail) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddInternalErrorCode(StatusInvalidEmail).
			AddResponseMessage(StatusInvalidEmail.ErrorMessage(usr.ErrInvalidEmail.Error())).
			AddError(err).
			Build()
	}

	if errors.Is(err, request.ErrMustConfirmEmail) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddInternalErrorCode(StatusEmailIsNotConfirmed).
			AddResponseMessage(StatusEmailIsNotConfirmed.ErrorMessage(request.ErrMustConfirmEmail.Error())).
			AddError(err).
			Build()
	}

	if errInvalidPassword := new(pw.InsecurePasswordError); errors.As(err, errInvalidPassword) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddInternalErrorCode(StatusInsecurePassword).
			AddResponseMessage(StatusInsecurePassword.ErrorMessage(errInvalidPassword.Reason())).
			AddError(err).
			Build()
	}

	if errors.Is(err, usr.ErrMissingConfirmToken) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddInternalErrorCode(httpmux.StatusMissingParameter).
			AddResponseMessage(httpmux.StatusMissingParameter.ErrorMessage(usr.ErrMissingConfirmToken.Error())).
			AddError(err).
			Build()
	}

	if errors.Is(err, usr.ErrEmailAlreadyConfirmed) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddInternalErrorCode(StatusEmailAlreadyConfirmed).
			AddResponseMessage(StatusEmailAlreadyConfirmed.ErrorMessage(usr.ErrEmailAlreadyConfirmed.Error())).
			AddError(err).
			Build()
	}

	if errInvalidToken := new(usr.InvalidTokenError); errors.As(err, errInvalidToken) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnauthorized).
			AddError(err).
			Build()
	}

	if errors.Is(err, request.ErrMissingSessionOnLogout) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Info).
			AddStatusCode(http.StatusUnauthorized).
			WithoutResponse().
			AddError(err).
			Build()
	}

	if errFailedToSendEmail := new(request.SendingError); errors.As(err, errFailedToSendEmail) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusInternalServerError).
			WithoutResponse().
			AddError(err).
			Build()
	}

	if errFailedMakeURL := new(httpmux.URLMakingError); errors.As(err, errFailedMakeURL) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusInternalServerError).
			WithoutResponse().
			AddError(err).
			Build()
	}

	if errFailedHealthCheck := new(servicechecker.IsUnavailableError); errors.As(err, errFailedHealthCheck) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusServiceUnavailable).
			WithoutResponse().
			AddError(err).
			Build()
	}

	return httpmux.NewRequestingResultBuilder().
		SetType(httpmux.Error).
		AddStatusCode(http.StatusInternalServerError).
		AddError(err).
		Build()
}
