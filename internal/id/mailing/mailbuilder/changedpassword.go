package mailbuilder

import "github.com/zhuboris/never-expires/internal/id/lang"

type changedPasswordTemplateInput struct {
	Subject    string
	Header     string
	Body       string
	Annotation string
}

func (b Builder) newChangedPasswordTemplateInput(language lang.Language) (changedPasswordTemplateInput, error) {
	content := b.localesDict.ChangedPassword
	input, err := b.newMessageEmailTemplateInput(content, language)
	if err != nil {
		return changedPasswordTemplateInput{}, err
	}

	return changedPasswordTemplateInput{
		Subject:    input.subject,
		Header:     input.header,
		Body:       input.body,
		Annotation: input.annotation,
	}, nil
}
