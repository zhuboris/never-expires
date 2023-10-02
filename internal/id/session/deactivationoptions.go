package session

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zhuboris/never-expires/internal/id/usr"
)

type option func(context.Context) (pgtype.UUID, string, error)

const (
	byID     = "WHERE id = $1;"
	byUserID = "WHERE user_id = $1 AND is_active = true;"
)

func byUser() option {
	return func(ctx context.Context) (pgtype.UUID, string, error) {
		id, err := usr.ID(ctx)
		return id, byUserID, err
	}
}

func bySession(sessionID pgtype.UUID) option {
	return func(ctx context.Context) (pgtype.UUID, string, error) {
		return sessionID, byID, nil
	}
}
