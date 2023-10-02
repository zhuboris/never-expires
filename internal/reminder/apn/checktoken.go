package apn

import (
	"net/http"

	"github.com/sideshow/apns2"
)

func isTokenInactive(resp *apns2.Response) bool {
	const badDeviceTokenMsg = "BadDeviceToken"

	if resp == nil {
		return false
	}

	return resp.StatusCode == http.StatusGone || (resp.StatusCode == http.StatusBadRequest && resp.Reason == badDeviceTokenMsg)
}
