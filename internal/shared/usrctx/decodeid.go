package usrctx

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zhuboris/never-expires/internal/id/usr"
)

type ID struct{}

func (d ID) Decode(ctx context.Context) (pgtype.UUID, error) {
	id, err := usr.ID(ctx)
	if err != nil {
		err = errors.New("context do not contain userID")
	}

	return id, err
}
