package item

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type (
	Item struct {
		ID                pgtype.UUID `json:"id"`
		Name              string      `json:"name"`
		IsOpened          bool        `json:"is_opened"`
		BestBefore        time.Time   `json:"best_before"`
		ExpirationDate    time.Time   `json:"expiration_date"`
		HoursAfterOpening int         `json:"hours_after_opening"`
		DateAdded         time.Time   `json:"date_added"`
		Note              string      `json:"note"`
	}
	ResponseItem struct {
		ID                pgtype.UUID `json:"id"`
		Name              string      `json:"name"`
		IsOpened          bool        `json:"is_opened"`
		BestBefore        string      `json:"best_before"`
		ExpirationDate    string      `json:"expiration_date"`
		HoursAfterOpening *int        `json:"hours_after_opening"`
		DateAdded         string      `json:"date_added"`
		Note              string      `json:"note"`
	}
)

func (i *Item) ToResponseFormat() ResponseItem {
	var hoursAfterOpening *int
	if i.HoursAfterOpening != 0 {
		hoursAfterOpening = new(int)
		*hoursAfterOpening = i.HoursAfterOpening
	}

	return ResponseItem{
		ID:                i.ID,
		Name:              i.Name,
		IsOpened:          i.IsOpened,
		BestBefore:        i.BestBefore.UTC().Format(time.RFC3339),
		ExpirationDate:    i.ExpirationDate.UTC().Format(time.RFC3339),
		HoursAfterOpening: hoursAfterOpening,
		DateAdded:         i.DateAdded.UTC().Format(time.RFC3339),
		Note:              i.Note,
	}
}

func (i *Item) isEqual(other *Item) bool {
	return i.ID == other.ID &&
		i.BestBefore == other.BestBefore &&
		i.Name == other.Name &&
		i.IsOpened == other.IsOpened &&
		i.HoursAfterOpening == other.HoursAfterOpening &&
		i.Note == other.Note
}

func (i *Item) updateExpirationDate(oldItem Item) {
	i.ExpirationDate = oldItem.ExpirationDate
	if i.IsOpened && !oldItem.IsOpened {
		i.ExpirationDate = i.timeAfterOpening()
	}

	if !i.IsOpened {
		i.ExpirationDate = i.BestBefore
	}
}

func (i *Item) timeAfterOpening() time.Time {
	duration := time.Duration(i.HoursAfterOpening) * time.Hour
	return time.Now().
		Add(duration)
}

type Entity struct {
	ID                *pgtype.UUID `json:"id"`
	Name              *string      `json:"name"`
	IsOpened          *bool        `json:"is_opened"`
	BestBefore        *time.Time   `json:"best_before"`
	ExpirationDate    *time.Time   `json:"expiration_date"`
	HoursAfterOpening *int         `json:"hours_after_opening"`
	DateAdded         *time.Time   `json:"date_added"`
	Note              *string      `json:"note"`
}

func newFromItem(item Item) *Entity {
	return &Entity{
		ID:                &item.ID,
		Name:              &item.Name,
		IsOpened:          &item.IsOpened,
		BestBefore:        &item.BestBefore,
		ExpirationDate:    &item.ExpirationDate,
		HoursAfterOpening: &item.HoursAfterOpening,
		DateAdded:         &item.DateAdded,
		Note:              &item.Note,
	}
}

func (e Entity) item() *Item {
	var item Item
	if e.ID != nil {
		item.ID = *e.ID
	}

	if e.Name != nil {
		item.Name = *e.Name
	}

	if e.IsOpened != nil {
		item.IsOpened = *e.IsOpened
	}
	if e.BestBefore != nil {
		item.BestBefore = *e.BestBefore
	}

	if e.ExpirationDate != nil {
		item.ExpirationDate = *e.ExpirationDate
	}

	if e.HoursAfterOpening != nil {
		item.HoursAfterOpening = *e.HoursAfterOpening
	}

	if e.DateAdded != nil {
		item.DateAdded = *e.DateAdded
	}

	if e.Note != nil {
		item.Note = *e.Note
	}

	return &item
}

type Items []Item

func (i *Items) ToResponseFormat() *[]ResponseItem {
	response := make([]ResponseItem, 0, len(*i))
	for _, item := range *i {
		response = append(response, item.ToResponseFormat())
	}

	return &response
}

type ToCopy struct {
	OriginalID pgtype.UUID
	NewID      pgtype.UUID
	DateAdded  time.Time
}
