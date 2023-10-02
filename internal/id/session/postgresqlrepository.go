package session

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
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

func (r PostgresqlRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func (r PostgresqlRepository) add(ctx context.Context, session Session) (pgtype.UUID, error) {
	const sql = `
		INSERT INTO sessions (user_id, device, refresh_jwt)
		VALUES ($1, $2, $3)
		RETURNING id;
	`

	var id pgtype.UUID
	err := r.pool.QueryRow(ctx, sql, session.UserID, session.Device, session.RefreshJWT).Scan(&id)

	return id, postgresql.HandleQueryErr(err)
}

func (r PostgresqlRepository) contains(ctx context.Context, session Session) error {
	const sql = `
		SELECT EXISTS(
		    SELECT 1 FROM sessions
		    WHERE id = $1
			AND user_id = $2
			AND refresh_jwt = $3
			AND is_active = true
		) AS contains;		
	`

	var result bool
	err := r.pool.
		QueryRow(ctx, sql, session.ID, session.UserID, session.RefreshJWT).
		Scan(&result)

	if result || err != nil {
		return postgresql.HandleQueryErr(err)
	}

	return postgresql.ErrNoMatches
}

func (r PostgresqlRepository) deactivate(ctx context.Context, opts option) error {
	const deactivateQuery = `
		UPDATE sessions
		SET is_active = false
	`

	id, queryOption, err := opts(ctx)
	if err != nil {
		return err
	}

	sql := deactivateQuery + queryOption
	_, err = r.pool.Exec(ctx, sql, id)

	return postgresql.HandleQueryErr(err)
}

func (r PostgresqlRepository) isDeviceNewWhenUserHadSessionsBefore(ctx context.Context, session Session) (bool, error) {
	const sql = `
		WITH past_sessions AS(
		    SELECT device FROM sessions
		    WHERE user_id = $1
		    AND id != $2
		), device_usage_check AS (
		    SELECT NOT EXISTS (
		        SELECT 1 FROM past_sessions
		   		WHERE device ILIKE $3
		    ) AS device_never_used
		)
		SELECT (device_never_used AND EXISTS (SELECT 1 FROM past_sessions)) AS is_device_new
		FROM device_usage_check;
	`

	var isNew bool
	err := r.pool.QueryRow(ctx, sql, session.UserID, session.ID, session.Device).
		Scan(&isNew)
	if err != nil {
		err = postgresql.HandleQueryErr(err)
	}

	return isNew, err
}
