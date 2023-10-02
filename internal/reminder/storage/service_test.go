package storage

import (
	"context"
	"errors"
	"regexp"
	"testing"

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
	reservedName := "reserved name"
	repoTestingInputtedName := NewMockrepository(t)
	repoTestingInputtedName.EXPECT().
		add(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, storage Storage, uuid pgtype.UUID) (bool, *Storage, error) {
			if storage.Name == reservedName {
				return false, nil, nil
			}

			return true, &storage, nil
		}).
		Twice()

	repoTestingRepeatSuccess := NewMockrepository(t)
	repoTestingRepeatSuccess.EXPECT().
		add(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, storage Storage, uuid pgtype.UUID) (bool, *Storage, error) {
			if isStorageIDValid := storage.ID.Valid; !isStorageIDValid {
				return false, nil, errors.New("unexpected error, id must be valid")
			}

			containsDigit, err := regexp.MatchString(`\d`, storage.Name)
			if err != nil || !containsDigit {
				return false, nil, nil

			}

			return true, &storage, nil
		}).
		Times(6)

	repoTestingRepeatFail := NewMockrepository(t)
	repoTestingRepeatFail.EXPECT().
		add(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, storage Storage, uuid pgtype.UUID) (bool, *Storage, error) {
			if isStorageIDValid := storage.ID.Valid; !isStorageIDValid {
				return false, nil, errors.New("unexpected error, id must be valid")
			}

			return false, nil, nil
		}).
		Times(6)

	repoWithoutRepeat := NewMockrepository(t)
	repoWithoutRepeat.EXPECT().
		add(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, storage Storage, uuid pgtype.UUID) (bool, *Storage, error) {
			if isStorageIDValid := storage.ID.Valid; !isStorageIDValid {
				return false, nil, errors.New("unexpected error, id must be valid")
			}

			return true, &storage, nil
		}).
		Twice()

	type fields struct {
		repoMock      repository
		idDecoderMock userIDDecoder
	}

	tests := []struct {
		name          string
		fields        fields
		inputtedName  string
		wantResult    bool
		wantError     bool
		expectedError error
	}{
		{
			name: "user not exists",
			fields: fields{
				repoMock:      repoWithoutRepeat,
				idDecoderMock: newDecoderOfInvalidID(t),
			},
			wantResult: false,
			wantError:  true,
		},
		{
			name: "inputted name is reserved",
			fields: fields{
				repoMock:      repoTestingInputtedName,
				idDecoderMock: newDecoderOfValidID(t),
			},
			inputtedName:  reservedName,
			wantResult:    false,
			wantError:     true,
			expectedError: ErrStorageNameNotUnique,
		},
		{
			name: "inputted name is free",
			fields: fields{
				repoMock:      repoTestingInputtedName,
				idDecoderMock: newDecoderOfValidID(t),
			},
			inputtedName: "free name",
			wantResult:   true,
		},
		{
			name: "name is reserved till generation with numbers",
			fields: fields{
				repoMock:      repoTestingRepeatSuccess,
				idDecoderMock: newDecoderOfValidID(t),
			},
			wantResult: true,
		},
		{
			name: "fail to generate name",
			fields: fields{
				repoMock:      repoTestingRepeatFail,
				idDecoderMock: newDecoderOfValidID(t),
			},
			wantResult:    false,
			wantError:     true,
			expectedError: ErrStorageNameNotUnique,
		},
		{
			name: "successful attempt to add with valid id",
			fields: fields{
				repoMock:      repoWithoutRepeat,
				idDecoderMock: newDecoderOfValidID(t),
			},
			wantResult: true,
		},
		{
			name: "successful attempt to add with invalid id",
			fields: fields{
				repoMock:      repoWithoutRepeat,
				idDecoderMock: newDecoderOfValidID(t),
			},
			wantResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(tt.fields.repoMock, metricsMock)
			service.usrID = tt.fields.idDecoderMock

			result, err := service.Add(context.Background(), tt.inputtedName)

			isCorrectExpectedError := (err != nil) == tt.wantError
			require.Truef(t, isCorrectExpectedError, "error = %v, wantErr %v", err, tt.wantError)

			if tt.wantResult {
				assert.NotNilf(t, result, "expected to get result but method did not return it")
				assert.Truef(t, result.ID.Valid, "method must make id valid")
			}

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError, "wrong expected error")
			}
		})
	}
}

