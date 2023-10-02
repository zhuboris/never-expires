package apn

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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

func (r PostgresqlRepository) notifications(ctx context.Context, dataCh chan<- notificationData) error {
	const (
		timeTillExpireToSend = "25 hours"
		timeSinceAddedToSend = "3 hours"
	)

	const sql = `
		WITH expiring_items AS (
    		SELECT ii.name AS name, ii.expiration_date as expiration_date, d.token AS device_token
    		FROM items_info ii
    		INNER JOIN items i ON i.id = ii.id
    		INNER JOIN storages s ON i.storage_id = s.id
    		INNER JOIN ios_devices d ON d.user_id = s.owner_id
    		WHERE ii.expiration_date BETWEEN NOW() AND (NOW() + $1::INTERVAL)
    		AND ii.added_date < (NOW() - $2::INTERVAL)
		)
		SELECT
			device_token,
    		COUNT(name) AS expiring_items,
    		(
        		SELECT name
       		 	FROM expiring_items ei2
        		WHERE ei2.device_token = ei.device_token
       		 	ORDER BY ei2.expiration_date
        		LIMIT 1
    	) AS closest_expiring_item_name
		FROM expiring_items ei
		GROUP BY device_token;
	`

	rows, err := r.pool.Query(ctx, sql, timeTillExpireToSend, timeSinceAddedToSend)
	if err != nil {
		return postgresql.HandleQueryErr(err)
	}

	for rows.Next() {
		var data notificationData
		err := rows.Scan(&data.DeviceToken, &data.ExpiringSoonItemsCount, &data.ClosestExpiringItemName)
		if err != nil {
			return postgresql.HandleQueryErr(err)
		}

		dataCh <- data
	}

	return nil
}

func (r PostgresqlRepository) addDeviceToken(ctx context.Context, userID pgtype.UUID, token string) error {
	const sql = `
		INSERT INTO ios_devices (token, user_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING;
	`

	_, err := r.pool.Exec(ctx, sql, token, userID)
	return postgresql.HandleQueryErr(err)
}

func (r PostgresqlRepository) removeDeviceTokens(ctx context.Context, tokens []string) error {
	const sqlFormat = `
		DELETE FROM ios_devices
		WHERE token IN (%s); 
	`

	if len(tokens) == 0 {
		return nil
	}

	placeholder := makePlaceholder(len(tokens))
	args := makeArgs(tokens)

	sql := fmt.Sprintf(sqlFormat, placeholder)
	_, err := r.pool.Exec(ctx, sql, args...)

	return postgresql.HandleQueryErr(err)
}

func makePlaceholder(paramsCount int) string {
	const (
		firstPlaceholderIndex = 1
		dollar                = '$'
		separatorAndDollar    = ", $"
	)

	if paramsCount == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteRune(dollar)
	builder.WriteString(strconv.Itoa(firstPlaceholderIndex))

	for i := 1; i < paramsCount; i++ {
		index := i + firstPlaceholderIndex
		builder.WriteString(separatorAndDollar)
		builder.WriteString(strconv.Itoa(index))
	}

	return builder.String()
}

func makeArgs(input []string) []any {
	result := make([]any, len(input))

	for i, value := range input {
		result[i] = value
	}

	return result
}
