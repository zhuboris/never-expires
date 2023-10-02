package deletionnotifier

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zhuboris/never-expires/internal/shared/postgresql"
)

type UserPostgresqlRepository struct {
	pool *pgxpool.Pool
}

func NewUserPostgresqlRepository(pool *pgxpool.Pool) (*UserPostgresqlRepository, error) {
	if pool == nil {
		return nil, postgresql.ErrPoolInitRequired
	}

	return &UserPostgresqlRepository{
		pool: pool,
	}, nil
}

func (r UserPostgresqlRepository) popIDsToDelete(ctx context.Context, limit int) ([]string, error) {
	const sql = `
		WITH batch AS (
		    SELECT id
		    FROM  users_to_delete
		    LIMIT $1
		)
		DELETE FROM users_to_delete
		WHERE id IN (SELECT id FROM batch)
		
		RETURNING id;
	`

	rows, err := r.pool.Query(ctx, sql, limit)
	if err != nil {
		return nil, postgresql.HandleQueryErr(err)
	}

	var ids []string
	for rows.Next() {
		var id string
		if scanError := rows.Scan(&id); scanError != nil {
			err = errors.Join(scanError, err)
			continue
		}

		ids = append(ids, id)
	}

	return ids, postgresql.HandleQueryErr(err)
}
