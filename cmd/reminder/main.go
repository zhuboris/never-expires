package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/zhuboris/never-expires/internal/reminder"
	"github.com/zhuboris/never-expires/internal/reminder/api"
	"github.com/zhuboris/never-expires/internal/reminder/apn"
	"github.com/zhuboris/never-expires/internal/reminder/item"
	"github.com/zhuboris/never-expires/internal/reminder/storage"
	"github.com/zhuboris/never-expires/internal/shared/postgresql"
	"github.com/zhuboris/never-expires/internal/shared/prometheusexporter"
	"github.com/zhuboris/never-expires/internal/shared/runapi"
	"github.com/zhuboris/never-expires/internal/shared/zaplog"
)

const serverListenAddrKey = "SERVER_ADDRESS"

const (
	apiName                = "reminderAPI"
	prometheusExporterName = "prometheusExporter"
)

const allowedInitDurationForInit = 1 * time.Minute

func main() {
	if logger, err := run(); err != nil {
		handleError(logger, err)
	}
}

func run() (*zap.Logger, error) {
	const (
		reminderRepoName = "reminderRepo"
		itemsRepoName    = "itemsRepo"
		storagesRepoName = "storagesRepo"
		apnsRepoName     = "apnsRepo"
	)

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

	reminderDBPool, err := postgresql.MakePool(ctx, config)
	if err != nil {
		return logger, err
	}

	var (
		storagesRepo = storage.NewPostgresqlRepository(reminderDBPool)
		itemsRepo    = item.NewPostgresqlRepository(reminderDBPool)
		apnsRepo     = apn.NewPostgresqlRepository(reminderDBPool)
	)

	logger = logger.With(zap.String("service", "reminder"))

	prometheusExporter := prometheusexporter.New()
	itemsStatusMetric, err := prometheusExporter.NewServiceStatus(itemsRepoName)
	if err != nil {
		return logger, fmt.Errorf("items repo status metric is was not registered, %w", err)
	}

	storagesStatusMetric, err := prometheusExporter.NewServiceStatus(storagesRepoName)
	if err != nil {
		return logger, fmt.Errorf("storages repo status metric is was not registered, %w", err)
	}

	apnsStatusMetric, err := prometheusExporter.NewServiceStatus(apnsRepoName)
	if err != nil {
		return logger, fmt.Errorf("apns repo status metric is was not registered, %w", err)
	}

	var (
		serverAddr      = os.Getenv(serverListenAddrKey)
		itemsService    = item.NewService(itemsRepo, itemsStatusMetric)
		storagesService = storage.NewService(storagesRepo, storagesStatusMetric)
		apnsService     = apn.NewDeviceService(apnsRepo, apnsStatusMetric)
	)

	server := api.NewServer(serverAddr, storagesService, itemsService, apnsService, logger, prometheusExporter)

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	toRun := map[string]runapi.Runner{
		apiName:                server,
		prometheusExporterName: prometheusExporter,
	}

	logger.Info(fmt.Sprintf("Starting APIs: %s", runapi.RunnersList(toRun)))
	err = runapi.AllAsync(ctx, cancel, toRun)
	return logger, err
}

func handleError(logger *zap.Logger, err error) {
	if errors.Is(err, zaplog.ErrFailedToMakeLogger) || logger == nil {
		log.Fatal(err)
	}

	defer logger.Sync()
	logger.Fatal("API is shutdown", zap.Error(err))
}
