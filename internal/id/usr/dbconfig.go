package usr

import "github.com/zhuboris/never-expires/internal/shared/postgresql"

func DBConfig() (postgresql.Config, error) {
	const (
		dbUsernameKey = "AUTH_PG_USERNAME"
		dbPasswordKey = "AUTH_PG_PASSWORD"
		dbHostKey     = "AUTH_PG_HOST"
		dbPortKey     = "AUTH_PG_PORT"
		dbNamedKey    = "AUTH_PG_DBNAME"
	)

	return postgresql.NewConfig(postgresql.ConfigEnvsAddress{
		UsernameKey: dbUsernameKey,
		PasswordKey: dbPasswordKey,
		HostKey:     dbHostKey,
		PortKey:     dbPortKey,
		DBNameKey:   dbNamedKey,
	})
}
