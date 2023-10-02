package postgresql

import (
	"errors"
	"fmt"
	"os"

	"github.com/zhuboris/never-expires/internal/shared/fromenv"
)

type ConfigEnvsAddress struct {
	UsernameKey string
	PasswordKey string
	HostKey     string
	PortKey     string
	DBNameKey   string
}

type Config struct {
	username string
	password string
	host     string
	port     int
	dbName   string
}

func NewConfig(envs ConfigEnvsAddress) (Config, error) {
	port, err := fromenv.Int(envs.PortKey)
	if err != nil {
		return Config{}, err
	}

	config := Config{
		username: os.Getenv(envs.UsernameKey),
		password: os.Getenv(envs.PasswordKey),
		host:     os.Getenv(envs.HostKey),
		port:     port,
		dbName:   os.Getenv(envs.DBNameKey),
	}

	if !config.isValid() {
		return Config{}, errors.New("cannot get needed valid data from env")
	}

	return config, nil
}

func (c Config) ConnString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", c.username, c.password, c.host, c.port, c.dbName)
}

func (c Config) Info() string {
	return fmt.Sprintf("postgersql config for db: db name %q, username %q", c.dbName, c.username)
}

func (c Config) isValid() bool {
	return c.username != "" && c.password != "" && c.host != "" && c.port != 0 && c.dbName != ""
}
