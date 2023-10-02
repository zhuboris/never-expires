package postgresql

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zhuboris/never-expires/internal/shared/try"
)

func MakePool(ctx context.Context, config Config) (*pgxpool.Pool, error) {
	const attemptDelay = 5 * time.Second

	newPoolFunc := func() (*pgxpool.Pool, error) {
		return pgxpool.New(ctx, config.ConnString())
	}

	return try.GetWithAttempts(ctx, newPoolFunc, attemptDelay)
}
