package usr

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zhuboris/never-expires/internal/id/tkn"
)

type userIDKeyType struct{}

var userIDKey userIDKeyType

func ID(ctx context.Context) (pgtype.UUID, error) {
	if userID, ok := ctx.Value(userIDKey).(pgtype.UUID); ok && userID.Valid {
		return userID, nil
	}

	return pgtype.UUID{}, tkn.ErrUnauthorized
}
func WithUserID(ctx context.Context, id pgtype.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}
