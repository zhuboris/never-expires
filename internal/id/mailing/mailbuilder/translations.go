package mailbuilder

import (
	"encoding/json"
	"errors"
	"io"
	"os"
)

type (
	translationKeys struct {
		Register         emailWithButtonContent  `json:"register"`
		ConfirmEmail     emailWithButtonContent  `json:"confirm_email"`
		ResetPassword    emailWithButtonContent  `json:"reset_password"`
		NewPassword      newPasswordEmailContent `json:"new_password"`
		ChangeEmail      emailWithButtonContent  `json:"change_email"`
		GoogleConnection messageEmailContent     `json:"google_connection"`
		AppleConnection  messageEmailContent     `json:"apple_connection"`
		ChangedPassword  messageEmailContent     `json:"changed_password"`
		NewDevice        newDeviceEmailContent   `json:"new_device"`
		Annotation       translations            `json:"annotation"`
	}
	emailWithButtonContent struct {
		messageEmailContent

		ClickSuggestion translations `json:"click_suggestion"`
		Button          translations `json:"button"`
	}
	newPasswordEmailContent struct {
		messageEmailContent

		Form translations `json:"form"`
	}
	newDeviceEmailContent struct {
		messageEmailContent

		TimeRow      translations `json:"time_row"`
		DeviceRow    translations `json:"device_row"`
		IPAddressRow translations `json:"ip_address_row"`
		Warning      translations `json:"warning"`
	}
	messageEmailContent struct {
		Subject translations `json:"subject"`
		Header  translations `json:"header"`
		Body    translations `json:"body"`
	}
)

func loadTranslationsFromJSON(path string) (translationKeys, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return translationKeys{}, errors.Join(ErrTranslationsFileNotFound, err)
	}

	defer jsonFile.Close()

	bytes, err := io.ReadAll(jsonFile)
	if err != nil {
		return translationKeys{}, err
	}

	var result translationKeys
	if err := json.Unmarshal(bytes, &result); err != nil {
		return translationKeys{}, err
	}

	return result, nil
}
