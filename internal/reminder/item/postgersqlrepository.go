package item

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
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

func (r PostgresqlRepository) byID(ctx context.Context, userID, id pgtype.UUID) (*Item, error) {
	const sql = `
		WITH users_items AS(
		    SELECT i.id FROM items i
		    LEFT JOIN storages s
		    ON i.storage_id = s.id
		    WHERE s.owner_id = $1		    
		)
		SELECT
			name,
			is_opened,
			best_before,
			expiration_date,
			hours_after_opening,
			added_date,
			note
		FROM items_info
		WHERE id = $2
		AND id IN (SELECT id FROM users_items);
	`

	item := new(Item)
	err := r.pool.QueryRow(ctx, sql, userID, id).
		Scan(&item.Name, &item.IsOpened, &item.BestBefore, &item.ExpirationDate, &item.HoursAfterOpening, &item.DateAdded, &item.Note)
	return item, err
}

func (r PostgresqlRepository) all(ctx context.Context, userID pgtype.UUID, filters ...Filter) (*Items, error) {
	const sqlFormat = `
		SELECT
		    ii.id,
			ii.name,
			ii.is_opened,
			ii.best_before,
			ii.expiration_date,
			ii.hours_after_opening,
			ii.added_date,
			ii.note
		FROM items_info ii
		LEFT JOIN items i on i.id = ii.id
		WHERE i.storage_id IN (SELECT id FROM storages WHERE owner_id = $1)
		%s
		ORDER BY expiration_date, name DESC;		
	`
	const nextQueryParamIndex = 2

	var filterQuery string
	params := make([]any, 0, len(filters))
	if len(filters) != 0 {
		filterQuery, params = makeQueryFilteringPath(nextQueryParamIndex, filters)
	}

	sql := fmt.Sprintf(sqlFormat, filterQuery)
	rows, err := r.pool.Query(ctx, sql, append([]any{userID}, params...)...)
	if err != nil {
		return nil, err
	}

	items := make(Items, 0)
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Name, &item.IsOpened, &item.BestBefore, &item.ExpirationDate, &item.HoursAfterOpening, &item.DateAdded, &item.Note)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return &items, nil
}

func (r PostgresqlRepository) add(ctx context.Context, userID, storageID pgtype.UUID, toAdd Item) (isStorageExist, isAdded bool, newItem *Item, err error) {
	const sql = `
		WITH shared_name AS (
    		SELECT name
    		FROM shared_types_of_items
    		WHERE lower(name) = lower($1)
		), private_name AS(
			INSERT INTO private_types_of_items (name, user_id)
			SELECT $1, $2
			WHERE NOT EXISTS (SELECT * FROM shared_name)
			ON CONFLICT (lower(name), user_id) DO NOTHING
		), existing_storage AS(
		    SELECT id FROM storages
		    WHERE owner_id = $2
		    AND id = $3
		), new_item AS (
		    INSERT INTO items (id, storage_id)
		    SELECT $4, id 
		    FROM existing_storage 
		    
		    RETURNING id
		), inserted_item AS (
		    INSERT INTO items_info (id, name, is_opened, added_date, best_before, hours_after_opening, note)
			SELECT id, $1, $5, $6, $7, $8, $9
			FROM new_item
		    RETURNING id, name, is_opened, best_before, expiration_date, hours_after_opening, added_date, note
		)
		SELECT  
		    EXISTS(SELECT 1 FROM existing_storage) AS storage_exists,
		    EXISTS(SELECT 1 FROM new_item) AS is_added,
		    (SELECT id FROM inserted_item),
		  	(SELECT expiration_date FROM inserted_item),
		    (SELECT added_date FROM inserted_item);
	`

	scannedItem := newFromItem(toAdd)
	err = r.pool.QueryRow(ctx, sql, toAdd.Name, userID, storageID, toAdd.ID, toAdd.IsOpened, toAdd.DateAdded, toAdd.BestBefore, toAdd.HoursAfterOpening, toAdd.Note).
		Scan(&isStorageExist, &isAdded, &scannedItem.ID, &scannedItem.ExpirationDate, &scannedItem.DateAdded)
	return isStorageExist, isAdded, scannedItem.item(), err
}

