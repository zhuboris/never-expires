package mailbuilder

import "github.com/zhuboris/never-expires/internal/id/lang"

type messageEmailTemplateInput struct {
	subject    string
	header     string
	body       string
	annotation string
}

func (b Builder) newMessageEmailTemplateInput(content messageEmailContent, language lang.Language) (messageEmailTemplateInput, error) {
	subject, err := content.Subject.requestedOrDefaultValue(language)
	if err != nil {
		return messageEmailTemplateInput{}, err
	}

	header, err := content.Header.requestedOrDefaultValue(language)
	if err != nil {
		return messageEmailTemplateInput{}, err
	}

	body, err := content.Body.requestedOrDefaultValue(language)
	if err != nil {
		return messageEmailTemplateInput{}, err
	}

	annotation, err := b.localesDict.Annotation.requestedOrDefaultValue(language)
	if err != nil {
		return messageEmailTemplateInput{}, err
	}

	return messageEmailTemplateInput{
		subject:    subject,
		header:     header,
		body:       body,
		annotation: annotation,
	}, nil
}
