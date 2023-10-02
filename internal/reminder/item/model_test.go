package item

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestItem_ToResponseFormat(t *testing.T) {
	intFive := 5

	type fields struct {
		ID                pgtype.UUID
		Name              string
		IsOpened          bool
		BestBefore        time.Time
		ExpirationDate    time.Time
		HoursAfterOpening int
		DateAdded         time.Time
		Note              string
	}
	tests := []struct {
		name   string
		fields fields
		want   ResponseItem
	}{
		{
			name: "hours after opening is not provided",
			fields: fields{
				ID:             pgtype.UUID{Valid: true},
				Name:           "test item",
				IsOpened:       false,
				BestBefore:     time.Date(2023, time.July, 23, 13, 45, 0, 0, time.UTC),
				ExpirationDate: time.Date(2023, time.July, 24, 13, 45, 0, 0, time.UTC),
				DateAdded:      time.Date(2023, time.July, 24, 13, 45, 0, 0, time.UTC),
				Note:           "some note",
			},
			want: ResponseItem{
				ID:                pgtype.UUID{Valid: true},
				Name:              "test item",
				IsOpened:          false,
				BestBefore:        "2023-07-23T13:45:00Z",
				ExpirationDate:    "2023-07-24T13:45:00Z",
				HoursAfterOpening: nil,
				DateAdded:         "2023-07-24T13:45:00Z",
				Note:              "some note",
			},
		},
		{
			name: "HoursAfterOpening is zero",
			fields: fields{
				ID:                pgtype.UUID{Valid: true},
				Name:              "test item",
				IsOpened:          false,
				BestBefore:        time.Date(2023, time.July, 23, 13, 45, 0, 0, time.UTC),
				ExpirationDate:    time.Date(2023, time.July, 24, 13, 45, 0, 0, time.UTC),
				DateAdded:         time.Date(2023, time.July, 24, 13, 45, 0, 0, time.UTC),
				HoursAfterOpening: 0,
				Note:              "some note",
			},
			want: ResponseItem{
				ID:                pgtype.UUID{Valid: true},
				Name:              "test item",
				IsOpened:          false,
				BestBefore:        "2023-07-23T13:45:00Z",
				ExpirationDate:    "2023-07-24T13:45:00Z",
				HoursAfterOpening: nil,
				DateAdded:         "2023-07-24T13:45:00Z",
				Note:              "some note",
			},
		},
		{
			name: "hours after opening is provided",
			fields: fields{
				ID:                pgtype.UUID{Valid: true},
				Name:              "test item",
				IsOpened:          false,
				BestBefore:        time.Date(2023, time.July, 23, 13, 45, 0, 0, time.UTC),
				ExpirationDate:    time.Date(2023, time.July, 24, 13, 45, 0, 0, time.UTC),
				DateAdded:         time.Date(2023, time.July, 24, 13, 45, 0, 0, time.UTC),
				HoursAfterOpening: 5,
				Note:              "some note",
			},
			want: ResponseItem{
				ID:                pgtype.UUID{Valid: true},
				Name:              "test item",
				IsOpened:          false,
				BestBefore:        "2023-07-23T13:45:00Z",
				ExpirationDate:    "2023-07-24T13:45:00Z",
				HoursAfterOpening: &intFive,
				DateAdded:         "2023-07-24T13:45:00Z",
				Note:              "some note",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Item{
				ID:                tt.fields.ID,
				Name:              tt.fields.Name,
				IsOpened:          tt.fields.IsOpened,
				BestBefore:        tt.fields.BestBefore,
				ExpirationDate:    tt.fields.ExpirationDate,
				HoursAfterOpening: tt.fields.HoursAfterOpening,
				DateAdded:         tt.fields.DateAdded,
				Note:              tt.fields.Note,
			}

			result := i.ToResponseFormat()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestItems_ToResponseFormat(t *testing.T) {
	intFive := 5

	tests := []struct {
		name  string
		items Items
		want  *[]ResponseItem
	}{
		{
			name:  "no items",
			items: Items{},
			want:  &[]ResponseItem{},
		},
		{
			name:  "nil items",
			items: nil,
			want:  &[]ResponseItem{},
		},
		{
			name: "some items",
			items: Items{
				{
					ID:             pgtype.UUID{Valid: true},
					Name:           "test item with hours after opening is not provided",
					IsOpened:       false,
					BestBefore:     time.Date(2023, time.July, 23, 13, 45, 0, 0, time.UTC),
					ExpirationDate: time.Date(2023, time.July, 24, 13, 45, 0, 0, time.UTC),
					DateAdded:      time.Date(2023, time.July, 24, 13, 45, 0, 0, time.UTC),
					Note:           "some note",
				},
				{
					ID:                pgtype.UUID{Valid: true},
					Name:              "test item with hours after opening is zero",
					IsOpened:          false,
					BestBefore:        time.Date(2023, time.July, 23, 13, 45, 0, 0, time.UTC),
					ExpirationDate:    time.Date(2023, time.July, 24, 13, 45, 0, 0, time.UTC),
					DateAdded:         time.Date(2023, time.July, 24, 13, 45, 0, 0, time.UTC),
					HoursAfterOpening: 0,
					Note:              "some note",
				},
				{
					ID:                pgtype.UUID{Valid: true},
					Name:              "test item with hours after opening is provided",
					IsOpened:          false,
					BestBefore:        time.Date(2023, time.July, 23, 13, 45, 0, 0, time.UTC),
					ExpirationDate:    time.Date(2023, time.July, 24, 13, 45, 0, 0, time.UTC),
					DateAdded:         time.Date(2023, time.July, 24, 13, 45, 0, 0, time.UTC),
					HoursAfterOpening: 5,
					Note:              "some note",
				},
			},
			want: &[]ResponseItem{
				{
					ID:                pgtype.UUID{Valid: true},
					Name:              "test item with hours after opening is not provided",
					IsOpened:          false,
					BestBefore:        "2023-07-23T13:45:00Z",
					ExpirationDate:    "2023-07-24T13:45:00Z",
					HoursAfterOpening: nil,
					DateAdded:         "2023-07-24T13:45:00Z",
					Note:              "some note",
				},
				{
					ID:                pgtype.UUID{Valid: true},
					Name:              "test item with hours after opening is zero",
					IsOpened:          false,
					BestBefore:        "2023-07-23T13:45:00Z",
					ExpirationDate:    "2023-07-24T13:45:00Z",
					HoursAfterOpening: nil,
					DateAdded:         "2023-07-24T13:45:00Z",
					Note:              "some note",
				},
				{
					ID:                pgtype.UUID{Valid: true},
					Name:              "test item with hours after opening is provided",
					IsOpened:          false,
					BestBefore:        "2023-07-23T13:45:00Z",
					ExpirationDate:    "2023-07-24T13:45:00Z",
					HoursAfterOpening: &intFive,
					DateAdded:         "2023-07-24T13:45:00Z",
					Note:              "some note",
				},
			},
		},
	}

	for _, tt := range tests {
		result := tt.items.ToResponseFormat()
		assert.Equal(t, tt.want, result)
	}
}
