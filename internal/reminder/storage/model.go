package storage

import "github.com/jackc/pgx/v5/pgtype"

type Storage struct {
	ID         pgtype.UUID `json:"id"`
	Name       string      `json:"name"`
	ItemsCount int         `json:"items_count"`
	IsDefault  bool        `json:"is_default"`
}

type Entity struct {
	ID         *pgtype.UUID `json:"id"`
	Name       *string      `json:"name"`
	ItemsCount *int         `json:"items_count"`
	IsDefault  *bool        `json:"is_default"`
}

func (e Entity) storage() *Storage {
	var result Storage
	if e.ID != nil {
		result.ID = *e.ID
	}

	if e.Name != nil {
		result.Name = *e.Name
	}

	if e.ItemsCount != nil {
		result.ItemsCount = *e.ItemsCount
	}

	if e.IsDefault != nil {
		result.IsDefault = *e.IsDefault
	}

	return &result
}
