package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/zhuboris/never-expires/internal/id/mailing/mailqueue"
	"github.com/zhuboris/never-expires/internal/id/mailing/mailsender"
	"github.com/zhuboris/never-expires/internal/id/mailing/rabbitmq"
	"github.com/zhuboris/never-expires/internal/shared/zaplog"
)

const (
	smptUsername = "SMPT_USERNAME"
	smptPassword = "SMPT_PASSWORD"
	smptHost     = "SMPT_HOST"
	smptPort     = "SMPT_PORT"
	smptFrom     = "SMPT_FROM"
)

const (
	serviceNameLogKey    = "service"
	smptClientName       = "smptClient"
	rabbitMQConsumerName = "rabbitMQConsumer"
)
const allowedInitDurationForInit = 1 * time.Minute

const numberOfWorkers = 5

func main() {
	logger, err := zaplog.NewLogger()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		logger.Fatal("API is shutdown")
		logger.Sync()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runWorkersPool(ctx, numberOfWorkers, logger)
}

func runWorkersPool(ctx context.Context, workersCount int, logger *zap.Logger) {
	var wg sync.WaitGroup
	wg.Add(workersCount)

	for i := 0; i < workersCount; i++ {
		workerLogger := logger.With(zap.Int("workerNumber", i))

		go func() {
			defer wg.Done()

			err := runWorker(ctx, workerLogger)
			if err != nil {
				workerLogger.Error("worker stopped with error", zap.Error(err))
			}
		}()
	}

	wg.Wait()
}

func runWorker(cancelCtx context.Context, logger *zap.Logger) error {
	logger = logger.With(zap.String("api", "mailSender"))
	smtpConfig, err := mailsender.NewConfig(mailsender.ConfigEnvs{
		UsernameKey: smptUsername,
		PasswordKey: smptPassword,
		HostKey:     smptHost,
		PortKey:     smptPort,
		FromKey:     smptFrom,
	})
	if err != nil {
		return err
	}

	smptInitCtx, cancel := context.WithTimeout(context.Background(), allowedInitDurationForInit)
	defer cancel()

	smtpClient, err := mailsender.NewSMTPClient(smptInitCtx, smtpConfig, logger.With(zap.String(serviceNameLogKey, smptClientName)))
	if err != nil {
		return fmt.Errorf("mail client creation failed, %w", err)
	}

	defer smtpClient.Quit()

	rabbitMQConsumer, err := rabbitmq.NewConsumer(mailqueue.QueueName, logger.With(zap.String(serviceNameLogKey, rabbitMQConsumerName)))
	if err != nil {
		return fmt.Errorf("rabbitMQ consumer client creation failed, %w", err)
	}

	worker := mailsender.NewWorker(smtpClient, rabbitMQConsumer)
	return worker.DoWork(cancelCtx)
}
