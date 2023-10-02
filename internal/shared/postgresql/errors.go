package postgresql

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrPoolInitRequired       = errors.New("passed pool is nil")
	ErrNoMatches              = errors.New("no matches")
	ErrAddedDuplicateOfUnique = errors.New("attempt to add duplicate to unique column")
)

func HandleQueryErr(err error) error {
	if err != nil {
		return fmt.Errorf("unexpected postgresql db error: %w", err)
	}

	return nil
}

func CheckErrorForUniqueViolation(err error) error {
	const uniqueViolationCode = "23505"

	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == uniqueViolationCode {
		err = errors.Join(ErrAddedDuplicateOfUnique, err)
	}

	return err
}
