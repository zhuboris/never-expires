package dislocation

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type location struct {
	Country string `json:"country"`
	City    string `json:"city"`
}

func (l location) String() string {
	if l.Country == "" {
		return ""
	}

	if l.City == "" {
		return l.Country
	}

	return l.Country + ", " + l.City
}

const clientTimeoutValue = time.Second * 10

var httpClient = &http.Client{
	Timeout: clientTimeoutValue,
}

var errFailedGettingLocation = errors.New("failed to get location")

func ByIP(ip string) (string, error) {
	url := fmt.Sprintf("https://ipapi.co/%s/json/", ip)

	resp, err := httpClient.Get(url)
	if err != nil {
		return "", errors.Join(errFailedGettingLocation, err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Join(errFailedGettingLocation, err)
	}

	ipLocation := new(location)
	err = json.Unmarshal(body, ipLocation)
	if err != nil || ipLocation.String() == "" {
		return "", errors.Join(errFailedGettingLocation, err)
	}

	return ipLocation.String(), nil
}
