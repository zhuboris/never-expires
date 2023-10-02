package item

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/zhuboris/never-expires/internal/reminder/queryerr"
	"github.com/zhuboris/never-expires/internal/shared/postgresql"
	"github.com/zhuboris/never-expires/internal/shared/servicechecker"
)

type statusDisplayMock struct{}

func (m statusDisplayMock) Set(error) {}

var metricsMock statusDisplayMock

func TestNewService(t *testing.T) {
	repo := NewMockrepository(t)
	service := NewService(repo, metricsMock)
	require.Implements(t, (*repository)(nil), repo, "mock is not implement required interface")
	require.Equal(t, repo, service.repo, "mock is not suitable")
}

func TestService_Add(t *testing.T) {
	var (
		notExistingStorageID = pgtype.UUID{
			Bytes: [16]byte{1},
			Valid: true,
		}
		existingStorageID = pgtype.UUID{
			Bytes: [16]byte{2},
			Valid: true,
		}
		invalidStorageID = pgtype.UUID{
			Valid: false,
		}
		alreadyExistingItemID = pgtype.UUID{
			Bytes: [16]byte{3},
			Valid: true,
		}
		invalidItemID = pgtype.UUID{
			Bytes: [16]byte{4},
			Valid: false,
		}
		validItemID = pgtype.UUID{
			Bytes: [16]byte{5},
			Valid: true,
		}
	)

	repoMock := NewMockrepository(t)
	repoMock.EXPECT().
		add(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID pgtype.UUID, storageId pgtype.UUID, item Item) (isStorageExist, isAdded bool, newItem *Item, err error) {
			if !storageId.Valid {
				return false, false, nil, errors.New("invalid storage uuid")
			}

			if storageId == notExistingStorageID {
				return false, false, nil, nil
			}

			if item.ID == alreadyExistingItemID {
				return true, false, nil, nil
			}

			return true, true, &Item{}, nil
		})

	type fields struct {
		repoMock      repository
		idDecoderMock userIDDecoder
	}
	type args struct {
		storageID pgtype.UUID
		toAdd     Item
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		requireError  require.ErrorAssertionFunc
		expectedError error
	}{
		{
			name: "invalid userID",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfInvalidID(t),
			},
			args: args{
				storageID: existingStorageID,
				toAdd:     Item{ID: validItemID},
			},
			requireError: require.Error,
		},
		{
			name: "storage not exists",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfValidID(t),
			},
			args: args{
				storageID: notExistingStorageID,
				toAdd:     Item{ID: validItemID},
			},
			requireError:  require.Error,
			expectedError: queryerr.ErrStorageNotExists,
		},
		{
			name: "invalid storageID",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfValidID(t),
			},
			args: args{
				storageID: invalidStorageID,
				toAdd:     Item{ID: validItemID},
			},
			requireError: require.Error,
		},
		{
			name: "item with given ID is already in db",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfValidID(t),
			},
			args: args{
				storageID: existingStorageID,
				toAdd:     Item{ID: alreadyExistingItemID},
			},
			requireError:  require.Error,
			expectedError: postgresql.ErrAddedDuplicateOfUnique,
		},
		{
			name: "invalid itemID",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfValidID(t),
			},
			args: args{
				storageID: existingStorageID,
				toAdd:     Item{ID: invalidItemID},
			},
			requireError: require.NoError,
		},
		{
			name: "itemID has default value",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfValidID(t),
			},
			args: args{
				storageID: existingStorageID,
				toAdd:     Item{ID: pgtype.UUID{}},
			},
			requireError: require.NoError,
		},
		{
			name: "valid item",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfValidID(t),
			},
			args: args{
				storageID: existingStorageID,
				toAdd:     Item{ID: validItemID},
			},
			requireError: require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(tt.fields.repoMock, metricsMock)
			service.usrID = tt.fields.idDecoderMock

			item, err := service.Add(context.Background(), tt.args.storageID, tt.args.toAdd)

			tt.requireError(t, err)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError, "wrong expected error")
			}

			if err == nil { // if NO error
				assert.NotNil(t, item, "item must exist when no error")
			}
		})
	}
}

