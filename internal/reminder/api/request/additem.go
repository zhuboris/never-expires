package request

import (
	"errors"
	"net/http"

	"github.com/zhuboris/never-expires/internal/shared/reqbody"
	"github.com/zhuboris/never-expires/internal/shared/rwjson"
)

type AddItemRequest struct {
	items ItemService
}

func NewAddItemRequest(items ItemService) *AddItemRequest {
	return &AddItemRequest{
		items: items,
	}
}

func (req AddItemRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	body := new(itemData)
	if err := reqbody.Decode(body, r.Body); err != nil {
		return errors.Join(ErrInvalidBody, err)
	}

	if err := body.checkIfStorageIDValid(); err != nil {
		return err
	}

	toAdd, err := body.toValidItemWithAddedDateRequired()
	if err != nil {
		return err
	}

	newItem, err := req.items.Add(r.Context(), body.StorageID, toAdd)
	if err != nil {
		return err
	}

	return rwjson.WriteJSON(w, http.StatusOK, newItem.ToResponseFormat())
}
