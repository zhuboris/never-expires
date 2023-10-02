package mailbuilder

import "github.com/zhuboris/never-expires/internal/id/lang"

type emailWithButtonTemplateInput struct {
	clickSuggestion string
	button          string

	messageEmailTemplateInput
}

func (b Builder) newEmailWithButtonTemplateInput(content emailWithButtonContent, language lang.Language) (emailWithButtonTemplateInput, error) {
	messageInput, err := b.newMessageEmailTemplateInput(content.messageEmailContent, language)
	if err != nil {
		return emailWithButtonTemplateInput{}, err
	}

	clickSuggestion, err := content.ClickSuggestion.requestedOrDefaultValue(language)
	if err != nil {
		return emailWithButtonTemplateInput{}, err
	}

	button, err := content.Button.requestedOrDefaultValue(language)
	if err != nil {
		return emailWithButtonTemplateInput{}, err
	}

	return emailWithButtonTemplateInput{
		messageEmailTemplateInput: messageInput,
		clickSuggestion:           clickSuggestion,
		button:                    button,
	}, nil
}
