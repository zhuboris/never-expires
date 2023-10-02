package test

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

type PostgresConfig struct {
	Username string
	Password string
	Host     string
	Port     int
	DBName   string
}

func (c PostgresConfig) connStringToPostgres() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d", c.Username, c.Password, c.Host, c.Port)
}

func (c PostgresConfig) connStringToDB() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", c.Username, c.Password, c.Host, c.Port, c.DBName)
}

func SetupDatabase(t *testing.T, config PostgresConfig, pathToSQLFile string) *pgxpool.Pool {
	t.Helper()
	connStr := config.connStringToPostgres()
	pool, err := pgxpool.New(context.Background(), connStr)
	require.NoErrorf(t, err, "Unable to connect to postgres: %v\n", err)

	_, err = pool.Exec(context.Background(), "CREATE DATABASE "+config.DBName)
	require.NoErrorf(t, err, "Unable to create test database: %v\n", err)
	pool.Close()

	connStr = config.connStringToDB()
	pool, err = pgxpool.New(context.Background(), connStr)
	require.NoErrorf(t, err, "Unable to connect to database: %v\n", err)

	executeSQLFile(t, pool, pathToSQLFile)
	return pool
}

func DropDatabase(t *testing.T, pool *pgxpool.Pool, config PostgresConfig) {
	t.Helper()
	pool.Close()

	connStr := config.connStringToPostgres()
	pool, err := pgxpool.New(context.Background(), connStr)
	require.NoErrorf(t, err, "Unable to connect to database: %v\n", err)

	defer pool.Close()

	_, err = pool.Exec(context.Background(), "DROP DATABASE "+config.DBName)
	require.NoErrorf(t, err, "Unable to drop test database: %v\n", err)
}

func executeSQLFile(t *testing.T, pool *pgxpool.Pool, path string) {
	t.Helper()
	if path == "" {
		t.Logf("SQL file not provided for %q test", t.Name())
		return
	}

	file, err := os.Open(path)
	require.NoErrorf(t, err, "Unable to open sql file: %v\n", err)

	defer file.Close()

	content, err := io.ReadAll(file)
	require.NoErrorf(t, err, "Reading sql file error: %v\n", err)

	sql := string(content)
	_, err = pool.Exec(context.Background(), sql)
	require.NoErrorf(t, err, "Executing sql file error: %v\n", err)
}
