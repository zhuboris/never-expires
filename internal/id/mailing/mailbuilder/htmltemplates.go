package mailbuilder

import (
	"errors"
	"html/template"
)

const (
	emailWithButtonTemplatePath  = "web/emails/templates/with_button.html"
	newPasswordEmailTemplatePath = "web/emails/templates/new_password.html"
	newDeviceEmailTemplatePath   = "web/emails/templates/new_device.html"
	messageEmailTemplatePath     = "web/emails/templates/message.html"
)

type htmlTemplates struct {
	emailWithButton  *template.Template
	newPasswordEmail *template.Template
	newDeviceEmail   *template.Template
	messageEmail     *template.Template
}

func newHtmlTemplates() (*htmlTemplates, error) {
	emailWithButtonTemplate, err := parseTemplate(emailWithButtonTemplatePath)
	if err != nil {
		return nil, err
	}

	newPasswordEmailTemplate, err := parseTemplate(newPasswordEmailTemplatePath)
	if err != nil {
		return nil, err
	}

	newDeviceEmailTemplate, err := parseTemplate(newDeviceEmailTemplatePath)
	if err != nil {
		return nil, err
	}

	messageEmailTemplate, err := parseTemplate(messageEmailTemplatePath)
	if err != nil {
		return nil, err
	}

	return &htmlTemplates{
		emailWithButton:  emailWithButtonTemplate,
		newPasswordEmail: newPasswordEmailTemplate,
		newDeviceEmail:   newDeviceEmailTemplate,
		messageEmail:     messageEmailTemplate,
	}, nil
}

func parseTemplate(path string) (*template.Template, error) {
	emailTemplate, err := template.ParseFiles(path)
	if err != nil {
		templateError := InvalidTemplateError{
			path: path,
		}
		return nil, errors.Join(templateError, err)
	}

	return emailTemplate, nil
}
