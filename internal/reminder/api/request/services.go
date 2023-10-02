package request

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zhuboris/never-expires/internal/reminder/item"
	"github.com/zhuboris/never-expires/internal/reminder/storage"
)

type (
	StorageService interface {
		All(ctx context.Context) ([]*storage.Storage, error)
		Add(ctx context.Context, name string) (*storage.Storage, error)
		AddWithID(ctx context.Context, storageID pgtype.UUID, name string) (*storage.Storage, error)
		Update(ctx context.Context, updated storage.Storage) (*storage.Storage, error)
		Clear(ctx context.Context, storageID pgtype.UUID) error
		Delete(ctx context.Context, storageID pgtype.UUID) error
		Status(ctx context.Context) error
	}
	ItemService interface {
		ByID(ctx context.Context, id pgtype.UUID) (*item.Item, error)
		All(ctx context.Context, filters ...item.Filter) (*item.Items, error)
		Add(ctx context.Context, storageID pgtype.UUID, toAdd item.Item) (*item.Item, error)
		Update(ctx context.Context, updatedItem item.Item) (*item.Item, error)
		Delete(ctx context.Context, itemID pgtype.UUID) error
		Copy(ctx context.Context, toCopy item.ToCopy) (*item.Item, error)
		SearchSavedNames(ctx context.Context, toSearch string, limit int) (*[]string, error)
		Status(ctx context.Context) error
	}
	ApnsService interface {
		AddDeviceToken(ctx context.Context, token string) error
		Status(ctx context.Context) error
	}
)
