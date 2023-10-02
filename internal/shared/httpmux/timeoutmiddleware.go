package httpmux

import (
	"context"
	"net/http"
	"time"
)

func timeoutMiddleware(duration time.Duration, f errorHandledFunc) errorHandledFunc {
	if duration == 0 {
		return f
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		if _, ok := ctx.Deadline(); ok {
			return f(w, r)
		}

		ctx, cancel := context.WithTimeout(ctx, duration)
		defer cancel()

		r = r.WithContext(ctx)
		return f(w, r)
	}
}
