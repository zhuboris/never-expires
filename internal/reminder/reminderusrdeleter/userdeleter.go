package reminderusrdeleter

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type repository interface {
	deleteUsersData(ctx context.Context, ids []string) error
}

type UserDeleter struct {
	repo   repository
	logger *zap.Logger
}

func New(repo repository, logger *zap.Logger) *UserDeleter {
	return &UserDeleter{
		repo:   repo,
		logger: logger,
	}
}

func (d UserDeleter) DeleteUsers(ctx context.Context, usersIDs []string) {
	err := d.repo.deleteUsersData(ctx, usersIDs)
	d.logDeleting(err, usersIDs)
}

func (d UserDeleter) logDeleting(err error, usersIDs []string) {
	msg := "Successfully deleted"
	logLvl := zapcore.InfoLevel
	if err != nil {
		msg = "Failed to delete"
		logLvl = zapcore.ErrorLevel
	}

	d.logger.Log(logLvl, msg, zap.Error(err), zap.Strings("ids", usersIDs))
}
