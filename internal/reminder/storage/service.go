package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zhuboris/never-expires/internal/reminder/genuuid"
	"github.com/zhuboris/never-expires/internal/reminder/queryerr"
	"github.com/zhuboris/never-expires/internal/reminder/storage/randname"
	"github.com/zhuboris/never-expires/internal/shared/postgresql"
	"github.com/zhuboris/never-expires/internal/shared/servicechecker"
	"github.com/zhuboris/never-expires/internal/shared/usrctx"
)

const (
	undeletableStorageName   = "Fridge"
	secondDefaultStorageName = "Medicines"
	thirdDefaultStorageName  = "Locker"
)

var defaults = [3]string{undeletableStorageName, secondDefaultStorageName, thirdDefaultStorageName}

type (
	repository interface {
		allByOwnerID(ctx context.Context, ownerID pgtype.UUID, defaultNames [3]string) ([]*Storage, error)
		add(ctx context.Context, toAdd Storage, ownerID pgtype.UUID) (bool, *Storage, error)
		update(ctx context.Context, updated Storage, ownerID pgtype.UUID) (bool, *Storage, error)
		clear(ctx context.Context, storageID, ownerID pgtype.UUID) (bool, error)
		delete(ctx context.Context, storageID, ownerID pgtype.UUID) (bool, error)
		isForbiddenToDelete(ctx context.Context, storageID, ownerID pgtype.UUID) (bool, error)
		Ping(ctx context.Context) error
	}
	userIDDecoder interface {
		Decode(ctx context.Context) (pgtype.UUID, error)
	}
)

type Service struct {
	repo         repository
	usrID        userIDDecoder
	statusMetric servicechecker.StatusDisplay
}

func NewService(repo repository, statusDisplay servicechecker.StatusDisplay) *Service {
	return &Service{
		repo:         repo,
		usrID:        usrctx.ID{},
		statusMetric: statusDisplay,
	}
}

func (s Service) All(ctx context.Context) ([]*Storage, error) {
	userID, err := s.usrID.Decode(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.allByOwnerID(ctx, userID, defaults)
}

func (s Service) Add(ctx context.Context, name string) (*Storage, error) {
	return s.AddWithID(ctx, pgtype.UUID{}, name)
}

func (s Service) AddWithID(ctx context.Context, storageID pgtype.UUID, name string) (*Storage, error) {
	const allowedAttempts = 5

	userID, err := s.usrID.Decode(ctx)
	if err != nil {
		return nil, err
	}

	genuuid.MakeValidIfNeeded(&storageID)

	var (
		toAdd = Storage{
			ID:   storageID,
			Name: name,
		}
		isDone     bool
		newStorage *Storage
	)

	if toAdd.Name != "" {
		return s.addWithGivenName(ctx, toAdd, userID)
	}

	isDone, newStorage, err = s.tryAddWithRandomNames(ctx, toAdd, userID, allowedAttempts)
	if isDone || err != nil {
		return newStorage, err
	}

	return s.addWithLongerRandomName(ctx, toAdd, userID)
}

func (s Service) Update(ctx context.Context, updated Storage) (*Storage, error) {
	userID, err := s.usrID.Decode(ctx)
	if err != nil {
		return nil, err
	}

	isUpdated, result, err := s.repo.update(ctx, updated, userID)
	if err != nil {
		return nil, checkForStorageNotUniqueNameError(err)
	}

	return result, errorIfStorageNotExists(isUpdated)
}

func (s Service) Clear(ctx context.Context, storageID pgtype.UUID) error {
	userID, err := s.usrID.Decode(ctx)
	if err != nil {
		return err
	}

	isDone, err := s.repo.clear(ctx, storageID, userID)
	if err != nil {
		return err
	}

	return errorIfStorageNotExists(isDone)
}

func (s Service) Delete(ctx context.Context, storageID pgtype.UUID) error {
	userID, err := s.usrID.Decode(ctx)
	if err != nil {
		return err
	}

	isActionForbidden, err := s.repo.isForbiddenToDelete(ctx, storageID, userID)
	if err != nil {
		return err
	}

	if isActionForbidden {
		return ErrDeletingNotAllowed
	}

	_, err = s.repo.delete(ctx, storageID, userID)
	return err
}

func (s Service) Status(ctx context.Context) error {
	return servicechecker.Ping(ctx, s.repo, s.statusMetric, "sessionsRepository")
}

func (s Service) addWithGivenName(ctx context.Context, toAdd Storage, userID pgtype.UUID) (*Storage, error) {
	isDone, newStorage, err := s.repo.add(ctx, toAdd, userID)
	if err != nil {
		return nil, err
	}

	return newStorage, errorIfStorageNameNotUnique(isDone)
}

func (s Service) tryAddWithRandomNames(ctx context.Context, toAdd Storage, userID pgtype.UUID, allowedAttempts int) (isDone bool, newStorage *Storage, err error) {
	for i := 0; i < allowedAttempts; i++ {
		isDone, newStorage, err := s.addWithRandomName(ctx, toAdd, userID)
		if isDone || err != nil {
			return true, newStorage, err
		}
	}

	return false, nil, nil
}

func (s Service) addWithLongerRandomName(ctx context.Context, toAdd Storage, userID pgtype.UUID) (*Storage, error) {
	toAdd.Name = randname.StorageNameWithDigits()
	isDone, newStorage, err := s.repo.add(ctx, toAdd, userID)
	if err != nil {
		return nil, err
	}

	return newStorage, errorIfStorageNameNotUnique(isDone)
}

func (s Service) addWithRandomName(ctx context.Context, toAdd Storage, userID pgtype.UUID) (bool, *Storage, error) {
	toAdd.Name = randname.StorageName()
	isDone, newStorage, err := s.repo.add(ctx, toAdd, userID)
	if err != nil {
		return false, nil, err
	}

	return isDone, newStorage, nil
}

func checkForStorageNotUniqueNameError(err error) error {
	if errors.Is(err, postgresql.ErrAddedDuplicateOfUnique) {
		err = errors.Join(ErrStorageNameNotUnique, err)
	}

	return err
}

func errorIfStorageNotExists(isExists bool) error {
	if !isExists {
		return queryerr.ErrStorageNotExists
	}

	return nil
}

func errorIfStorageNameNotUnique(isUnique bool) error {
	if !isUnique {
		return ErrStorageNameNotUnique
	}

	return nil
}
