package request

import (
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zhuboris/never-expires/internal/shared/uuidformat"
)

func idFromPath(path, route string) (pgtype.UUID, error) {
	rawID := strings.TrimPrefix(path, route)
	if rawID == "" {
		return pgtype.UUID{}, ErrMissingParam
	}

	return uuidformat.StrToPgtype(rawID)
}
