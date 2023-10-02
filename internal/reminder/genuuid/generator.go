package genuuid

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func MakeValidIfNeeded(input *pgtype.UUID) {
	if input.Valid {
		return

	}

	input.Bytes = uuid.New()
	input.Valid = true
}
