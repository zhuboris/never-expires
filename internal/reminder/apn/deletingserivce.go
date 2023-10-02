package apn

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
)

type (
	badTokensGettingRepo interface {
		popBadTokens(ctx context.Context, limit int64) ([]string, error)
	}
	allTokensStorageRepo interface {
		removeDeviceTokens(ctx context.Context, tokens []string) error
	}
)

type InactiveTokensDeletingService struct {
	badTokensRepo     badTokensGettingRepo
	tokensStorageRepo allTokensStorageRepo
	logger            *zap.Logger
}

func NewInactiveTokensDeletingService(tokensStorageRepo allTokensStorageRepo, badTokensRepo badTokensGettingRepo, logger *zap.Logger) *InactiveTokensDeletingService {
	return &InactiveTokensDeletingService{
		badTokensRepo:     badTokensRepo,
		tokensStorageRepo: tokensStorageRepo,
		logger:            logger,
	}
}

func (s InactiveTokensDeletingService) RunWithCtx(ctx context.Context) error {
	const (
		batchSize    = 100
		timeoutValue = 2 * time.Minute
	)

	var err error

loop:
	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			break loop
		default:
			if err = s.deleteBatchWithTimeout(ctx, timeoutValue, batchSize); err != nil {
				break loop
			}
		}
	}

	s.logShutdown(err)
	return err
}

func (s InactiveTokensDeletingService) deleteBatchWithTimeout(ctx context.Context, timeoutValue time.Duration, batchSize int64) error {
	ctx, cancel := context.WithTimeout(ctx, timeoutValue)
	defer cancel()

	return s.deleteBatch(ctx, batchSize)
}

func (s InactiveTokensDeletingService) deleteBatch(ctx context.Context, size int64) error {
	tokens, err := s.badTokensRepo.popBadTokens(ctx, size)
	if err != nil {
		s.logDeletingFail(err)
		return err
	}

	countToDelete := len(tokens)
	if countToDelete == 0 {
		s.logDeletingFail(err)
		return errors.New("no saved bad tokens left")
	}

	err = s.tokensStorageRepo.removeDeviceTokens(ctx, tokens)
	if err != nil {
		s.logDeletingFail(err)
		return err
	}

	s.logDeletingSuccess(countToDelete)
	return nil
}

func (s InactiveTokensDeletingService) logShutdown(err error) {
	s.logger.Error("Error processing deleting", zap.Error(err))
}

func (s InactiveTokensDeletingService) logDeletingFail(err error) {
	s.logger.Error("Failed to delete a batch", zap.Error(err))
}

func (s InactiveTokensDeletingService) logDeletingSuccess(deletedCount int) {
	s.logger.Info("Successfully deleted a batch", zap.Int("deletedCount", deletedCount))
}
