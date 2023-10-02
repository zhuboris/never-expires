package deletionnotifier

import (
	"context"
	"errors"
	"sync"

	"go.uber.org/zap"
)

type UserDeleter interface {
	DeleteUsers(ctx context.Context, usersIDs []string)
}

type userRepository interface {
	popIDsToDelete(ctx context.Context, limit int) ([]string, error)
}

type UserDeletionNotifier struct {
	repo     userRepository
	deleters []UserDeleter
	logger   *zap.Logger
}

func New(repo userRepository, logger *zap.Logger) *UserDeletionNotifier {
	return &UserDeletionNotifier{
		repo:   repo,
		logger: logger,
	}
}

func (n *UserDeletionNotifier) RegisterSubscriber(subscriber UserDeleter) {
	n.deleters = append(n.deleters, subscriber)
}

func (n *UserDeletionNotifier) DeleteAll(ctx context.Context) error {
	const batchSize = 100

	n.logger.Info("Deleter is up")

	var err error
	for err == nil {
		err = n.deleteBatch(ctx, batchSize)
	}

	n.logger.Error("Deleter is shutdown", zap.Error(err))
	return err
}

func (n *UserDeletionNotifier) deleteBatch(ctx context.Context, batchSize int) error {
	if n.deleters == nil {
		return nil
	}

	toDelete, err := n.repo.popIDsToDelete(ctx, batchSize)
	if err != nil {
		return err
	}

	if len(toDelete) == 0 {
		return errors.New("nothing left to delete")
	}

	n.doDeletions(ctx, toDelete)
	return nil
}

func (n *UserDeletionNotifier) doDeletions(ctx context.Context, toDelete []string) {
	var wg sync.WaitGroup
	deletersNumber := len(n.deleters)
	wg.Add(deletersNumber)

	for i := range n.deleters {
		i := i

		go func(ctx context.Context) {
			defer wg.Done()
			n.deleters[i].DeleteUsers(ctx, toDelete)
		}(ctx)
	}

	wg.Wait()
}
