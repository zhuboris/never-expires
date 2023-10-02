package request

import (
	"context"
	"net/http"
	"net/url"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zhuboris/never-expires/internal/reminder/api/endpoint"
)

type deleteOptionFunc func(ctx context.Context, storageID pgtype.UUID) error

const (
	deleteOption = "delete"
	clearOption  = "clear"
)

type DeleteStorageRequest struct {
	storages StorageService
}

func NewDeleteStorageRequest(storages StorageService) *DeleteStorageRequest {
	return &DeleteStorageRequest{
		storages: storages,
	}
}

func (req DeleteStorageRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	id, err := idFromPath(r.URL.Path, endpoint.StoragesWithParam)
	if err != nil {
		return err
	}

	optionFunc, err := req.matchOption(r.URL.Query())
	if err != nil {
		return err
	}

	if err := optionFunc(r.Context(), id); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (req DeleteStorageRequest) matchOption(query url.Values) (deleteOptionFunc, error) {
	optionRaw := query.Get(endpoint.OptionQueryKey)
	switch optionRaw {
	case "":
		return nil, ErrMissingParam
	case deleteOption:
		return req.storages.Delete, nil
	case clearOption:
		return req.storages.Clear, nil
	default:
		return nil, ErrOptionNotExists
	}
}
