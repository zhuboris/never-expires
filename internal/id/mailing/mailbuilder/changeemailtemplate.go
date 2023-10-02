package mailbuilder

import "github.com/zhuboris/never-expires/internal/id/lang"

type changeEmailTemplateInput struct {
	Subject         string
	Header          string
	Body            string
	ClickSuggestion string
	Button          string
	Link            string
	Annotation      string
}

func (b Builder) newChangeEmailTemplateInput(link string, language lang.Language) (changeEmailTemplateInput, error) {
	content := b.localesDict.ChangeEmail
	input, err := b.newEmailWithButtonTemplateInput(content, language)
	if err != nil {
		return changeEmailTemplateInput{}, err
	}

	return changeEmailTemplateInput{
		Subject:         input.subject,
		Header:          input.header,
		Body:            input.body,
		ClickSuggestion: input.clickSuggestion,
		Button:          input.button,
		Link:            link,
		Annotation:      input.annotation,
	}, nil
}
