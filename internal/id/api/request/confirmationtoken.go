package request

import (
	"context"
	"net/http"
	"net/url"

	"github.com/zhuboris/never-expires/internal/id/api/endpoint"
	"github.com/zhuboris/never-expires/internal/shared/httpmux"
)

const tokenQueryName = "token"

func confirmEmailURL(ctx context.Context, r *http.Request, authService AuthService, email string) (string, error) {
	token, err := authService.AddEmailConfirmationToken(ctx, email)
	if err != nil {
		return "", err
	}

	result, err := urlWithToken(r, endpoint.ConfirmEmail, token)
	if err != nil {
		return "", err
	}

	return result, nil
}

func urlWithToken(r *http.Request, endpoint, token string) (string, error) {
	link, err := httpmux.EndpointURL(endpoint, r)
	if err != nil {
		return "", err
	}

	return addTokenToURL(link, token)
}

func addTokenToURL(rawURL, token string) (string, error) {
	link, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	query := link.Query()
	query.Add(tokenQueryName, token)
	link.RawQuery = query.Encode()
	return link.String(), nil
}
