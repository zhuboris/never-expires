package reqbody

import (
	"encoding/json"
	"io"
)

func Decode(to any, from io.ReadCloser) error {
	defer from.Close()

	return json.NewDecoder(from).Decode(to)
}
