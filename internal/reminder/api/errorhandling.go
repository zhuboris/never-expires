package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/zhuboris/never-expires/internal/id/tkn"
	"github.com/zhuboris/never-expires/internal/reminder/api/request"
	"github.com/zhuboris/never-expires/internal/reminder/item"
	"github.com/zhuboris/never-expires/internal/reminder/queryerr"
	"github.com/zhuboris/never-expires/internal/reminder/storage"
	"github.com/zhuboris/never-expires/internal/shared/httpmux"
	"github.com/zhuboris/never-expires/internal/shared/uuidformat"
)

const (
	StatusInvalidUUID      httpmux.StatusCode = 1004
	StatusInvalidOption    httpmux.StatusCode = 1005
	StatusInvalidQueryData httpmux.StatusCode = 1006

	StatusUUIDIsReserved httpmux.StatusCode = 3003

	StatusItemNotFound             httpmux.StatusCode = 4001
	StatusStorageNotFound          httpmux.StatusCode = 4002
	StatusStorageNameAlreadyExists httpmux.StatusCode = 4003
	StatusDeletingNotAllowed       httpmux.StatusCode = 4004
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

	if errors.Is(err, httpmux.ErrTimeout) || errors.Is(err, context.DeadlineExceeded) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusRequestTimeout).
			AddError(err).
			Build()
	}

	if errors.Is(err, httpmux.ErrCanceled) || errors.Is(err, context.Canceled) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(StatusClientClosedRequest).
			WithoutResponse().
			AddError(err).
			Build()
	}

	if errors.Is(err, httpmux.ErrUnauthorized) || errors.Is(err, tkn.ErrUnauthorized) {
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

	if errors.Is(err, request.ErrInvalidBody) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddResponseMessage(httpmux.StatusInvalidJSONBody.ErrorMessage(request.ErrInvalidBody.Error())).
			AddError(err).
			Build()
	}

	if errors.Is(err, request.ErrMissingParam) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddResponseMessage(httpmux.StatusMissingParameter.ErrorMessage(request.ErrMissingParam.Error())).
			AddError(err).
			Build()
	}

	if errors.Is(err, request.ErrMissingRequiredField) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddResponseMessage(httpmux.StatusMissingParameter.ErrorMessage(request.ErrMissingRequiredField.Error())).
			AddError(err).
			Build()
	}

	if errors.Is(err, request.ErrOptionNotExists) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddResponseMessage(StatusInvalidOption.ErrorMessage(request.ErrOptionNotExists.Error())).
			AddError(err).
			Build()
	}

	if errors.Is(err, request.ErrInvalidQuery) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddResponseMessage(StatusInvalidQueryData.ErrorMessage(request.ErrInvalidQuery.Error())).
			AddError(err).
			Build()
	}

	if errors.Is(err, request.ErrInvalidTimeFormat) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddResponseMessage(httpmux.StatusInvalidJSONBody.ErrorMessage(err.Error())).
			AddError(err).
			Build()
	}

	if errors.Is(err, uuidformat.ErrInvalidUUID) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddResponseMessage(StatusInvalidUUID.ErrorMessage(uuidformat.ErrInvalidUUID.Error())).
			AddError(err).
			Build()
	}

	if errors.Is(err, request.ErrInvalidUUIDInBody) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddResponseMessage(StatusInvalidUUID.ErrorMessage(request.ErrInvalidUUIDInBody.Error())).
			AddError(err).
			Build()
	}

	if errors.Is(err, request.ErrNewUUIDNotUnique) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddResponseMessage(StatusUUIDIsReserved.ErrorMessage(request.ErrNewUUIDNotUnique.Error())).
			AddError(err).
			Build()
	}

	if errors.Is(err, queryerr.ErrStorageNotExists) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddResponseMessage(StatusStorageNotFound.ErrorMessage(queryerr.ErrStorageNotExists.Error())).
			AddError(err).
			Build()
	}

	if errors.Is(err, item.ErrItemNotExists) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddResponseMessage(StatusItemNotFound.ErrorMessage(item.ErrItemNotExists.Error())).
			AddError(err).
			Build()
	}

	if errors.Is(err, storage.ErrStorageNameNotUnique) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddResponseMessage(StatusStorageNameAlreadyExists.ErrorMessage(storage.ErrStorageNameNotUnique.Error())).
			AddError(err).
			Build()
	}

	if errors.Is(err, storage.ErrDeletingNotAllowed) {
		return httpmux.NewRequestingResultBuilder().
			SetType(httpmux.Error).
			AddStatusCode(http.StatusUnprocessableEntity).
			AddResponseMessage(StatusDeletingNotAllowed.ErrorMessage(storage.ErrDeletingNotAllowed.Error())).
			AddError(err).
			Build()
	}

	return httpmux.NewRequestingResultBuilder().
		SetType(httpmux.Error).
		AddStatusCode(http.StatusInternalServerError).
		AddError(err).
		Build()
}
