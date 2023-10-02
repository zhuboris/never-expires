package uuidformat

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStrToPgtype(t *testing.T) {
	tests := []struct {
		name        string
		uuid        string
		wantIsValid bool
		wantError   bool
	}{
		{
			name:        "valid uuid",
			uuid:        "52bc1e00-c18d-4251-86c1-92a1d46cf532",
			wantIsValid: true,
			wantError:   false,
		},
		{
			name:        "valid uuid json",
			uuid:        `"52bc1e00-c18d-4251-86c1-92a1d46cf532"`,
			wantIsValid: true,
			wantError:   false,
		},
		{
			name:        "invalid uuid length",
			uuid:        "52bc1e00-c18d-4251-86c1-92a1d46cf2",
			wantIsValid: false,
			wantError:   true,
		},
		{
			name:        "invalid uuid json with valid uuid length",
			uuid:        `"52bc10-c18d-4251-86c1-92a1d46cf2"`,
			wantIsValid: false,
			wantError:   true,
		},
		{
			name:        "invalid uuid with valid length",
			uuid:        "52bc1e00!c18d-4251-86c1-92a1d46cf2",
			wantIsValid: false,
			wantError:   true,
		},
		{
			name:        "empty string",
			uuid:        "",
			wantIsValid: false,
			wantError:   true,
		},
		{
			name:        "empty string json",
			uuid:        `""`,
			wantIsValid: false,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StrToPgtype(tt.uuid)

			isCorrectExpectedError := (err != nil) == tt.wantError
			require.Truef(t, isCorrectExpectedError, "StrToPgtype(): error = %v, wantErr %v", err, tt.wantError)
			assert.Truef(t, got.Valid == tt.wantIsValid, "StrToPgtype(): uuid validity = %t, want = %t, input: %q, len: %d", got.Valid, tt.wantIsValid, tt.uuid, len(tt.uuid))
		})
	}
}

func BenchmarkStrToPgtype(b *testing.B) {
	const uuid = "97e5078e-9236-4253-8c9a-5ad0088d996b"
	for i := 0; i < b.N; i++ {
		_, _ = StrToPgtype(uuid)
	}
}

func BenchmarkStrToPgtypeOld(b *testing.B) {
	const uuid = "97e5078e-9236-4253-8c9a-5ad0088d996b"

	strToPgtypeOld := func(uuidRaw string) (pgtype.UUID, error) {
		uuidRaw = fmt.Sprintf(`"%s"`, uuidRaw)
		var id pgtype.UUID
		if err := id.UnmarshalJSON([]byte(uuidRaw)); err != nil {
			return pgtype.UUID{}, errors.Join(ErrInvalidUUID, err)
		}

		return id, nil
	}

	for i := 0; i < b.N; i++ {
		_, _ = strToPgtypeOld(uuid)
	}
}
