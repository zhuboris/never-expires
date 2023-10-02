package main

import (
	"context"
	"errors"
	"log"
	"time"

	"go.uber.org/zap"

	"github.com/zhuboris/never-expires/internal/id/usr"
	"github.com/zhuboris/never-expires/internal/id/usr/deletionnotifier"
	"github.com/zhuboris/never-expires/internal/reminder"
	"github.com/zhuboris/never-expires/internal/reminder/reminderusrdeleter"
	"github.com/zhuboris/never-expires/internal/shared/postgresql"
	"github.com/zhuboris/never-expires/internal/shared/zaplog"
)

const allowedInitDurationForInit = 1 * time.Minute

const (
	userRepoName      = "userRepo"
	reminderRepoName  = "reminderRepo"
	serviceNameLogKey = "service"
	notifierName      = "notifier"
	reminderName      = "reminder"
)

func main() {
	if logger, err := run(); err != nil {
		handleError(logger, err)
	}
}

func run() (*zap.Logger, error) {
	logger, err := zaplog.NewLogger()
	if err != nil {
		return logger, err
	}

	userRepoConfig, err := usr.DBConfig()
	if err != nil {
		return logger, err
	}

	reminderRepoConfig, err := reminder.DBConfig()
	if err != nil {
		return logger, err
	}

	var (
		userPostgresqlConfig     = postgresql.NewNamedConfig(userRepoName, userRepoConfig)
		reminderPostgresqlConfig = postgresql.NewNamedConfig(reminderRepoName, reminderRepoConfig)
	)

	ctx, cancel := context.WithTimeout(context.Background(), allowedInitDurationForInit)
	pools, err := postgresql.MakePoolsAsync(ctx, cancel, userPostgresqlConfig, reminderPostgresqlConfig)
	if err != nil {
		return logger, err
	}

	userRepo, err := deletionnotifier.NewUserPostgresqlRepository(pools[userPostgresqlConfig])
	if err != nil {
		return logger, err
	}

	reminderRepo, err := reminderusrdeleter.NewPostgresqlRepository(pools[reminderPostgresqlConfig])
	if err != nil {
		return logger, err
	}

	var (
		deleteNotifier      = deletionnotifier.New(userRepo, logger.With(zap.String(serviceNameLogKey, notifierName)))
		reminderUserDeleter = reminderusrdeleter.New(reminderRepo, logger.With(zap.String(serviceNameLogKey, reminderName)))
	)

	deleteNotifier.RegisterSubscriber(reminderUserDeleter)
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	err = deleteNotifier.DeleteAll(ctx)
	return logger, err
}

func handleError(logger *zap.Logger, err error) {
	if errors.Is(err, zaplog.ErrFailedToMakeLogger) || logger == nil {
		log.Fatal(err)
	}

	defer logger.Sync()
	logger.Fatal("API is shutdown", zap.Error(err))
}
