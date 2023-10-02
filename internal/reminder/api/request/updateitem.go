package request

import (
	"errors"
	"net/http"

	"github.com/zhuboris/never-expires/internal/reminder/api/endpoint"
	"github.com/zhuboris/never-expires/internal/shared/reqbody"

	"github.com/zhuboris/never-expires/internal/shared/rwjson"
)

type UpdateItemRequest struct {
	items ItemService
}

func NewUpdateItemRequest(items ItemService) *UpdateItemRequest {
	return &UpdateItemRequest{
		items: items,
	}
}

func (req UpdateItemRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	body := new(itemData)
	if err := reqbody.Decode(body, r.Body); err != nil {
		return errors.Join(ErrInvalidBody, err)
	}

	id, err := idFromPath(r.URL.Path, endpoint.ItemsWithParam)
	if err != nil {
		return err
	}

	itemFromBody, err := body.toValidItem()
	if err != nil {
		return err
	}

	itemFromBody.ID = id
	updatedItem, err := req.items.Update(r.Context(), itemFromBody)
	if err != nil {
		return err
	}

	return rwjson.WriteJSON(w, http.StatusOK, updatedItem.ToResponseFormat())
}
