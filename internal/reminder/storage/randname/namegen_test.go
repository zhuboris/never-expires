package randname

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageName(t *testing.T) {
	for i := 0; i < 1000; i++ {
		result := StorageName()
		assert.NotEmptyf(t, result, "generated name is empty, name: %q", result)
	}
}

func TestStorageNameWithDigits(t *testing.T) {
	const (
		containsDigitsPattern = `\d`
		testingIterations     = 1000
	)

	regex, err := regexp.Compile(containsDigitsPattern)
	require.NoError(t, err, "Failed to compile regex")

	for i := 0; i < testingIterations; i++ {
		result := StorageNameWithDigits()

		require.NotEmptyf(t, result, "Generated name is empty, name: %q", result)
		containsDigit := regex.MatchString(result)
		assert.Truef(t, containsDigit, "Generated name do not contains digits, name: %q", result)
	}
}
