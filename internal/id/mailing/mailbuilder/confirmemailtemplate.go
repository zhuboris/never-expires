package mailbuilder

import "github.com/zhuboris/never-expires/internal/id/lang"

type confirmEmailTemplateInput struct {
	Subject         string
	Header          string
	Body            string
	ClickSuggestion string
	Button          string
	Link            string
	Annotation      string
}

func (b Builder) newConfirmEmailTemplateInput(link string, language lang.Language) (confirmEmailTemplateInput, error) {
	content := b.localesDict.ConfirmEmail
	input, err := b.newEmailWithButtonTemplateInput(content, language)
	if err != nil {
		return confirmEmailTemplateInput{}, err
	}

	return confirmEmailTemplateInput{
		Subject:         input.subject,
		Header:          input.header,
		Body:            input.body,
		ClickSuggestion: input.clickSuggestion,
		Button:          input.button,
		Link:            link,
		Annotation:      input.annotation,
	}, nil
}