func TestService_AddWithID(t *testing.T) {
	var (
		reservedID = pgtype.UUID{
			Bytes: [16]byte{1},
			Valid: true,
		}
		freeID = pgtype.UUID{
			Bytes: [16]byte{2},
			Valid: true,
		}
		invalidID = pgtype.UUID{Valid: false}
	)

	repo := NewMockrepository(t)
	repo.EXPECT().
		add(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, storage Storage, uuid pgtype.UUID) (bool, *Storage, error) {
			if isStorageIDValid := storage.ID.Valid; !isStorageIDValid {
				return false, nil, errors.New("unexpected error, id must be valid")
			}

			if storage.ID == reservedID {
				return false, nil, postgresql.ErrAddedDuplicateOfUnique
			}

			return true, &storage, nil
		})

	type fields struct {
		repoMock      repository
		idDecoderMock userIDDecoder
	}

	tests := []struct {
		name          string
		fields        fields
		storageID     pgtype.UUID
		wantResult    bool
		wantError     bool
		expectedError error
	}{
		{
			name: "user not exists",
			fields: fields{
				repoMock:      repo,
				idDecoderMock: newDecoderOfInvalidID(t),
			},
			storageID:  freeID,
			wantResult: false,
			wantError:  true,
		},
		{
			name: "given id is reserved",
			fields: fields{
				repoMock:      repo,
				idDecoderMock: newDecoderOfValidID(t),
			},
			storageID:     reservedID,
			wantResult:    false,
			wantError:     true,
			expectedError: postgresql.ErrAddedDuplicateOfUnique,
		},
		{
			name: "given id is invalid",
			fields: fields{
				repoMock:      repo,
				idDecoderMock: newDecoderOfValidID(t),
			},
			storageID:  invalidID,
			wantResult: true,
			wantError:  false,
		},
		{
			name: "given id is valid and free",
			fields: fields{
				repoMock:      repo,
				idDecoderMock: newDecoderOfValidID(t),
			},
			storageID:  freeID,
			wantResult: true,
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(tt.fields.repoMock, metricsMock)
			service.usrID = tt.fields.idDecoderMock

			result, err := service.AddWithID(context.Background(), tt.storageID, "")

			isCorrectExpectedError := (err != nil) == tt.wantError
			require.Truef(t, isCorrectExpectedError, "error = %v, wantErr %v", err, tt.wantError)

			if tt.wantResult {
				assert.NotNilf(t, result, "expected to get result but func did not return it")
				assert.Truef(t, result.ID.Valid, "method must make id valid")

				if tt.storageID.Valid {
					assert.Equal(t, tt.storageID, result.ID, "method must not change id if it was valid")
				}
			}

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError, "wrong expected error")
			}
		})
	}
}

func TestService_All(t *testing.T) {
	storages := make([]*Storage, 2)

	repo := NewMockrepository(t)
	repo.EXPECT().
		allByOwnerID(mock.Anything, mock.Anything, mock.Anything).
		Return(storages, nil)

	type fields struct {
		repoMock      repository
		idDecoderMock userIDDecoder
	}

	tests := []struct {
		name      string
		fields    fields
		want      []*Storage
		wantError bool
	}{
		{
			name: "user exists",
			fields: fields{
				repoMock:      repo,
				idDecoderMock: newDecoderOfValidID(t),
			},
			want:      storages,
			wantError: false,
		},
		{
			name: "user not exists",
			fields: fields{
				repoMock:      repo,
				idDecoderMock: newDecoderOfInvalidID(t),
			},
			want:      nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(tt.fields.repoMock, metricsMock)
			service.usrID = tt.fields.idDecoderMock

			result, err := service.All(context.Background())

			isCorrectExpectedError := (err != nil) == tt.wantError
			require.Truef(t, isCorrectExpectedError, "error = %v, wantErr %v", err, tt.wantError)

			if err == nil { // if NO err
				assert.Equal(t, storages, result, "incorrect result")
			}
		})
	}
}

