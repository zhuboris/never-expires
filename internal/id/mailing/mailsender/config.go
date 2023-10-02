package mailsender

import (
	"errors"
	"fmt"
	"net/smtp"
	"os"

	"github.com/zhuboris/never-expires/internal/shared/fromenv"
)

type Config struct {
	username string
	password string
	host     string
	port     int
	from     string
}

type ConfigEnvs struct {
	UsernameKey string
	PasswordKey string
	HostKey     string
	PortKey     string
	FromKey     string
}

func NewConfig(envs ConfigEnvs) (Config, error) {
	port, err := fromenv.Int(envs.PortKey)
	if err != nil {
		return Config{}, err
	}

	config := Config{
		username: os.Getenv(envs.UsernameKey),
		password: os.Getenv(envs.PasswordKey),
		host:     os.Getenv(envs.HostKey),
		port:     port,
		from:     os.Getenv(envs.FromKey),
	}

	if !config.isValid() {
		return Config{}, errors.New("cannot get needed valid data from env")
	}

	return config, nil
}

func (c Config) auth() smtp.Auth {
	return smtp.PlainAuth("", c.username, c.password, c.host)
}

func (c Config) serverName() string {
	return fmt.Sprintf("%s:%d", c.host, c.port)
}

func (c Config) isValid() bool {
	return c.username != "" && c.password != "" && c.host != "" && c.port != 0 && c.from != ""
}
