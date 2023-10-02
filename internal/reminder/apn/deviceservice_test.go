package apn

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/zhuboris/never-expires/internal/shared/servicechecker"
)

type statusDisplayMock struct{}

func (m statusDisplayMock) Set(error) {}

var metricsMock statusDisplayMock

func TestNewDeviceService(t *testing.T) {
	repo := NewMockdeviceRepo(t)
	service := NewDeviceService(repo, metricsMock)
	require.Implements(t, (*deviceRepo)(nil), repo, "mock is not implement required interface")
	require.Equal(t, repo, service.repo, "mock is not suitable")
}

func TestDeviceService_AddDeviceToken(t *testing.T) {
	repoMock := NewMockdeviceRepo(t)
	repoMock.EXPECT().
		addDeviceToken(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	tests := []struct {
		name          string
		token         string
		idDecoderMock userIDDecoder
		requireError  require.ErrorAssertionFunc
	}{
		{
			name:          "invalid user id",
			token:         "anyToken",
			idDecoderMock: newDecoderOfInvalidID(t),
			requireError:  require.Error,
		},
		{
			name:          "empty token",
			token:         "",
			idDecoderMock: newDecoderOfValidID(t),
			requireError:  require.Error,
		},
		{
			name:          "valid user id and token",
			token:         "anyToken",
			idDecoderMock: newDecoderOfValidID(t),
			requireError:  require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewDeviceService(repoMock, metricsMock)
			service.usrID = tt.idDecoderMock

			err := service.AddDeviceToken(context.Background(), tt.token)

			tt.requireError(t, err)
		})
	}
}

func TestDeviceService_Status(t *testing.T) {
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
			repo := NewMockdeviceRepo(t)
			repo.EXPECT().
				Ping(mock.Anything).
				Return(tt.expectedErrorType)
			service := NewDeviceService(repo, metricsMock)

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
		Return(pgtype.UUID{Valid: true}, nil).Maybe()
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
