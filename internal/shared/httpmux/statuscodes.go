package httpmux

type (
	ErrorMessage struct {
		StatusCode   StatusCode `json:"status_code"`
		ErrorMessage string     `json:"error"`
	}
	SuccessMessage struct {
		StatusCode     StatusCode `json:"status_code"`
		SuccessMessage string     `json:"success_message"`
	}
)

// StatusCode is type for custom response error code statuses. It has 4 digit format, where first digit is a category as in HTTP errors
//
// 1xxx: HTTP Request forming errors
//
// 2xxx: Authentication errors
//
// 3xxx: Data validation errors
type StatusCode int

func (c StatusCode) SuccessMessage(msg string) *SuccessMessage {
	return &SuccessMessage{
		StatusCode:     c,
		SuccessMessage: msg,
	}
}

func (c StatusCode) ErrorMessage(msg string) *ErrorMessage {
	return &ErrorMessage{
		StatusCode:   c,
		ErrorMessage: msg,
	}
}

const (
	StatusInvalidJSONBody      StatusCode = 1001
	StatusUnexistingHTTPMethod StatusCode = 1002
	StatusMissingParameter     StatusCode = 1003
)
