package request

import (
	"errors"
	"io"
	"net/http"

	"github.com/zhuboris/never-expires/internal/reminder/api/endpoint"
	"github.com/zhuboris/never-expires/internal/reminder/storage"
	"github.com/zhuboris/never-expires/internal/shared/reqbody"
	"github.com/zhuboris/never-expires/internal/shared/rwjson"
)

type UpdateStorageRequest struct {
	storages StorageService
}

func NewUpdateStorageRequest(storages StorageService) *UpdateStorageRequest {
	return &UpdateStorageRequest{
		storages: storages,
	}
}

func (req UpdateStorageRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	updatedStorage, err := req.updatedStorage(r)
	if err != nil {
		return err
	}

	result, err := req.storages.Update(r.Context(), updatedStorage)
	if err != nil {
		return err
	}

	return rwjson.WriteJSON(w, http.StatusOK, result)
}

func (req UpdateStorageRequest) updatedStorage(r *http.Request) (storage.Storage, error) {
	id, err := idFromPath(r.URL.Path, endpoint.StoragesWithParam)
	if err != nil {
		return storage.Storage{}, err
	}

	body, err := req.decodedBody(r.Body)
	if err != nil {
		return storage.Storage{}, err
	}

	return storage.Storage{
		ID:   id,
		Name: body.Name,
	}, nil
}

func (req UpdateStorageRequest) decodedBody(rawBody io.ReadCloser) (*storageData, error) {
	body := new(storageData)
	if err := reqbody.Decode(body, rawBody); err != nil {
		return nil, errors.Join(ErrInvalidBody, err)
	}

	if body.isMissingRequiredField() {
		return nil, ErrMissingRequiredField
	}

	return body, nil
}
