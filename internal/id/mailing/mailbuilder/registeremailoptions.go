package mailbuilder

import "html/template"

type RegisterTemplateOption func() (confirmationLink string, template *template.Template)

func (b Builder) WithConfirmationButton(buttonURL string) RegisterTemplateOption {
	return func() (confirmationLink string, template *template.Template) {
		return buttonURL, b.templates.emailWithButton
	}
}

func (b Builder) WithoutConfirmationButton() RegisterTemplateOption {
	return func() (confirmationLink string, template *template.Template) {
		return "", b.templates.messageEmail
	}
}
