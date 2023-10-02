package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/zhuboris/never-expires/internal/id/api"
	"github.com/zhuboris/never-expires/internal/id/api/request"
	"github.com/zhuboris/never-expires/internal/id/authservice"
	"github.com/zhuboris/never-expires/internal/id/mailing/mailbuilder"
	"github.com/zhuboris/never-expires/internal/id/mailing/mailqueue"
	"github.com/zhuboris/never-expires/internal/id/mailing/rabbitmq"
	"github.com/zhuboris/never-expires/internal/id/session"
	"github.com/zhuboris/never-expires/internal/id/usr"
	"github.com/zhuboris/never-expires/internal/id/usr/oauth/applesignin"
	"github.com/zhuboris/never-expires/internal/id/usr/oauth/googleoauthios"
	"github.com/zhuboris/never-expires/internal/shared/postgresql"
	"github.com/zhuboris/never-expires/internal/shared/prometheusexporter"
	"github.com/zhuboris/never-expires/internal/shared/runapi"
	"github.com/zhuboris/never-expires/internal/shared/zaplog"
)

const authServerListenAddrKey = "AUTH_SERVER_ADDRESS"

const (
	apiLogKey              = "api"
	rabbitMQLogKey         = "rabbitMQProducer"
	apiName                = "authAPI"
	prometheusExporterName = "prometheusExporter"
	rabbitMQName           = "rabbitMQ"
)

const (
	userRepoName     = "userRepo"
	sessionsRepoName = "sessionsRepo"
)

const allowedInitDurationForInit = 1 * time.Minute

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

	config, err := usr.DBConfig()
	if err != nil {
		return logger, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), allowedInitDurationForInit)
	defer cancel()

	authDBPool, err := postgresql.MakePool(ctx, config)
	if err != nil {
		return logger, err
	}

	userRepo, err := usr.NewPostgresqlRepository(authDBPool)
	if err != nil {
		return logger, err
	}

	sessionsRepo, err := session.NewPostgresqlRepository(authDBPool)
	if err != nil {
		return logger, err
	}

	oAuthGoogleIOSService, err := googleoauthios.NewService()
	if err != nil {
		return logger, fmt.Errorf("google oAuth service for iOS creation failed, %w", err)
	}

	appleSignInService, err := applesignin.NewService()
	if err != nil {
		return logger, fmt.Errorf("apple signIn service creation failed, %w", err)
	}

	prometheusExporter := prometheusexporter.New()
	userStatusMetric, err := prometheusExporter.NewServiceStatus(userRepoName)
	if err != nil {
		return logger, fmt.Errorf("user repo status metric is was not registered, %w", err)
	}

	sessionsStatusMetric, err := prometheusExporter.NewServiceStatus(sessionsRepoName)
	if err != nil {
		return logger, fmt.Errorf("session repo status metric is was not registered, %w", err)
	}

	var (
		userService    = usr.NewService(userRepo, oAuthGoogleIOSService, appleSignInService, userStatusMetric)
		sessionService = session.NewService(sessionsRepo, sessionsStatusMetric)
		authService    = authservice.New(userService, sessionService)
	)

	mailBuilder, err := mailbuilder.New()
	if err != nil {
		return nil, fmt.Errorf("mail builder creation failed, %w", err)
	}

	rabbitMQProducer, err := rabbitmq.NewProducer(mailqueue.QueueName, logger.With(zap.String(apiLogKey, rabbitMQLogKey)))
	if err != nil {
		return nil, fmt.Errorf("rabbitMQ producer creation failed, %w", err)
	}

	emailQueue := mailqueue.NewEmailQueue(rabbitMQProducer)

	logger = logger.With(zap.String(apiLogKey, apiName))
	request.InitEmailSender(mailBuilder, emailQueue, logger)

	var (
		authAddr   = os.Getenv(authServerListenAddrKey)
		authServer = api.NewServer(authAddr, authService, logger, prometheusExporter)
	)

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	toRun := map[string]runapi.Runner{
		apiName:                authServer,
		prometheusExporterName: prometheusExporter,
		rabbitMQName:           rabbitMQProducer,
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
