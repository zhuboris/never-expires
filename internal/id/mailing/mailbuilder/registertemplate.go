package mailbuilder

import (
	"github.com/zhuboris/never-expires/internal/id/lang"
)

type registerTemplateInput struct {
	Subject         string
	Header          string
	Body            string
	ClickSuggestion string
	Button          string
	Link            string
	Annotation      string
}

func (b Builder) newRegisterTemplateInput(link string, language lang.Language) (registerTemplateInput, error) {
	content := b.localesDict.Register
	input, err := b.newEmailWithButtonTemplateInput(content, language)
	if err != nil {
		return registerTemplateInput{}, err
	}

	return registerTemplateInput{
		Subject:         input.subject,
		Header:          input.header,
		Body:            input.body,
		ClickSuggestion: input.clickSuggestion,
		Button:          input.button,
		Link:            link,
		Annotation:      input.annotation,
	}, nil
}
