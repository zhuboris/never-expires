package fromenv

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInt(t *testing.T) {
	tests := []struct {
		name           string
		envValue       string
		expectedResult int
		wantError      bool
	}{
		{
			name:           "correct integer in string",
			envValue:       "10",
			expectedResult: 10,
			wantError:      false,
		},
		{
			name:           "string with valid integer starting with zeros",
			envValue:       "000010",
			expectedResult: 10,
			wantError:      false,
		},
		{
			name:           "string with valid integer starting with plus",
			envValue:       "+10",
			expectedResult: 10,
			wantError:      false,
		},
		{
			name:           "string with negative integer",
			envValue:       "-10",
			expectedResult: -10,
			wantError:      false,
		},
		{
			name:           "string with negative integer starting with zeros",
			envValue:       "-00000010",
			expectedResult: -10,
			wantError:      false,
		},
		{
			name:           "string with chars",
			envValue:       "test",
			expectedResult: 0,
			wantError:      true,
		},
		{
			name:           "string starting with space",
			envValue:       " 10",
			expectedResult: 0,
			wantError:      true,
		},
		{
			name:           "string ending with space",
			envValue:       "10 ",
			expectedResult: 0,
			wantError:      true,
		},
		{
			name:           "string with non digits between",
			envValue:       "10-10",
			expectedResult: 0,
			wantError:      true,
		},
		{
			name:           "string with valid number but not integer",
			envValue:       "10.10",
			expectedResult: 0,
			wantError:      true,
		},
		{
			name:           "string with float",
			envValue:       "10.10",
			expectedResult: 0,
			wantError:      true,
		},
		{
			name:           "string with integer bigger than int max value",
			envValue:       "10000000000000000000000",
			expectedResult: 0,
			wantError:      true,
		},
		{
			name:           "empty string",
			envValue:       "",
			expectedResult: 0,
			wantError:      true,
		},
		{
			name:           "string with 0",
			envValue:       "0",
			expectedResult: 0,
			wantError:      false,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := fmt.Sprintf("TEST_INT_%d", i)
			err := os.Setenv(key, tt.envValue)
			require.NoError(t, err, "failed to setup test env")

			t.Cleanup(func() {
				err := os.Unsetenv(key)
				require.NoError(t, err, "failed to cleanup")
			})

			result, err := Int(key)

			if err != nil {
				require.Truef(t, tt.wantError, "error = %v, wantErr %v", err, tt.wantError)
			}

			assert.Equal(t, tt.expectedResult, result, "incorrect result")
		})
	}
}
