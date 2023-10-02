package request

import (
	"net/http"

	"github.com/zhuboris/never-expires/internal/reminder/api/endpoint"
	"github.com/zhuboris/never-expires/internal/shared/rwjson"
)

type GetItemRequest struct {
	items ItemService
}

func NewGetItemRequest(items ItemService) *GetItemRequest {
	return &GetItemRequest{
		items: items,
	}
}

func (req GetItemRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	id, err := idFromPath(r.URL.Path, endpoint.ItemsWithParam)
	if err != nil {
		return err
	}

	item, err := req.items.ByID(r.Context(), id)
	if err != nil {
		return err
	}

	return rwjson.WriteJSON(w, http.StatusOK, item.ToResponseFormat())
}
