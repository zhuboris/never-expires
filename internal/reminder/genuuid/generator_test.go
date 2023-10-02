package genuuid

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestMakeValidIfNeeded(t *testing.T) {
	tests := []struct {
		name string
		uuid *pgtype.UUID
	}{
		{
			name: "valid uuid",
			uuid: &pgtype.UUID{Bytes: [16]byte{151, 229, 7, 142, 146, 54, 66, 83, 140, 154, 90, 208, 8, 141, 153, 107}, Valid: true},
		},
		{
			name: "not set uuid",
			uuid: &pgtype.UUID{},
		},
		{
			name: "invalid uuid",
			uuid: &pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MakeValidIfNeeded(tt.uuid)
			assert.True(t, tt.uuid.Valid, "uuid is invalid when it must become valid")

			_, err := uuid.Parse(uuid.UUID(tt.uuid.Bytes[:]).String())
			assert.NoError(t, err, "uuid is invalid")
		})
	}
}

func FuzzMakeValidIfNeeded(f *testing.F) {
	f.Fuzz(func(t *testing.T, bytesUUID []byte) {
		if len(bytesUUID) != 16 {
			t.Skip() // We only care about data of length 16
		}

		result := &pgtype.UUID{Bytes: [16]byte(bytesUUID)}

		MakeValidIfNeeded(result)
		assert.True(t, result.Valid, "uuid is invalid when it must become valid")
	})
}
