package apn

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/token"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/zhuboris/never-expires/internal/shared/appleconfig"
)

const (
	hoursToExpire = 25
	payloadFormat = `{"aps":{"alert":{"title-loc-key":"%s","title-loc-args":["%s"],"loc-key":"%s","loc-args":["%d"]}},"sound": "default"}`
	titleKey      = "PUSH_EXPIRING_APN_TITLE"
	bodyKey       = "PUSH_EXPIRING_APN_BODY"
	collapseID    = "items_expiring"
	priority      = 10
)

const deviceTokenLogKey = "deviceToken"

type (
	notificationDataRepo interface {
		notifications(ctx context.Context, dataCh chan<- notificationData) error
	}
	badTokensSavingRepo interface {
		addBadToken(ctx context.Context, token string) error
	}
)

type SenderService struct {
	bundleID           string
	notificationsRepo  notificationDataRepo
	inactiveTokensRepo badTokensSavingRepo
	client             *apns2.Client
	logger             *zap.Logger
	exporter           metricsExporter
	counter            sendingCounter
}

func NewSenderService(notificationsRepo notificationDataRepo, inactiveTokensRepo badTokensSavingRepo, logger *zap.Logger, exporter metricsExporter) (*SenderService, error) {
	config, err := appleconfig.New()
	if err != nil {
		return nil, err
	}

	client, err := makeClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create apn client: %w", err)
	}

	counter, err := exporter.NewAttemptsCounter("apns")
	if err != nil {
		return nil, err
	}

	return &SenderService{
		bundleID:           config.ClientID(),
		notificationsRepo:  notificationsRepo,
		inactiveTokensRepo: inactiveTokensRepo,
		logger:             logger,
		client:             client,
		exporter:           exporter,
		counter:            counter,
	}, nil
}

func (s *SenderService) RunWithCtx(ctx context.Context) error {
	defer s.saveMetricsToFile()
	if err := s.recordTimeMetric(startTimeName); err != nil {
		return err
	}

	defer s.recordTimeMetric(finishTimeName)

	dataCh := make(chan notificationData)
	go func(ctx context.Context) {
		err := s.notificationsRepo.notifications(ctx, dataCh)
		close(dataCh)

		if err != nil {
			s.logger.Error("Query error", zap.Error(err))
		}
	}(ctx)

	return s.notifyAll(ctx, dataCh)
}

func (s *SenderService) notifyAll(ctx context.Context, dataCh <-chan notificationData) error {
	const workersCount = 20

	wg := new(sync.WaitGroup)
	wg.Add(workersCount)
	for i := 0; i < workersCount; i++ {
		i := i
		go s.runWorker(ctx, dataCh, wg, i)
	}

	wg.Wait()
	if err := ctx.Err(); err != nil {
		return err
	}

	return errors.New("apns sender is shutdown after it finished all job")
}

func (s *SenderService) runWorker(ctx context.Context, dataCh <-chan notificationData, wg *sync.WaitGroup, workerNumber int) {
	const (
		timeoutValue       = time.Second * 10
		workerNumberLogKey = "workerNumber"
	)

	s.logger.Info("worker started", zap.Int(workerNumberLogKey, workerNumber))
	defer s.logger.Info("worker finished", zap.Int(workerNumberLogKey, workerNumber))

	defer wg.Done()
	for {
		select {
		case data, isOpened := <-dataCh:
			if !isOpened {
				return
			}

			ctx, cancel := context.WithTimeout(ctx, timeoutValue)
			s.notify(ctx, data)
			cancel()
		case <-ctx.Done():
			return
		}
	}
}

func (s *SenderService) notify(ctx context.Context, data notificationData) {
	var (
		payload      = fmt.Sprintf(payloadFormat, titleKey, data.ClosestExpiringItemName, bodyKey, data.ExpiringSoonItemsCount)
		notification = &apns2.Notification{
			CollapseID:  collapseID,
			DeviceToken: data.DeviceToken,
			Topic:       s.bundleID,
			Expiration:  time.Now().Add(time.Hour * hoursToExpire),
			Priority:    priority,
			Payload:     payload,
			PushType:    apns2.PushTypeAlert,
		}
	)

	resp, err := s.client.PushWithContext(ctx, notification)

	isSuccess := isSendWithSuccess(err, resp)
	if !isSuccess && isTokenInactive(resp) {
		s.storeInactiveToken(ctx, data.DeviceToken)
	}

	s.incrementCounter(isSuccess)
	s.logResponse(err, resp, data.DeviceToken)
}

func (s *SenderService) storeInactiveToken(ctx context.Context, token string) {
	err := s.inactiveTokensRepo.addBadToken(ctx, token)
	s.logSavingBadToken(err, token)
}

func (s *SenderService) logResponse(err error, resp *apns2.Response, deviceToken string) {
	msg := "Send with success"
	logLevel := zapcore.InfoLevel
	if !isSendWithSuccess(err, resp) {
		msg = "Failed to send"
		logLevel = zapcore.ErrorLevel
	}

	s.logger.Log(logLevel, msg, zap.String(deviceTokenLogKey, deviceToken), zap.Any("response", resp), zap.Error(err))
}

func (s *SenderService) logSavingBadToken(err error, deviceToken string) {
	msg := "Bad token is saved"
	logLevel := zapcore.InfoLevel
	if err != nil {
		msg = "Failed to save bad token"
		logLevel = zapcore.ErrorLevel
	}

	s.logger.Log(logLevel, msg, zap.String(deviceTokenLogKey, deviceToken), zap.Error(err))
}

func makeClient(config appleconfig.Config) (*apns2.Client, error) {
	authToken, err := tokenFromConfig(config)
	if err != nil {
		return nil, err
	}

	client := apns2.NewTokenClient(authToken).Production()
	return client, nil
}

func tokenFromConfig(config appleconfig.Config) (*token.Token, error) {
	keyBytes := []byte(config.PrivateKey())
	authKey, err := token.AuthKeyFromBytes(keyBytes)
	if err != nil {
		return nil, err
	}

	authToken := &token.Token{
		AuthKey: authKey,
		KeyID:   config.KeyID(),
		TeamID:  config.TeamID(),
	}

	return authToken, nil
}

func isSendWithSuccess(err error, resp *apns2.Response) bool {
	return err == nil && resp.Sent()
}
