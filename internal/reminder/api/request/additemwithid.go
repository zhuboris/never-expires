package request

import (
	"errors"
	"net/http"

	"github.com/zhuboris/never-expires/internal/reminder/api/endpoint"
	"github.com/zhuboris/never-expires/internal/shared/reqbody"
	"github.com/zhuboris/never-expires/internal/shared/rwjson"
)

type AddItemWithIDRequest struct {
	items ItemService
}

func NewAddItemWithIDRequest(items ItemService) *AddItemWithIDRequest {
	return &AddItemWithIDRequest{
		items: items,
	}
}

func (req AddItemWithIDRequest) Handle(w http.ResponseWriter, r *http.Request) error {
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

	id, err := idFromPath(r.URL.Path, endpoint.ItemsWithParam)
	if err != nil {
		return err
	}

	toAdd.ID = id
	newItem, err := req.items.Add(r.Context(), body.StorageID, toAdd)
	if err != nil {
		return err
	}

	return rwjson.WriteJSON(w, http.StatusOK, newItem.ToResponseFormat())
}
