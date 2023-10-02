package request

import (
	"errors"
	"net/http"

	"github.com/zhuboris/never-expires/internal/shared/reqbody"
	"github.com/zhuboris/never-expires/internal/shared/rwjson"
)

type CopyItemRequest struct {
	items ItemService
}

func NewCopyItemRequest(items ItemService) *CopyItemRequest {
	return &CopyItemRequest{
		items: items,
	}
}

func (req CopyItemRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	body := new(copyData)
	if err := reqbody.Decode(body, r.Body); err != nil {
		return errors.Join(ErrInvalidBody, err)
	}

	toCopy, err := body.toValidCopyData()
	if err != nil {
		return err
	}

	newCopy, err := req.items.Copy(r.Context(), toCopy)
	if err != nil {
		return err
	}

	return rwjson.WriteJSON(w, http.StatusOK, newCopy.ToResponseFormat())
}
