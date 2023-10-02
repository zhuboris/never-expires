package request

import (
	"net/http"

	"github.com/zhuboris/never-expires/internal/reminder/api/endpoint"
)

type DeleteItemRequest struct {
	items ItemService
}

func NewDeleteItemRequest(items ItemService) *DeleteItemRequest {
	return &DeleteItemRequest{
		items: items,
	}
}

func (req DeleteItemRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	id, err := idFromPath(r.URL.Path, endpoint.ItemsWithParam)
	if err != nil {
		return err
	}

	if err := req.items.Delete(r.Context(), id); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