func TestService_Clear(t *testing.T) {
	var (
		notExistingStorageID = pgtype.UUID{
			Bytes: [16]byte{1},
			Valid: true,
		}
		existingStorageID = pgtype.UUID{
			Bytes: [16]byte{2},
			Valid: true,
		}
	)

	repoMock := NewMockrepository(t)
	repoMock.EXPECT().
		clear(mock.Anything, existingStorageID, mock.Anything).
		Return( /*isCleared*/ true, nil)
	repoMock.EXPECT().
		clear(mock.Anything, notExistingStorageID, mock.Anything).
		Return( /*isCleared*/ false, nil)

	type fields struct {
		repoMock      repository
		idDecoderMock userIDDecoder
	}
	tests := []struct {
		name          string
		fields        fields
		storageID     pgtype.UUID
		wantError     bool
		expectedError error
	}{
		{
			name: "invalid user id",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfInvalidID(t),
			},
			storageID: existingStorageID,
			wantError: true,
		},
		{
			name: "clear not existing storage",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfValidID(t),
			},
			storageID:     notExistingStorageID,
			wantError:     true,
			expectedError: queryerr.ErrStorageNotExists,
		},
		{
			name: "clear existing storage",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfValidID(t),
			},
			storageID: existingStorageID,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(tt.fields.repoMock, metricsMock)
			service.usrID = tt.fields.idDecoderMock

			err := service.Clear(context.Background(), tt.storageID)

			isCorrectExpectedError := (err != nil) == tt.wantError
			require.Truef(t, isCorrectExpectedError, "error = %v, wantErr %v", err, tt.wantError)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError, "wrong expected error")
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	var (
		forbiddenToDeleteStorageID = pgtype.UUID{
			Bytes: [16]byte{1},
			Valid: true,
		}
		allowedToDeleteStorageID = pgtype.UUID{
			Bytes: [16]byte{2},
			Valid: true,
		}
		notExistingStorageID = pgtype.UUID{
			Bytes: [16]byte{3},
			Valid: true,
		}
		invalidStorageID = pgtype.UUID{
			Valid: false,
		}
	)

	repoMock := NewMockrepository(t)
	repoMock.EXPECT().
		delete(mock.Anything, mock.Anything, mock.Anything).
		Return( /*isDeleted*/ true, nil)
	repoMock.EXPECT().
		isForbiddenToDelete(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, storageID, userID pgtype.UUID) (bool, error) {
			return storageID == forbiddenToDeleteStorageID, nil
		})

	type fields struct {
		repoMock      repository
		idDecoderMock userIDDecoder
	}
	tests := []struct {
		name          string
		fields        fields
		storageID     pgtype.UUID
		wantError     bool
		expectedError error
	}{
		{
			name: "invalid user id",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfInvalidID(t),
			},
			storageID: allowedToDeleteStorageID,
			wantError: true,
		},
		{
			name: "deletion is forbidden",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfValidID(t),
			},
			storageID:     forbiddenToDeleteStorageID,
			wantError:     true,
			expectedError: ErrDeletingNotAllowed,
		},
		{
			name: "deleting not existing storage",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfValidID(t),
			},
			storageID: notExistingStorageID,
			wantError: false,
		},
		{
			name: "deleting invalid storage",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfValidID(t),
			},
			storageID: invalidStorageID,
			wantError: false,
		},
		{
			name: "deleting existing storage",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfValidID(t),
			},
			storageID: allowedToDeleteStorageID,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(tt.fields.repoMock, metricsMock)
			service.usrID = tt.fields.idDecoderMock

			err := service.Delete(context.Background(), tt.storageID)

			isCorrectExpectedError := (err != nil) == tt.wantError
			require.Truef(t, isCorrectExpectedError, "error = %v, wantErr %v", err, tt.wantError)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError, "wrong expected error")
			}
		})
	}
}

func TestService_Update(t *testing.T) {
	var (
		notExistingStorage = Storage{
			Name: "not existing",
		}
		storageWithNotUniqueName = Storage{
			Name: "not unique",
		}
		existingStorage = Storage{
			Name: "existing",
		}
	)

	repoMock := NewMockrepository(t)
	repoMock.EXPECT().
		update(mock.Anything, existingStorage, mock.Anything).
		Return( /*idUpdated*/ true, &existingStorage, nil)
	repoMock.EXPECT().
		update(mock.Anything, storageWithNotUniqueName, mock.Anything).
		Return( /*idUpdated*/ false, nil, ErrStorageNameNotUnique)
	repoMock.EXPECT().
		update(mock.Anything, notExistingStorage, mock.Anything).
		Return( /*idUpdated*/ false, nil, nil)

	type fields struct {
		repoMock      repository
		idDecoderMock userIDDecoder
	}
	tests := []struct {
		name           string
		fields         fields
		storage        Storage
		wantError      bool
		expectedResult *Storage
		expectedError  error
	}{
		{
			name: "invalid user id",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfInvalidID(t),
			},
			storage:   existingStorage,
			wantError: true,
		},
		{
			name: "update not existing storage",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfValidID(t),
			},
			storage:       notExistingStorage,
			wantError:     true,
			expectedError: queryerr.ErrStorageNotExists,
		},
		{
			name: "update storage with reserved name",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfValidID(t),
			},
			storage:       storageWithNotUniqueName,
			wantError:     true,
			expectedError: ErrStorageNameNotUnique,
		},
		{
			name: "update existing storage",
			fields: fields{
				repoMock:      repoMock,
				idDecoderMock: newDecoderOfValidID(t),
			},
			storage:        existingStorage,
			wantError:      false,
			expectedResult: &existingStorage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(tt.fields.repoMock, metricsMock)
			service.usrID = tt.fields.idDecoderMock

			result, err := service.Update(context.Background(), tt.storage)

			isCorrectExpectedError := (err != nil) == tt.wantError
			require.Truef(t, isCorrectExpectedError, "error = %v, wantErr %v", err, tt.wantError)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError, "wrong expected error")
			}

			if err == nil {
				assert.Equalf(t, tt.expectedResult, result, "result: %#v\n, expected: %#v\n", result, tt.expectedResult)
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
