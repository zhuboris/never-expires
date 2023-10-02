package httpmux

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type URLMakingError struct {
	host     string
	endpoint string
}

func (e URLMakingError) Error() string {
	return fmt.Sprintf("failed getting url path to route, host: %q, endpoint: %q", e.host, e.endpoint)
}

func EndpointURL(endpoint string, r *http.Request) (string, error) {
	path, err := url.JoinPath("https://", r.Host, endpoint)
	if err != nil {
		err = errors.Join(URLMakingError{
			r.Host, endpoint,
		}, err)
	}

	return path, err
}

func RemoveSubdomain(host string) string {
	const domainPathsWithoutSubdomain = 2
	const domainIndex = 1

	parts := strings.Split(host, ".")
	if len(parts) > domainPathsWithoutSubdomain {
		return strings.Join(parts[domainIndex:], ".")
	}

	return host
}
