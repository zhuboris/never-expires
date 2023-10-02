package request

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/zhuboris/never-expires/internal/reminder/api/endpoint"
	"github.com/zhuboris/never-expires/internal/shared/rwjson"
)

const defaultSuggestionsLimit = 10

type ItemsAutocompleteSuggestionsRequest struct {
	items ItemService
}

func NewItemsAutocompleteSuggestionsRequest(items ItemService) *ItemsAutocompleteSuggestionsRequest {
	return &ItemsAutocompleteSuggestionsRequest{
		items: items,
	}
}

func (req ItemsAutocompleteSuggestionsRequest) Handle(w http.ResponseWriter, r *http.Request) error {
	query := r.URL.Query()
	toSearch := query.Get(endpoint.SearchQueryKey)
	if toSearch == "" {
		return ErrMissingRequiredField
	}

	limit := req.limit(query)
	names, err := req.items.SearchSavedNames(r.Context(), toSearch, limit)
	if err != nil {
		return err
	}

	responseBody := struct {
		Data *[]string `json:"data"`
	}{names}

	return rwjson.WriteJSON(w, http.StatusOK, responseBody)
}

func (req ItemsAutocompleteSuggestionsRequest) limit(query url.Values) int {
	limitRaw := query.Get(endpoint.SearchLimitQueryKey)
	limit, err := strconv.Atoi(limitRaw)
	if err != nil {
		return defaultSuggestionsLimit
	}

	return limit
}
