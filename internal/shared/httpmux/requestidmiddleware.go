package httpmux

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

func requestIDMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if err := checkContext(ctx); err != nil {
			id := makeRequestID(r)
			r = r.WithContext(withRequestID(ctx, id))
		}

		next(w, r)
	}
}

func makeRequestID(r *http.Request) uuid.UUID {
	id, err := tryGetFromHeader(r)
	if err != nil {
		id = uuid.New()
	}

	return id
}

func checkContext(ctx context.Context) error {
	if ctx == nil {
		return errors.New("no context")
	}

	id, _ := ctx.Value(requestIDKey).(string) // the check is skipped because it will happen when parsing uuid after
	_, err := uuid.Parse(id)

	return err
}

func tryGetFromHeader(r *http.Request) (uuid.UUID, error) {
	id := r.Header.Get(requestIDHeader)
	return uuid.Parse(id)
}
