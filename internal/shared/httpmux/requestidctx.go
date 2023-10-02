package httpmux

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type requestIDKeyType struct{}

var requestIDKey requestIDKeyType

func requestID(ctx context.Context) (uuid.UUID, error) {
	if id, ok := ctx.Value(requestIDKey).(uuid.UUID); ok && (id != uuid.Nil) {
		return id, nil
	}

	return uuid.Nil, errors.New("context does not contain a valid request id")
}

func withRequestID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}
