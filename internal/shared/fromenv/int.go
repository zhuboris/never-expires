package fromenv

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

func Int(envKey string) (int, error) {
	value := os.Getenv(envKey)
	result, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.Join(errFailedGetEnv(envKey, value), err)
	}

	return result, nil
}

func errFailedGetEnv(key, value string) error {
	return fmt.Errorf("cannot get needed valid data from env: env key: %q, env value: %q", key, value)
}