func (r PostgresqlRepository) update(ctx context.Context, userID pgtype.UUID, item Item) (bool, error) {
	const sql = `
		WITH users_items AS(
		    SELECT i.id FROM items i
		    LEFT JOIN storages s
		    ON i.storage_id = s.id
		    WHERE s.owner_id = $1
		), updated AS (
		    UPDATE items_info
			SET 
			    is_opened = $2,
		        best_before = $3,
			    expiration_date = $4,
			    hours_after_opening = $5,
		        note = $6
		    WHERE id IN (SELECT id FROM users_items)
			AND id = $7

		    RETURNING 1
		)
		SELECT EXISTS (SELECT  1 FROM updated) AS is_updated
	`

	var isUpdated bool
	err := r.pool.QueryRow(ctx, sql, userID, item.IsOpened, item.BestBefore, item.ExpirationDate, item.HoursAfterOpening, item.Note, item.ID).
		Scan(&isUpdated)

	return isUpdated, err
}

func (r PostgresqlRepository) delete(ctx context.Context, userID, itemID pgtype.UUID) (bool, error) {
	const sql = `
		WITH users_items AS(
		    SELECT i.id FROM items i
		    LEFT JOIN storages s
		    ON i.storage_id = s.id
		    WHERE s.owner_id = $1		    
		), deleted AS(
			DELETE FROM items i
			USING users_items ui
			WHERE i.id = ui.id
			AND i.id = $2
		
			RETURNING 1
		)
		SELECT EXISTS (SELECT  1 FROM deleted) AS is_deleted;
	`

	var isDeleted bool
	err := r.pool.QueryRow(ctx, sql, userID, itemID).
		Scan(&isDeleted)

	return isDeleted, err
}

func (r PostgresqlRepository) copy(ctx context.Context, userID pgtype.UUID, toCopy ToCopy) (isItemExistExist, isCopied bool, newItem *Item, err error) {
	const sql = `
		WITH existing_item AS(
		    SELECT ii.*, i.storage_id FROM items i
		    LEFT JOIN storages s
		    ON i.storage_id = s.id
		    LEFT JOIN items_info ii on i.id = ii.id
		    WHERE s.owner_id = $1
		    AND i.id = $2
		), new_item AS (
		    INSERT INTO items (id, storage_id)
		    SELECT $3, storage_id 
		    FROM existing_item
		    
		    RETURNING id
		), inserted_item AS (
			INSERT INTO items_info (id, name, is_opened, added_date, best_before, expiration_date, hours_after_opening, note)
		    SELECT ni.id AS new_id, name, is_opened, $4, best_before, expiration_date, hours_after_opening, note
		    FROM existing_item AS ei, new_item AS ni
			
			RETURNING id, name, is_opened, best_before, expiration_date, hours_after_opening, added_date, note
		)
		SELECT 
		    EXISTS(SELECT 1 FROM existing_item) AS storage_exists,
		    EXISTS(SELECT 1 FROM new_item) AS is_added,
		    (SELECT id FROM inserted_item),
		    (SELECT name FROM inserted_item),
		    (SELECT is_opened FROM inserted_item),
		    (SELECT best_before FROM inserted_item),
		    (SELECT expiration_date FROM inserted_item),
		    (SELECT hours_after_opening FROM inserted_item),
		    (SELECT added_date FROM inserted_item),
		    (SELECT note FROM inserted_item);
	`

	scannedItem := new(Entity)
	err = r.pool.QueryRow(ctx, sql, userID, toCopy.OriginalID, toCopy.NewID, toCopy.DateAdded).
		Scan(
			&isItemExistExist,
			&isCopied,
			&scannedItem.ID,
			&scannedItem.Name,
			&scannedItem.IsOpened,
			&scannedItem.BestBefore,
			&scannedItem.ExpirationDate,
			&scannedItem.HoursAfterOpening,
			&scannedItem.DateAdded,
			&scannedItem.Note,
		)
	return isItemExistExist, isCopied, scannedItem.item(), err
}

func (r PostgresqlRepository) searchSavedNames(ctx context.Context, userID pgtype.UUID, searchPattern string, limit int) (*[]string, error) {
	const sql = `
		WITH all_available_names AS (
		    SELECT name, 1 AS sort_weight FROM private_types_of_items
			WHERE user_id = $1
			UNION ALL
		    SELECT name, 2 AS sort_order FROM shared_types_of_items	
		)
		SELECT name FROM all_available_names
		WHERE name ~* $2 
		AND lower(name) != lower($2)
		ORDER BY sort_weight, length(name), name
		LIMIT $3;
	`

	if limit <= 0 {
		return &[]string{}, nil
	}

	rows, err := r.pool.Query(ctx, sql, userID, searchPattern, limit)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, limit)
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return nil, err
		}

		result = append(result, name)
	}

	return &result, nil
}
