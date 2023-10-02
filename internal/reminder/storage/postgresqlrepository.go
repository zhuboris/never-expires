package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zhuboris/never-expires/internal/shared/postgresql"
)

type PostgresqlRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresqlRepository(pool *pgxpool.Pool) *PostgresqlRepository {
	return &PostgresqlRepository{
		pool: pool,
	}
}

func (r PostgresqlRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func (r PostgresqlRepository) allByOwnerID(ctx context.Context, ownerID pgtype.UUID, defaultNames [3]string) ([]*Storage, error) {
	const sqlToGet = `
		WITH defaults AS (
			INSERT INTO storages (name, owner_id) 
			VALUES ($2, $1), ($3, $1), ($4, $1)
			ON CONFLICT (name, owner_id) DO NOTHING
			        
			RETURNING id, name, owner_id
		), saved_default AS (
			INSERT INTO users_default_storages (user_id, storage_id) 
			SELECT owner_id, id
    		FROM defaults 
    		WHERE name = $2
    		ON CONFLICT DO NOTHING
    		
			RETURNING storage_id
		)
		SELECT
    		id,
    		name,
    		(SELECT COUNT(*) FROM items WHERE storage_id = s.id) AS items_contain,
    		EXISTS (
       			SELECT 1 FROM users_default_storages ds
        		WHERE s.owner_id = ds.user_id
          		AND s.id = ds.storage_id
        		)
        		OR EXISTS
            		(SELECT 1 FROM saved_default sd
            		WHERE s.id = sd.storage_id
        		) AS is_default
		FROM (
    		(SELECT id, name, owner_id FROM storages
     		WHERE owner_id = $1)
    		UNION ALL
    		(SELECT id, name, owner_id FROM defaults)
		)  AS s
		ORDER BY items_contain DESC;
	`

	rows, err := r.pool.Query(ctx, sqlToGet, ownerID, defaultNames[0], defaultNames[1], defaultNames[2])
	if err != nil {
		return nil, postgresql.HandleQueryErr(err)
	}

	storages := make([]*Storage, 0)
	for rows.Next() {
		storage := new(Storage)
		err := rows.Scan(&storage.ID, &storage.Name, &storage.ItemsCount, &storage.IsDefault)
		if err != nil {
			return nil, postgresql.HandleQueryErr(err)
		}

		storages = append(storages, storage)
	}

	return storages, nil
}

func (r PostgresqlRepository) add(ctx context.Context, toAdd Storage, ownerID pgtype.UUID) (bool, *Storage, error) {
	const sql = `
		WITH inserted_storage AS ( 
			INSERT INTO storages (id, name, owner_id)
			VALUES ($1, $2, $3)
			ON CONFLICT (name, owner_id) DO NOTHING
			    
			RETURNING id, name
		) 
		SELECT 
		    EXISTS (SELECT 1 FROM inserted_storage) AS is_added , 
		    s.id, 
		    s.name
		FROM (SELECT 1) as dummy
		LEFT JOIN inserted_storage s ON TRUE;
	`

	var (
		isAdded bool
		storage = new(Entity)
	)

	err := r.pool.QueryRow(ctx, sql, toAdd.ID, toAdd.Name, ownerID).
		Scan(&isAdded, &storage.ID, &storage.Name)
	if err != nil {
		err = postgresql.CheckErrorForUniqueViolation(err)
		return false, nil, postgresql.HandleQueryErr(err)
	}

	return isAdded, storage.storage(), nil
}

func (r PostgresqlRepository) update(ctx context.Context, updated Storage, ownerID pgtype.UUID) (bool, *Storage, error) {
	const sql = `
		WITH updated AS (
			UPDATE storages
			SET name = $1
			WHERE id = $2
			AND owner_id = $3
		
			RETURNING *
		)
		SELECT 
			EXISTS (SELECT 1 FROM updated) AS is_updated,
		    u.id,
		    u.name,
		    (SELECT COUNT(*) FROM items WHERE storage_id = u.id) AS items_contains,
			EXISTS (
		    	SELECT 1 FROM users_default_storages ds 
		    	WHERE u.owner_id = ds.user_id 
		    	AND u.id = ds.storage_id
		    ) AS is_default
		FROM (SELECT 1) as dummy
		LEFT JOIN updated u ON TRUE;
	`

	var (
		isUpdated bool
		storage   = new(Entity)
	)
	err := r.pool.QueryRow(ctx, sql, updated.Name, updated.ID, ownerID).
		Scan(&isUpdated, &storage.ID, &storage.Name, &storage.ItemsCount, &storage.IsDefault)
	if err != nil {
		err = postgresql.CheckErrorForUniqueViolation(err)
		return false, nil, postgresql.HandleQueryErr(err)
	}

	return isUpdated, storage.storage(), nil
}

func (r PostgresqlRepository) clear(ctx context.Context, storageID, ownerID pgtype.UUID) (bool, error) {
	const sql = `
		WITH user_storage AS (
			SELECT id FROM storages
			WHERE owner_id = $1
			AND id = $2
		)
		DELETE FROM items 
		WHERE storage_id = (SELECT id FROM  user_storage)
		
		RETURNING EXISTS (SELECT 1 FROM user_storage) AS is_own_storage;
	`

	var isOwn bool
	err := r.pool.QueryRow(ctx, sql, ownerID, storageID).
		Scan(&isOwn)
	if err != nil {
		return false, postgresql.HandleQueryErr(err)
	}

	return isOwn, nil
}

func (r PostgresqlRepository) delete(ctx context.Context, storageID, ownerID pgtype.UUID) (bool, error) {
	const sql = `
		WITH deleted AS(
			DELETE FROM storages
			WHERE owner_id = $1
			AND id = $2
		
			RETURNING 1
		)
		SELECT EXISTS (SELECT 1 FROM deleted AS is_deleted);
	`

	var isDeleted bool
	err := r.pool.QueryRow(ctx, sql, ownerID, storageID).
		Scan(&isDeleted)
	if err != nil {
		return false, postgresql.HandleQueryErr(err)
	}

	return isDeleted, nil
}

func (r PostgresqlRepository) isForbiddenToDelete(ctx context.Context, storageID, ownerID pgtype.UUID) (bool, error) {
	const sql = `
		SELECT EXISTS (
		    SELECT 1 FROM users_default_storages
		    WHERE user_id = $1
		    AND storage_id = $2
		)
	`

	var isDefaultStorage bool
	err := r.pool.QueryRow(ctx, sql, ownerID, storageID).
		Scan(&isDefaultStorage)
	if err != nil {
		return false, postgresql.HandleQueryErr(err)
	}

	return isDefaultStorage, nil
}
