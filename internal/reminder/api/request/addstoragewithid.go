package request

import (
	"errors"
	"net/http"

	"github.com/zhuboris/never-expires/internal/reminder/api/endpoint"
	"github.com/zhuboris/never-expires/internal/shared/reqbody"
	"github.com/zhuboris/never-expires/internal/shared/rwjson"
)

type AddStorageWithIDRequest struct {
	storages StorageService
}

func NewAddStorageWithIDRequest(storages StorageService) *AddStorageWithIDRequest {
	return &AddStorageWithIDRequest{
		storages: storages,
	}
}

func (req AddStorageWithIDRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	id, err := idFromPath(r.URL.Path, endpoint.StoragesWithParam)
	if err != nil {
		return err
	}

	body := new(storageData)
	if err := reqbody.Decode(body, r.Body); err != nil {
		return errors.Join(ErrInvalidBody, err)
	}

	activeStorage, err := req.storages.AddWithID(r.Context(), id, body.Name)
	if err != nil {
		return warpAddingError(err)
	}

	return rwjson.WriteJSON(w, http.StatusOK, activeStorage)
}
