package httpmux

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/zhuboris/never-expires/internal/id/api/response"
)

type resultType int

const (
	Info resultType = iota + 1
	Error
)

type RequestingResult struct {
	err               error
	statusCode        int
	internalErrorCode StatusCode
	responseMsg       any
	shouldResponse    bool
	resultType        resultType
}

func (r RequestingResult) AddToRespond(w http.ResponseWriter) {
	if !r.shouldResponse {
		return
	}

	if r.responseMsg == nil {
		w.WriteHeader(r.statusCode)
		return
	}

	writeResponse(w, r.statusCode, r)
}

func (r RequestingResult) responseCodes() (statusCode, internalErrorCode int) {
	if !r.shouldResponse {
		return http.StatusOK, 0
	}

	return r.statusCode, int(r.internalErrorCode)
}

type RequestingResultBuilder struct {
	result RequestingResult
}

func NewRequestingResultBuilder() *RequestingResultBuilder {
	return &RequestingResultBuilder{
		result: RequestingResult{
			statusCode:     http.StatusOK,
			shouldResponse: true,
			resultType:     Info,
		},
	}
}

func (rb *RequestingResultBuilder) AddResponseMessage(msg any) *RequestingResultBuilder {
	rb.result.responseMsg = msg
	return rb
}

func (rb *RequestingResultBuilder) SetType(resultType resultType) *RequestingResultBuilder {
	rb.result.resultType = resultType
	return rb
}

func (rb *RequestingResultBuilder) AddStatusCode(code int) *RequestingResultBuilder {
	rb.result.statusCode = code
	return rb
}

func (rb *RequestingResultBuilder) AddInternalErrorCode(code StatusCode) *RequestingResultBuilder {
	rb.result.internalErrorCode = code
	return rb
}

func (rb *RequestingResultBuilder) AddError(err error) *RequestingResultBuilder {
	rb.result.err = err
	return rb
}

func (rb *RequestingResultBuilder) WithoutResponse() *RequestingResultBuilder {
	rb.result.shouldResponse = false
	return rb
}

func (rb *RequestingResultBuilder) Build() RequestingResult {
	return rb.result
}

func logResult(logger *zap.Logger, result RequestingResult, elapsedTime time.Duration) {
	lvl := zapcore.ErrorLevel
	if result.resultType == Info {
		lvl = zapcore.InfoLevel
	}

	logger.Log(lvl, "response completed",
		zap.Any(responseLogKey, result.responseMsg),
		zap.Duration(elapsedTimeLogKey, elapsedTime),
		zap.Int(statusCodeLogKey, result.statusCode),
		zap.Error(result.err),
	)
}

func writeResponse(w http.ResponseWriter, code int, handledResult RequestingResult) {
	if err := response.WriteJSONData(w, code, handledResult.responseMsg); err == nil { // if NO err
		return
	}

	if code == http.StatusOK {
		code = http.StatusInternalServerError
	}

	errorMsg := fmt.Sprintf("Error writing response. Status code was %q", code)
	http.Error(w, errorMsg, 500)
}
