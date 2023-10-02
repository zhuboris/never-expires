package request

import (
	"net/http"

	"github.com/zhuboris/never-expires/internal/shared/rwjson"
)

type GetAllStoragesRequest struct {
	storages StorageService
}

func NewGetAllStoragesRequest(storages StorageService) *GetAllStoragesRequest {
	return &GetAllStoragesRequest{
		storages: storages,
	}
}

func (req GetAllStoragesRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	all, err := req.storages.All(r.Context())
	if err != nil {
		return err
	}

	return rwjson.WriteJSON(w, http.StatusOK, all)
}
