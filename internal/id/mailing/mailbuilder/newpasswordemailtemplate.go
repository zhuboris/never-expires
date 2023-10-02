package mailbuilder

import "github.com/zhuboris/never-expires/internal/id/lang"

type newPasswordTemplateInput struct {
	Subject    string
	Header     string
	Body       string
	Form       string
	Value      string
	Annotation string
}

func (b Builder) newNewPasswordTemplateInput(value string, language lang.Language) (newPasswordTemplateInput, error) {
	content := b.localesDict.NewPassword
	input, err := b.newMessageEmailTemplateInput(content.messageEmailContent, language)
	if err != nil {
		return newPasswordTemplateInput{}, err
	}

	form, err := content.Form.requestedOrDefaultValue(language)
	if err != nil {
		return newPasswordTemplateInput{}, err
	}

	return newPasswordTemplateInput{
		Subject:    input.subject,
		Header:     input.header,
		Body:       input.body,
		Form:       form,
		Value:      value,
		Annotation: input.annotation,
	}, nil
}
