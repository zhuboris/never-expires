package item

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zhuboris/never-expires/internal/reminder/genuuid"
	"github.com/zhuboris/never-expires/internal/reminder/queryerr"
	"github.com/zhuboris/never-expires/internal/shared/postgresql"
	"github.com/zhuboris/never-expires/internal/shared/servicechecker"
	"github.com/zhuboris/never-expires/internal/shared/usrctx"
)

type (
	repository interface {
		byID(ctx context.Context, userID, id pgtype.UUID) (*Item, error)
		all(ctx context.Context, userID pgtype.UUID, filters ...Filter) (*Items, error)
		add(ctx context.Context, userID, storageID pgtype.UUID, toAdd Item) (isStorageExist, isAdded bool, newItem *Item, err error)
		update(ctx context.Context, userID pgtype.UUID, item Item) (bool, error)
		delete(ctx context.Context, userID, itemID pgtype.UUID) (bool, error)
		copy(ctx context.Context, userID pgtype.UUID, toCopy ToCopy) (isItemExistExist, isCopied bool, newItem *Item, err error)
		searchSavedNames(ctx context.Context, userID pgtype.UUID, searchPattern string, limit int) (*[]string, error)
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

func (s Service) ByID(ctx context.Context, id pgtype.UUID) (*Item, error) {
	userID, err := s.usrID.Decode(ctx)
	if err != nil {
		return nil, err
	}

	item, err := s.repo.byID(ctx, userID, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrItemNotExists
	}

	return item, err
}

func (s Service) All(ctx context.Context, filters ...Filter) (*Items, error) {
	userID, err := s.usrID.Decode(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.all(ctx, userID, filters...)
}

func (s Service) Add(ctx context.Context, storageID pgtype.UUID, toAdd Item) (*Item, error) {
	userID, err := s.usrID.Decode(ctx)
	if err != nil {
		return nil, err
	}

	genuuid.MakeValidIfNeeded(&toAdd.ID)
	isStorageExist, isDone, newItem, err := s.repo.add(ctx, userID, storageID, toAdd)
	if err != nil {
		return nil, err
	}

	if !isStorageExist {
		return nil, queryerr.ErrStorageNotExists
	}

	if !isDone {
		return nil, postgresql.ErrAddedDuplicateOfUnique
	}

	return newItem, nil
}

func (s Service) Update(ctx context.Context, updatedItem Item) (*Item, error) {
	userID, err := s.usrID.Decode(ctx)
	if err != nil {
		return nil, err
	}

	oldItem, err := s.repo.byID(ctx, userID, updatedItem.ID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrItemNotExists
	}

	if err != nil {
		return nil, err
	}

	if updatedItem.isEqual(oldItem) {
		return oldItem, nil
	}

	updatedItem.updateExpirationDate(*oldItem)
	updatedItem.DateAdded = oldItem.DateAdded
	_, err = s.repo.update(ctx, userID, updatedItem)

	return &updatedItem, err
}

func (s Service) Delete(ctx context.Context, itemID pgtype.UUID) error {
	userID, err := s.usrID.Decode(ctx)
	if err != nil {
		return err
	}

	_, err = s.repo.delete(ctx, userID, itemID)
	return err
}

func (s Service) Copy(ctx context.Context, toCopy ToCopy) (*Item, error) {
	userID, err := s.usrID.Decode(ctx)
	if err != nil {
		return nil, err
	}

	genuuid.MakeValidIfNeeded(&toCopy.NewID)
	isOriginalExist, isDone, newItem, err := s.repo.copy(ctx, userID, toCopy)
	if err != nil {
		return nil, err
	}

	if !isOriginalExist {
		return nil, ErrItemNotExists
	}

	if !isDone {
		return nil, postgresql.ErrAddedDuplicateOfUnique
	}

	return newItem, nil
}

func (s Service) SearchSavedNames(ctx context.Context, toSearch string, limit int) (*[]string, error) {
	const regexPattern = `(^|\s)%s(.*)`

	userID, err := s.usrID.Decode(ctx)
	if err != nil {
		return nil, err
	}

	pattern := fmt.Sprintf(regexPattern, toSearch)
	return s.repo.searchSavedNames(ctx, userID, pattern, limit)
}

func (s Service) Status(ctx context.Context) error {
	return servicechecker.Ping(ctx, s.repo, s.statusMetric, "itemRepository")
}
