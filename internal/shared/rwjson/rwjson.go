package rwjson

import (
	"encoding/json"
	"errors"
	"net/http"
)

var ErrWriting = errors.New("error forming response")

func WriteJSON(w http.ResponseWriter, statusCode int, data any) error {
	const contentTypeHeader = "Content-Type"
	const contentTypeValue = "application/json"

	w.Header().Add(contentTypeHeader, contentTypeValue)
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		err = errors.Join(ErrWriting, err)
	}

	return err
}