func TestService_All(t *testing.T) {
	filters := []Filter{ByName("name"), ByStorageID(pgtype.UUID{Valid: true}), ByDateBefore(time.Now()), ByOpenedStatus( /*isOpened*/ true)}

	repoMock := NewMockrepository(t)
	repoMock.EXPECT().
		all(mock.Anything, mock.Anything).
		Return(&Items{}, nil)
	repoMock.EXPECT().
		all(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&Items{}, nil)

	tests := []struct {
		name          string
		idDecoderMock userIDDecoder
		filters       []Filter
		requireError  require.ErrorAssertionFunc
	}{
		{
			name:          "invalid user id",
			idDecoderMock: newDecoderOfInvalidID(t),
			requireError:  require.Error,
		},
		{
			name:          "valid user id",
			idDecoderMock: newDecoderOfValidID(t),
			requireError:  require.NoError,
		},
		{
			name:          "valid user id with filters",
			idDecoderMock: newDecoderOfValidID(t),
			filters:       filters,
			requireError:  require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(repoMock, metricsMock)
			service.usrID = tt.idDecoderMock

			items, err := service.All(context.Background(), tt.filters...)

			tt.requireError(t, err)
			if err == nil { // if NO error
				assert.NotNil(t, items, "result must be not nil when no error")
			}
		})
	}
}

func TestService_ByID(t *testing.T) {
	var (
		existingID = pgtype.UUID{
			Bytes: [16]byte{1},
			Valid: true,
		}
		notExistingID = pgtype.UUID{
			Bytes: [16]byte{2},
			Valid: true,
		}
		invalidUUID = pgtype.UUID{
			Valid: false,
		}
	)

	repoMock := NewMockrepository(t)
	repoMock.EXPECT().
		byID(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID pgtype.UUID, itemID pgtype.UUID) (*Item, error) {
			if itemID == invalidUUID {
				return nil, errors.New("invalid uuid")
			}

			if itemID == notExistingID {
				return nil, pgx.ErrNoRows
			}

			return &Item{}, nil
		})

	type fields struct {
		repoMock      repository
		idDecoderMock userIDDecoder
	}
	tests := []struct {
		name         string
		fields       fields
		itemID       pgtype.UUID
		requireError require.ErrorAssertionFunc
	}{
		{
			name: "invalid user id",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfInvalidID(t),
			},
			itemID:       existingID,
			requireError: require.Error,
		},
		{
			name: "item not exist",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfValidID(t),
			},
			itemID: notExistingID,
			requireError: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrItemNotExists)
			},
		},
		{
			name: "invalid uuid",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfValidID(t),
			},
			itemID:       invalidUUID,
			requireError: require.Error,
		},
		{
			name: "item exist",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfValidID(t),
			},
			itemID:       existingID,
			requireError: require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(tt.fields.repoMock, metricsMock)
			service.usrID = tt.fields.idDecoderMock

			item, err := service.ByID(context.Background(), tt.itemID)

			tt.requireError(t, err)
			if err == nil { // if NO error
				assert.NotNil(t, item, "item must exist when no error")
			}
		})
	}
}

