package request

import (
	"errors"
	"net/http"

	"github.com/zhuboris/never-expires/internal/shared/reqbody"
)

type AddApnsTokenRequest struct {
	apns ApnsService
}

func NewAddApnsTokenRequest(apns ApnsService) *AddApnsTokenRequest {
	return &AddApnsTokenRequest{
		apns: apns,
	}
}

func (req AddApnsTokenRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	var data apnsDeviceToken
	if err := reqbody.Decode(&data, r.Body); err != nil {
		return errors.Join(ErrInvalidBody, err)
	}

	if data.isMissing() {
		return ErrMissingRequiredField
	}

	if err := req.apns.AddDeviceToken(r.Context(), data.Token); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
