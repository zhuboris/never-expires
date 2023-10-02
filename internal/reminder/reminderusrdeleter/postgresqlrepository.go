package reminderusrdeleter

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zhuboris/never-expires/internal/shared/postgresql"
)

type PostgresqlRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresqlRepository(pool *pgxpool.Pool) (*PostgresqlRepository, error) {
	if pool == nil {
		return nil, postgresql.ErrPoolInitRequired
	}

	return &PostgresqlRepository{
		pool: pool,
	}, nil
}
func (r PostgresqlRepository) deleteUsersData(ctx context.Context, ids []string) error {
	const sql string = `
		WITH deleted_default_storage AS (
		    DELETE FROM users_default_storages
			WHERE user_id = ANY($1)
		),
		deleted_storages AS (
		    DELETE FROM storages
			WHERE owner_id = ANY($1)
		),
		deleted_ios_devices AS (
		    DELETE FROM ios_devices
			WHERE user_id = ANY($1)
		)
		DELETE FROM private_types_of_items
		WHERE user_id = ANY($1);
	`

	_, err := r.pool.Exec(ctx, sql, ids)
	return postgresql.HandleQueryErr(err)
}
