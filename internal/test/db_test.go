package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDB(t *testing.T) {
	config := PostgresConfig{
		Username: "postgres",
		Password: "12345",
		Host:     "localhost",
		Port:     5432,
		DBName:   "test",
	}

	pool := SetupDatabase(t, config, "./.sql")
	t.Cleanup(func() {
		DropDatabase(t, pool, config)
	})

	sql := `
		INSERT INTO test (name)
		VALUES ($1)
		RETURNING name;
	`
	name := "test"
	var result string
	err := pool.QueryRow(context.Background(), sql, name).
		Scan(&result)
	require.NoErrorf(t, err, "Unable to execute test query: %v\n", err)

	assert.Equal(t, name, result)
}
