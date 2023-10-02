package request

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/zhuboris/never-expires/internal/reminder/api/endpoint"
	"github.com/zhuboris/never-expires/internal/reminder/item"
	"github.com/zhuboris/never-expires/internal/shared/rwjson"
	"github.com/zhuboris/never-expires/internal/shared/uuidformat"
)

type GetAllItemsRequest struct {
	items ItemService
}

func NewGetAllItemsRequest(items ItemService) *GetAllItemsRequest {
	return &GetAllItemsRequest{
		items: items,
	}
}

func (req GetAllItemsRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	filters, err := req.requestedFilters(r.URL.Query())
	if err != nil {
		return err
	}

	all, err := req.items.All(r.Context(), filters...)
	if err != nil {
		return err
	}

	return rwjson.WriteJSON(w, http.StatusOK, all.ToResponseFormat())
}

func (req GetAllItemsRequest) requestedFilters(query url.Values) ([]item.Filter, error) {
	filters := make([]item.Filter, 0, len(query))
	filters, err := req.checkByStorageFilter(query, filters)
	if err != nil {
		return nil, err
	}

	filters, err = req.checkByOpenedFilter(query, filters)
	if err != nil {
		return nil, err
	}

	filters, err = req.checkByDateFilter(query, filters)
	if err != nil {
		return nil, err
	}

	filters = req.checkByNameFilter(query, filters)

	return filters, nil
}

func (req GetAllItemsRequest) checkByStorageFilter(query url.Values, filters []item.Filter) ([]item.Filter, error) {
	if storageID := query.Get(endpoint.StorageIDQueryKey); storageID != "" {
		id, err := uuidformat.StrToPgtype(storageID)
		if err != nil {
			return nil, ErrInvalidQuery
		}

		filters = append(filters, item.ByStorageID(id))
	}
	return filters, nil
}

func (req GetAllItemsRequest) checkByOpenedFilter(query url.Values, filters []item.Filter) ([]item.Filter, error) {
	if openedStatus := query.Get(endpoint.IsOpenedQueryKey); openedStatus != "" {
		isOpened, err := strToBool(openedStatus)
		if err != nil {
			return nil, err
		}

		filters = append(filters, item.ByOpenedStatus(isOpened))
	}
	return filters, nil
}

func (req GetAllItemsRequest) checkByDateFilter(query url.Values, filters []item.Filter) ([]item.Filter, error) {
	if beforeDate := query.Get(endpoint.BeforeDateQueryKey); beforeDate != "" {
		date, err := time.Parse(time.RFC3339, beforeDate)
		if err != nil {
			return nil, errors.Join(ErrInvalidQuery, InvalidTimeFormatError(beforeDate))
		}

		filters = append(filters, item.ByDateBefore(date))
	}
	return filters, nil
}

func (req GetAllItemsRequest) checkByNameFilter(query url.Values, filters []item.Filter) []item.Filter {
	if nameStarting := query.Get(endpoint.NameStartingQueryKey); nameStarting != "" {
		pattern := nameStarting + "%"
		filters = append(filters, item.ByName(pattern))
	}

	return filters
}

func strToBool(input string) (bool, error) {
	switch input {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, ErrInvalidQuery
	}
}
