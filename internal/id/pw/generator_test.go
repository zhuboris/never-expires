package pw

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkGenerate(b *testing.B) {
	for i := 0; i <= 20; i++ {
		n := i
		b.Run(fmt.Sprintf("inputted len = %d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = Generate(n)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	const minLength = int(optionsCount)

	tests := []struct {
		name           string
		inputtedLength int
		expectedLength int
	}{
		{
			name:           "Length more that min length",
			inputtedLength: minLength + 10,
			expectedLength: minLength + 10,
		},
		{
			name:           "Length equal min length",
			inputtedLength: minLength,
			expectedLength: minLength,
		},
		{
			name:           "Length less that min length",
			inputtedLength: minLength - 2,
			expectedLength: minLength,
		},
		{
			name:           "Zero length",
			inputtedLength: 0,
			expectedLength: minLength,
		},
		{
			name:           "Negative length",
			inputtedLength: -100,
			expectedLength: minLength,
		},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s: inputtedLength = %d", tt.name, tt.inputtedLength)
		t.Run(name, func(t *testing.T) {
			result := Generate(tt.inputtedLength)
			t.Logf("result = %s", result)

			assert.Equal(t, tt.expectedLength, len(result), "Incorrect password length")
			assert.True(t, containsLower(t, result), "The password does not contain any lowers")
			assert.True(t, containsUpper(t, result), "The password does not contain any uppers")
			assert.True(t, containsDigit(t, result), "The password does not contain any digits")
			assert.True(t, containsSpecial(t, result), "The password does not contain any specials")

			if len(result) >= minValidLength {
				err := Validate(result)
				assert.NoError(t, err, "The password does not fit required validation function")
			}
		})
	}
}

func containsLower(t *testing.T, s string) bool {
	t.Helper()
	regex := regexp.MustCompile(`[a-z]`)
	return regex.MatchString(s)
}

func containsUpper(t *testing.T, s string) bool {
	t.Helper()
	regex := regexp.MustCompile(`[A-Z]`)
	return regex.MatchString(s)
}

func containsDigit(t *testing.T, s string) bool {
	t.Helper()
	regex := regexp.MustCompile(`\d`)
	return regex.MatchString(s)
}

func containsSpecial(t *testing.T, s string) bool {
	t.Helper()
	regex := regexp.MustCompile(`\W`)
	return regex.MatchString(s)
}
