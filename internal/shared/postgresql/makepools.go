package postgresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NamedConfig struct {
	repoName string
	config   Config
}

func NewNamedConfig(repoName string, config Config) NamedConfig {
	return NamedConfig{
		repoName: repoName,
		config:   config,
	}
}

func (c NamedConfig) String() string {
	return fmt.Sprintf("repo %q with %q", c.repoName, c.config.Info())
}

type makePoolResult struct {
	configIndex int
	pool        *pgxpool.Pool
	err         error
}

func MakePoolsAsync(ctx context.Context, cancelInit context.CancelFunc, configs ...NamedConfig) (map[NamedConfig]*pgxpool.Pool, error) {
	defer cancelInit()

	numberToMake := len(configs)
	resultCh := make(chan makePoolResult, numberToMake)

	for i := range configs {
		i := i

		go func() {
			pool, err := MakePool(ctx, configs[i].config)
			resultCh <- makePoolResult{
				configIndex: i,
				pool:        pool,
				err:         err,
			}
		}()
	}

	var (
		err    error
		result = make(map[NamedConfig]*pgxpool.Pool, numberToMake)
	)

	for i := 0; i < numberToMake; i++ {
		makeResult := <-resultCh
		currentConfig := configs[makeResult.configIndex]

		if result[currentConfig] != nil {
			return nil, fmt.Errorf("pool for repo was already added, %q", currentConfig.String())
		}

		result[currentConfig] = makeResult.pool
		if makeResult.err != nil {
			currentError := fmt.Errorf("failed to init %q: %w", currentConfig.String(), makeResult.err)
			err = errors.Join(currentError, err)
		}
	}

	return result, err
}
