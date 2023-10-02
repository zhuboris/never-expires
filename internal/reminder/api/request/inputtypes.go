package request

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zhuboris/never-expires/internal/reminder/item"
)

type (
	itemData struct {
		Name              string      `json:"name"`
		DateAdded         string      `json:"date_added"`
		BestBefore        string      `json:"best_before"`
		IsOpened          bool        `json:"is_opened"`
		HoursAfterOpening int         `json:"hours_after_opening"`
		Note              string      `json:"note"`
		StorageID         pgtype.UUID `json:"storage_id"`
	}
	copyData struct {
		OriginalID pgtype.UUID `json:"original_id"`
		DateAdded  string      `json:"date_added"`
	}
	storageData struct {
		Name string `json:"name"`
	}
	apnsDeviceToken struct {
		Token string `json:"token"`
	}
)

func (d itemData) toValidItem() (item.Item, error) {
	if d.isMissingRequiredField() {
		return item.Item{}, ErrMissingRequiredField
	}

	bestBefore, err := time.Parse(time.RFC3339, d.BestBefore)
	if err != nil {
		return item.Item{}, InvalidTimeFormatError(d.BestBefore)
	}

	return item.Item{
		Name:              d.Name,
		BestBefore:        bestBefore.UTC(),
		IsOpened:          d.IsOpened,
		HoursAfterOpening: d.HoursAfterOpening,
		Note:              d.Note,
	}, nil
}

func (d itemData) toValidItemWithAddedDateRequired() (item.Item, error) {
	if d.DateAdded == "" {
		return item.Item{}, ErrMissingRequiredField
	}

	parsedItem, err := d.toValidItem()
	if err != nil {
		return item.Item{}, err
	}

	dateAdded, err := time.Parse(time.RFC3339, d.DateAdded)
	if err != nil {
		return item.Item{}, InvalidTimeFormatError(d.DateAdded)
	}

	parsedItem.DateAdded = dateAdded.UTC()
	return parsedItem, nil
}

func (d itemData) isMissingRequiredField() bool {
	return d.Name == "" || d.BestBefore == ""
}

func (d itemData) checkIfStorageIDValid() error {
	return checkIsUUIDValid(d.StorageID)
}

func (d copyData) toValidCopyData() (item.ToCopy, error) {
	if d.isMissingRequiredField() {
		return item.ToCopy{}, ErrMissingRequiredField
	}

	if err := d.checkIfOriginalIDValid(); err != nil {
		return item.ToCopy{}, err
	}

	dateAdded, err := time.Parse(time.RFC3339, d.DateAdded)
	if err != nil {
		return item.ToCopy{}, InvalidTimeFormatError(d.DateAdded)
	}

	return item.ToCopy{
		OriginalID: d.OriginalID,
		DateAdded:  dateAdded.UTC(),
	}, nil
}

func (d copyData) isMissingRequiredField() bool {
	return d.DateAdded == ""
}

func (d copyData) checkIfOriginalIDValid() error {
	return checkIsUUIDValid(d.OriginalID)
}

func (d storageData) isMissingRequiredField() bool {
	return d.Name == ""
}

func checkIsUUIDValid(uuid pgtype.UUID) error {
	if uuid == (pgtype.UUID{}) {
		return ErrMissingRequiredField
	}

	if !uuid.Valid {
		return ErrInvalidUUIDInBody
	}

	return nil
}

func (t apnsDeviceToken) isMissing() bool {
	return t.Token == ""
}
