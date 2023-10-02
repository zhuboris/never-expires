package device

import (
	"net/http"
	"strings"

	"github.com/mileusna/useragent"
)

const identifierHeader = "X-Device-Identifier"

func IdentifierHeaderValue(r *http.Request) string {
	return r.Header.Get(identifierHeader)
}

func Info(r *http.Request) string {
	if deviceID := IdentifierHeaderValue(r); deviceID != "" {
		return "ID: " + deviceID
	}

	userAgent := useragent.Parse(r.UserAgent())
	return extractDevice(userAgent)
}

func extractDevice(userAgent useragent.UserAgent) string {
	return strings.Join([]string{deviceType(userAgent), userAgent.Device, userAgent.OS, userAgent.Name}, "/")
}

func deviceType(ua useragent.UserAgent) string {
	switch {
	case ua.Desktop:
		return "Desktop"
	case ua.Mobile:
		return "Mobile"
	case ua.Tablet:
		return "Tablet"
	default:
		return "Unknown device"
	}
}
