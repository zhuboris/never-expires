package reminder

import "github.com/zhuboris/never-expires/internal/shared/postgresql"

func DBConfig() (postgresql.Config, error) {
	const (
		dbUsernameKey = "REMINDER_POSTGRESQL_USER"
		dbPasswordKey = "REMINDER_POSTGRESQL_PASSWORD"
		dbHostKey     = "REMINDER_POSTGRESQL_HOST"
		dbPortKey     = "REMINDER_POSTGRESQL_PORT"
		dbNamedKey    = "REMINDER_POSTGRESQL_DB"
	)

	return postgresql.NewConfig(postgresql.ConfigEnvsAddress{
		UsernameKey: dbUsernameKey,
		PasswordKey: dbPasswordKey,
		HostKey:     dbHostKey,
		PortKey:     dbPortKey,
		DBNameKey:   dbNamedKey,
	})
}