func TestService_Copy(t *testing.T) {
	var (
		notExistingToCopyID = pgtype.UUID{
			Bytes: [16]byte{1},
			Valid: true,
		}
		invalidToCopyID = pgtype.UUID{
			Bytes: [16]byte{2},
			Valid: false,
		}
		validToCopyID = pgtype.UUID{
			Bytes: [16]byte{3},
			Valid: true,
		}
		alreadyExistingNewID = pgtype.UUID{
			Bytes: [16]byte{4},
			Valid: true,
		}
		invalidNewID = pgtype.UUID{
			Bytes: [16]byte{5},
			Valid: false,
		}
		validNewID = pgtype.UUID{
			Bytes: [16]byte{6},
			Valid: true,
		}
	)

	repoMock := NewMockrepository(t)
	repoMock.EXPECT().
		copy(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID pgtype.UUID, toCopy ToCopy) (isItemExit, isCopied bool, newItem *Item, err error) {
			if isOriginalItemIDValid := toCopy.OriginalID.Valid; !isOriginalItemIDValid {
				return false, false, &Item{}, errors.New("invalid uuid")
			}

			if toCopy.OriginalID == notExistingToCopyID {
				return false, false, &Item{}, nil
			}

			if toCopy.NewID == alreadyExistingNewID {
				return true, false, &Item{}, nil
			}

			return true, true, &Item{ID: toCopy.NewID}, nil
		})

	type fields struct {
		repoMock      repository
		idDecoderMock userIDDecoder
	}
	tests := []struct {
		name          string
		fields        fields
		toCopy        ToCopy
		requireError  require.ErrorAssertionFunc
		expectedError error
	}{
		{
			name: "invalid userID",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfInvalidID(t),
			},
			toCopy: ToCopy{
				OriginalID: validToCopyID,
				NewID:      validNewID,
			},
			requireError: require.Error,
		},
		{
			name: "original item not exists",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfValidID(t),
			},
			toCopy: ToCopy{
				OriginalID: notExistingToCopyID,
				NewID:      validNewID,
			},
			requireError:  require.Error,
			expectedError: ErrItemNotExists,
		},
		{
			name: "invalid originalID",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfValidID(t),
			},
			toCopy: ToCopy{
				OriginalID: invalidToCopyID,
				NewID:      validNewID,
			},
			requireError: require.Error,
		},
		{
			name: "item with given new ID is already in db",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfValidID(t),
			},
			toCopy: ToCopy{
				OriginalID: validToCopyID,
				NewID:      alreadyExistingNewID,
			},
			requireError:  require.Error,
			expectedError: postgresql.ErrAddedDuplicateOfUnique,
		},
		{
			name: "invalid new itemID",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfValidID(t),
			},
			toCopy: ToCopy{
				OriginalID: validToCopyID,
				NewID:      invalidNewID,
			},
			requireError: require.NoError,
		},
		{
			name: "new itemID has default value",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfValidID(t),
			},
			toCopy: ToCopy{
				OriginalID: validToCopyID,
				NewID:      pgtype.UUID{},
			},
			requireError: require.NoError,
		},
		{
			name: "valid operation",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfValidID(t),
			},
			toCopy: ToCopy{
				OriginalID: validToCopyID,
				NewID:      validNewID,
			},
			requireError: require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(tt.fields.repoMock, metricsMock)
			service.usrID = tt.fields.idDecoderMock

			item, err := service.Copy(context.Background(), tt.toCopy)

			tt.requireError(t, err)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError, "wrong expected error")
			}

			if err == nil { // if NO error
				assert.True(t, item.ID.Valid, "when no error item must be not empty and with valid ID")
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	repoMock := NewMockrepository(t)
	repoMock.EXPECT().
		delete(mock.Anything, mock.Anything, mock.Anything).
		Return( /*isDeleted*/ true, nil)

	type fields struct {
		repoMock      repository
		idDecoderMock userIDDecoder
	}
	tests := []struct {
		name         string
		fields       fields
		itemID       pgtype.UUID
		requireError require.ErrorAssertionFunc
	}{
		{
			name: "invalid user id",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfInvalidID(t),
			},
			itemID:       pgtype.UUID{},
			requireError: require.Error,
		},
		{
			name: "any item id, valid user id",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfValidID(t),
			},
			itemID:       pgtype.UUID{},
			requireError: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(tt.fields.repoMock, metricsMock)
			service.usrID = tt.fields.idDecoderMock

			err := service.Delete(context.Background(), tt.itemID)

			tt.requireError(t, err)
		})
	}
}

func TestService_SearchSavedNames(t *testing.T) {
	repoMock := NewMockrepository(t)
	repoMock.EXPECT().
		searchSavedNames(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&[]string{}, nil)

	type args struct {
		toSearch string
		limit    int
	}
	tests := []struct {
		name          string
		idDecoderMock userIDDecoder
		args          args
		want          *[]string
		requireError  require.ErrorAssertionFunc
	}{
		{
			name:          "invalid user id",
			idDecoderMock: newDecoderOfInvalidID(t),
			requireError:  require.Error,
		},
		{
			name:          "limit less than zero",
			idDecoderMock: newDecoderOfValidID(t),
			args: args{
				toSearch: "name",
				limit:    -10,
			},
			requireError: require.NoError,
		},
		{
			name:          "zero limit",
			idDecoderMock: newDecoderOfValidID(t),
			args: args{
				toSearch: "name",
				limit:    0,
			},
			requireError: require.NoError,
		},
		{
			name:          "anything to search, valid user id and limit",
			idDecoderMock: newDecoderOfValidID(t),
			args: args{
				toSearch: "name",
				limit:    10,
			},
			requireError: require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(repoMock, metricsMock)
			service.usrID = tt.idDecoderMock

			result, err := service.SearchSavedNames(context.Background(), tt.args.toSearch, tt.args.limit)

			tt.requireError(t, err)
			if err == nil { // if NO error
				assert.NotNil(t, result, "result must be not nil when no error")
			}
		})
	}
}

