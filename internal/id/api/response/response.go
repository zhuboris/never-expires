package response

import (
	"fmt"
	"net/http"

	"github.com/zhuboris/never-expires/internal/shared/rwjson"
	"github.com/zhuboris/never-expires/internal/shared/str"
)

func WriteMessage(w http.ResponseWriter, statusCode int, rawMessage string) {
	msg := makeMessage(statusCode, rawMessage)
	err := rwjson.WriteJSON(w, statusCode, msg)
	handleWritingError(w, statusCode, rawMessage, err)
}

func WriteJSONData(w http.ResponseWriter, statusCode int, data any) error {
	return rwjson.WriteJSON(w, statusCode, data)
}

func makeMessage(statusCode int, rawMessage string) (msg any) {
	if statusCode < 200 && statusCode >= 300 {
		return struct {
			Error string `json:"error"`
		}{rawMessage}
	}

	return struct {
		Success string `json:"success"`
	}{rawMessage}
}

func handleWritingError(w http.ResponseWriter, statusCode int, rawMessage string, err error) {
	if err != nil {
		msg := fmt.Sprint(rwjson.ErrWriting.Error(), ", initial message was: ", rawMessage)
		http.Error(w, str.Capitalized(msg), statusCode)
	}
}
