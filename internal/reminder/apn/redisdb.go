package apn

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

type RedisDB struct {
	client *redis.Client
}

func NewRedisDB() (*RedisDB, error) {
	options, err := redisOptions()
	if err != nil {
		return nil, err
	}

	return &RedisDB{
		client: redis.NewClient(options),
	}, nil
}

func redisOptions() (*redis.Options, error) {
	const (
		addrEnvKey     = "REDIS_ADDR"
		usernameEnvKey = "REDIS_USERNAME"
		passwordEnvKey = "REDIS_PASSWORD"
	)

	const errFormat = "key %q is missing in envs"

	address := os.Getenv(addrEnvKey)
	if address == "" {
		return nil, fmt.Errorf(errFormat, addrEnvKey)
	}

	username := os.Getenv(usernameEnvKey)
	if address == "" {
		return nil, fmt.Errorf(errFormat, usernameEnvKey)
	}

	password := os.Getenv(passwordEnvKey)
	if address == "" {
		return nil, fmt.Errorf(errFormat, passwordEnvKey)
	}

	options := &redis.Options{
		Addr:     address,
		Username: username,
		Password: password,
	}
	return options, nil
}

const badTokensRepoKey = "apns_bad_token"

func (r RedisDB) addBadToken(ctx context.Context, token string) error {
	return r.client.
		SAdd(ctx, badTokensRepoKey, token).
		Err()
}

func (r RedisDB) popBadTokens(ctx context.Context, limit int64) ([]string, error) {
	return r.client.
		SPopN(ctx, badTokensRepoKey, limit).
		Result()
}
