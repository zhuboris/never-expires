package request

import (
	"errors"
	"net/http"

	"github.com/zhuboris/never-expires/internal/shared/reqbody"
	"github.com/zhuboris/never-expires/internal/shared/rwjson"
)

type AddStorageRequest struct {
	storages StorageService
}

func NewAddStorageRequest(storages StorageService) *AddStorageRequest {
	return &AddStorageRequest{
		storages: storages,
	}
}

func (req AddStorageRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	body := new(storageData)
	if err := reqbody.Decode(body, r.Body); err != nil {
		return errors.Join(ErrInvalidBody, err)
	}

	activeStorage, err := req.storages.Add(r.Context(), body.Name)
	if err != nil {
		return err
	}

	return rwjson.WriteJSON(w, http.StatusOK, activeStorage)
}
