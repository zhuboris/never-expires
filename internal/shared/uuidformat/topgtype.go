package uuidformat

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var ErrInvalidUUID = errors.New("invalid uuid")

func StrToPgtype(uuidRaw string) (pgtype.UUID, error) {
	uuidParsed, err := uuid.Parse(uuidRaw)
	if err != nil {
		return pgtype.UUID{}, errors.Join(ErrInvalidUUID, err)
	}

	return pgtype.UUID{Bytes: uuidParsed, Valid: true}, nil
}