func TestService_Update(t *testing.T) {
	var (
		notExistingItemID = pgtype.UUID{
			Bytes: [16]byte{1},
			Valid: true,
		}
		existingIDOfOpenedItem = pgtype.UUID{
			Bytes: [16]byte{2},
			Valid: true,
		}
		existingIDOfClosedItem = pgtype.UUID{
			Bytes: [16]byte{3},
			Valid: true,
		}

		bestBefore        = time.Now().Add(48 * time.Hour)
		updatedBestBefore = time.Now().Add(96 * time.Hour)

		closedItemInDB = Item{
			ID:                existingIDOfClosedItem,
			Name:              "name",
			IsOpened:          false,
			BestBefore:        bestBefore,
			ExpirationDate:    bestBefore,
			HoursAfterOpening: 10,
			DateAdded:         time.Now(),
			Note:              "note",
		}
		openedItemInDB = Item{
			ID:                existingIDOfOpenedItem,
			Name:              "name",
			IsOpened:          true,
			BestBefore:        bestBefore,
			ExpirationDate:    time.Now().Add(10 * time.Hour),
			HoursAfterOpening: 10,
			DateAdded:         time.Now(),
			Note:              "note",
		}
	)

	repoMock := NewMockrepository(t)
	repoMock.EXPECT().
		update(mock.Anything, mock.Anything, mock.Anything).
		Return(true, nil).Maybe()
	repoMock.EXPECT().
		byID(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID pgtype.UUID, itemID pgtype.UUID) (*Item, error) {
			if !itemID.Valid {
				return nil, errors.New("invalid uuid")
			}

			switch itemID {
			case existingIDOfClosedItem:
				return &closedItemInDB, nil
			case existingIDOfOpenedItem:
				return &openedItemInDB, nil
			default:
				return nil, pgx.ErrNoRows
			}
		}).Maybe()

	tests := []struct {
		name          string
		idDecoderMock userIDDecoder
		updatedItem   Item
		wantItem      *Item
		requireError  require.ErrorAssertionFunc
		expectedError error
	}{
		{
			name:          "invalid user id",
			idDecoderMock: newDecoderOfInvalidID(t),
			requireError:  require.Error,
		},
		{
			name:          "invalid itemID",
			idDecoderMock: newDecoderOfValidID(t),
			updatedItem:   Item{ID: pgtype.UUID{Valid: false}},
			requireError:  require.Error,
		},
		{
			name:          "item not exist",
			idDecoderMock: newDecoderOfValidID(t),
			updatedItem:   Item{ID: notExistingItemID},
			requireError:  require.Error,
			expectedError: ErrItemNotExists,
		},
		{
			name:          "item not changed",
			idDecoderMock: newDecoderOfValidID(t),
			updatedItem:   closedItemInDB,
			wantItem:      &closedItemInDB,
			requireError:  require.NoError,
		},
		{
			name: "open item",
			updatedItem: Item{
				ID:                existingIDOfClosedItem,
				Name:              openedItemInDB.Name,
				IsOpened:          true,
				BestBefore:        bestBefore,
				HoursAfterOpening: openedItemInDB.HoursAfterOpening,
				Note:              openedItemInDB.Note,
			},
			wantItem: &Item{
				ID:                closedItemInDB.ID,
				Name:              openedItemInDB.Name,
				IsOpened:          true,
				BestBefore:        openedItemInDB.BestBefore,
				ExpirationDate:    openedItemInDB.ExpirationDate,
				HoursAfterOpening: openedItemInDB.HoursAfterOpening,
				DateAdded:         openedItemInDB.DateAdded,
				Note:              openedItemInDB.Note,
			},
			idDecoderMock: newDecoderOfValidID(t),
			requireError:  require.NoError,
		},
		{
			name: "close item",
			updatedItem: Item{
				ID:                existingIDOfOpenedItem,
				Name:              closedItemInDB.Name,
				IsOpened:          false,
				BestBefore:        bestBefore,
				HoursAfterOpening: closedItemInDB.HoursAfterOpening,
				Note:              closedItemInDB.Note,
			},

			wantItem: &Item{
				ID:                openedItemInDB.ID,
				Name:              closedItemInDB.Name,
				IsOpened:          false,
				BestBefore:        closedItemInDB.BestBefore,
				ExpirationDate:    closedItemInDB.ExpirationDate,
				HoursAfterOpening: closedItemInDB.HoursAfterOpening,
				DateAdded:         closedItemInDB.DateAdded,
				Note:              closedItemInDB.Note,
			},
			idDecoderMock: newDecoderOfValidID(t),
			requireError:  require.NoError,
		},
		{
			name: "change bestBefore",
			updatedItem: Item{
				ID:                existingIDOfClosedItem,
				Name:              closedItemInDB.Name,
				IsOpened:          false,
				BestBefore:        updatedBestBefore,
				HoursAfterOpening: closedItemInDB.HoursAfterOpening,
				Note:              closedItemInDB.Note,
			},
			wantItem: &Item{
				ID:                closedItemInDB.ID,
				Name:              closedItemInDB.Name,
				IsOpened:          false,
				BestBefore:        updatedBestBefore,
				ExpirationDate:    updatedBestBefore,
				HoursAfterOpening: closedItemInDB.HoursAfterOpening,
				DateAdded:         closedItemInDB.DateAdded,
				Note:              closedItemInDB.Note,
			},
			idDecoderMock: newDecoderOfValidID(t),
			requireError:  require.NoError,
		},
		{
			name: "edit string fields",
			updatedItem: Item{
				ID:                existingIDOfClosedItem,
				Name:              "newName",
				IsOpened:          false,
				BestBefore:        closedItemInDB.BestBefore,
				HoursAfterOpening: closedItemInDB.HoursAfterOpening,
				Note:              "",
			},
			wantItem: &Item{
				ID:                closedItemInDB.ID,
				Name:              "newName",
				IsOpened:          false,
				BestBefore:        closedItemInDB.BestBefore,
				ExpirationDate:    closedItemInDB.BestBefore,
				HoursAfterOpening: closedItemInDB.HoursAfterOpening,
				DateAdded:         closedItemInDB.DateAdded,
				Note:              "",
			},
			idDecoderMock: newDecoderOfValidID(t),
			requireError:  require.NoError,
		},
		{
			name: "edit dateAdded does not change it, it is immutable",
			updatedItem: Item{
				ID:                existingIDOfClosedItem,
				Name:              closedItemInDB.Name,
				IsOpened:          closedItemInDB.IsOpened,
				BestBefore:        closedItemInDB.BestBefore,
				HoursAfterOpening: closedItemInDB.HoursAfterOpening,
				DateAdded:         time.Now().Add(time.Hour * 200),
				Note:              closedItemInDB.Note,
			},
			wantItem:      &closedItemInDB,
			idDecoderMock: newDecoderOfValidID(t),
			requireError:  require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(repoMock, metricsMock)
			service.usrID = tt.idDecoderMock

			result, err := service.Update(context.Background(), tt.updatedItem)

			tt.requireError(t, err)
			if err == nil {
				assertItemsEqual(t, *tt.wantItem, *result)
			}
		})
	}
}

