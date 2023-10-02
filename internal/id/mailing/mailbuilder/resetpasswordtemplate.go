package mailbuilder

import "github.com/zhuboris/never-expires/internal/id/lang"

type resetPasswordTemplateInput struct {
	Subject         string
	Header          string
	Body            string
	ClickSuggestion string
	Button          string
	Link            string
	Annotation      string
}

func (b Builder) newResetPasswordTemplateInput(link string, language lang.Language) (resetPasswordTemplateInput, error) {
	content := b.localesDict.ResetPassword
	input, err := b.newEmailWithButtonTemplateInput(content, language)
	if err != nil {
		return resetPasswordTemplateInput{}, err
	}

	return resetPasswordTemplateInput{
		Subject:         input.subject,
		Header:          input.header,
		Body:            input.body,
		ClickSuggestion: input.clickSuggestion,
		Button:          input.button,
		Link:            link,
		Annotation:      input.annotation,
	}, nil
}
