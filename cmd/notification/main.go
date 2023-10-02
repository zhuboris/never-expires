package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go.uber.org/zap"

	"github.com/zhuboris/never-expires/internal/reminder"
	"github.com/zhuboris/never-expires/internal/reminder/apn"
	"github.com/zhuboris/never-expires/internal/shared/postgresql"
	"github.com/zhuboris/never-expires/internal/shared/prometheusexporter"
	"github.com/zhuboris/never-expires/internal/shared/runapi"
	"github.com/zhuboris/never-expires/internal/shared/zaplog"
)

const allowedInitDurationForInit = 1 * time.Minute

const (
	apiName                = "apnSender"
	prometheusExporterName = "prometheusExporter"
	serviceLogKey          = "service"
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

	config, err := reminder.DBConfig()
	if err != nil {
		return logger, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), allowedInitDurationForInit)
	defer cancel()

	dbPool, err := postgresql.MakePool(ctx, config)
	if err != nil {
		return logger, err
	}

	apnsRepo := apn.NewPostgresqlRepository(dbPool)
	prometheusExporter := prometheusexporter.New()
	redisDB, err := apn.NewRedisDB()
	if err != nil {
		return nil, fmt.Errorf("redisDB init error: %w", err)
	}

	apnsSender, err := apn.NewSenderService(apnsRepo, redisDB, logger.With(zap.String(serviceLogKey, "APNs_sender")), prometheusExporter)
	if err != nil {
		return logger, err
	}

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	toRun := map[string]runapi.Runner{
		apiName:                apnsSender,
		prometheusExporterName: prometheusExporter,
	}

	logger.Info(fmt.Sprintf("Starting APIs: %s", runapi.RunnersList(toRun)))
	sendingError := runapi.AllAsync(ctx, cancel, toRun)

	cleanupService := apn.NewInactiveTokensDeletingService(apnsRepo, redisDB, logger.With(zap.String(serviceLogKey, "APNs_bad_tokens_deleter")))
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	cleanupError := cleanupService.RunWithCtx(ctx)

	return logger, errors.Join(sendingError, cleanupError)
}

func handleError(logger *zap.Logger, err error) {
	if errors.Is(err, zaplog.ErrFailedToMakeLogger) || logger == nil {
		log.Fatal(err)
	}

	defer logger.Sync()
	logger.Fatal("Service is shutdown", zap.Error(err))
}