func TestService_Status(t *testing.T) {
	tests := []struct {
		name              string
		requireError      require.ErrorAssertionFunc
		expectedErrorType error
	}{
		{
			name:              "unavailable",
			requireError:      require.Error,
			expectedErrorType: new(servicechecker.IsUnavailableError),
		},
		{
			name:         "up",
			requireError: require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockrepository(t)
			repo.EXPECT().
				Ping(mock.Anything).
				Return(tt.expectedErrorType)
			service := NewService(repo, metricsMock)

			resultErr := service.Status(context.Background())

			tt.requireError(t, resultErr)
			if resultErr != nil {
				assert.ErrorAs(t, resultErr, tt.expectedErrorType)
			}
		})
	}
}

func assertItemsEqual(t *testing.T, first, second Item) {
	t.Helper()

	assert.Equal(t, first.Name, second.Name)
	assert.Equal(t, first.ID, second.ID)
	assert.Equal(t, first.Note, second.Note)
	assert.Equal(t, first.IsOpened, second.IsOpened)
	assert.Equal(t, first.HoursAfterOpening, second.HoursAfterOpening)
	assertTimeHaveSameDate(t, first.BestBefore, second.BestBefore)
	assertTimeHaveSameDate(t, first.ExpirationDate, second.ExpirationDate)
	assertTimeHaveSameDate(t, first.DateAdded, second.DateAdded)
}

func assertTimeHaveSameDate(t *testing.T, first, second time.Time) {
	t.Helper()
	firstDate := dateByTime(t, first)
	secondDate := dateByTime(t, second)

	assert.Equalf(t, firstDate, secondDate, "dates are not same: first time = %v, second time = %v", first, second)
}

func dateByTime(t *testing.T, input time.Time) time.Time {
	t.Helper()
	return time.Date(input.Year(), input.Month(), input.Day(), 0, 0, 0, 0, input.Location())
}

func newDecoderOfValidID(t *testing.T) *MockuserIDDecoder {
	t.Helper()
	decoder := NewMockuserIDDecoder(t)
	decoder.EXPECT().
		Decode(mock.Anything).
		Return(pgtype.UUID{Valid: true}, nil)
	return decoder
}

func newDecoderOfInvalidID(t *testing.T) *MockuserIDDecoder {
	t.Helper()
	decoder := NewMockuserIDDecoder(t)
	decoder.EXPECT().
		Decode(mock.Anything).
		Return(pgtype.UUID{Valid: false}, errors.New("context do not contain userID"))
	return decoder
}
