package mailbuilder

import (
	"strings"
	"time"

	"github.com/zhuboris/never-expires/internal/id/dislocation"

	"github.com/mileusna/useragent"
)

type NotificationData struct {
	userAgent   useragent.UserAgent
	ip          string
	requestTime time.Time
}

func NewNotificationData(userAgent useragent.UserAgent, ip string, requestTime time.Time) NotificationData {
	return NotificationData{
		userAgent:   userAgent,
		ip:          ip,
		requestTime: requestTime,
	}
}

func (nd NotificationData) device() string {
	if nd.userAgent.Device != "" {
		return nd.userAgent.Device
	}

	device := strings.Join([]string{nd.userAgent.OS, nd.userAgent.Name}, " ")
	return strings.TrimSpace(device)
}

func (nd NotificationData) ipWithLocation() string {
	location, err := dislocation.ByIP(nd.ip)
	if err != nil {
		return nd.ip
	}

	return strings.Join([]string{nd.ip, location}, " ")
}

func (nd NotificationData) formattedTime() string {
	timeUTC := nd.requestTime.UTC()
	return timeUTC.Format("02 January 2006, 15:04 MST")
}
