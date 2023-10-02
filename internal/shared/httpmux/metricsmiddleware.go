package httpmux

import (
	"context"
	"net/http"
)

func (m *Mux) metricsRegisterMiddleware(endpoint string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := withEndpoint(r.Context(), endpoint)
		r = r.WithContext(ctx)

		handler(w, r)
	}
}

type endpointKeyType struct{}

var endpointKey endpointKeyType

func withEndpoint(ctx context.Context, endpoint string) context.Context {
	return context.WithValue(ctx, endpointKey, endpoint)
}

func endpoint(ctx context.Context) (string, bool) {
	value := ctx.Value(endpointKey)
	result, ok := value.(string)

	return result, ok
}
