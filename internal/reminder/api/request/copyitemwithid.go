package request

import (
	"errors"
	"net/http"

	"github.com/zhuboris/never-expires/internal/reminder/api/endpoint"
	"github.com/zhuboris/never-expires/internal/shared/reqbody"
	"github.com/zhuboris/never-expires/internal/shared/rwjson"
)

type CopyItemWithIDRequest struct {
	items ItemService
}

func NewCopyItemWithIDRequest(items ItemService) *CopyItemWithIDRequest {
	return &CopyItemWithIDRequest{
		items: items,
	}
}

func (req CopyItemWithIDRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	body := new(copyData)
	if err := reqbody.Decode(body, r.Body); err != nil {
		return errors.Join(ErrInvalidBody, err)
	}

	idNew, err := idFromPath(r.URL.Path, endpoint.ItemsMakeCopy)
	if err != nil {
		return err
	}

	toCopy, err := body.toValidCopyData()
	if err != nil {
		return err
	}

	toCopy.NewID = idNew
	newCopy, err := req.items.Copy(r.Context(), toCopy)
	if err != nil {
		return err
	}

	return rwjson.WriteJSON(w, http.StatusOK, newCopy.ToResponseFormat())
}
