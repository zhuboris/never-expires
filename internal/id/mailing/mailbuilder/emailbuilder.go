package mailbuilder

import (
	"bytes"
	"errors"
	"html/template"

	"github.com/zhuboris/never-expires/internal/id/lang"
	"github.com/zhuboris/never-expires/internal/id/usr/oauth"
)

const translationsDictionaryPath = "web/emails/translations.json"

type Builder struct {
	templates   *htmlTemplates
	localesDict translationKeys
}

func New() (*Builder, error) {
	templates, err := newHtmlTemplates()
	if err != nil {
		return nil, err
	}

	dict, err := loadTranslationsFromJSON(translationsDictionaryPath)
	if err != nil {
		return nil, err
	}

	return &Builder{
		templates:   templates,
		localesDict: dict,
	}, nil
}

func (b Builder) Register(recipient string, language lang.Language, option RegisterTemplateOption) ([]byte, error) {
	url, emailTemplate := option()

	input, err := b.newRegisterTemplateInput(url, language)
	if err != nil {
		return nil, err
	}

	return makeEmailFromTemplate(recipient, input.Subject, input, emailTemplate)
}

func (b Builder) ConfirmEmail(recipient, url string, language lang.Language) ([]byte, error) {
	input, err := b.newConfirmEmailTemplateInput(url, language)
	if err != nil {
		return nil, err
	}

	return makeEmailFromTemplate(recipient, input.Subject, input, b.templates.emailWithButton)
}

func (b Builder) ConfirmEmailOnChange(recipient, url string, language lang.Language) ([]byte, error) {
	input, err := b.newChangeEmailTemplateInput(url, language)
	if err != nil {
		return nil, err
	}

	return makeEmailFromTemplate(recipient, input.Subject, input, b.templates.emailWithButton)
}

func (b Builder) ResetPassword(recipient, url string, language lang.Language) ([]byte, error) {
	input, err := b.newResetPasswordTemplateInput(url, language)
	if err != nil {
		return nil, err
	}

	return makeEmailFromTemplate(recipient, input.Subject, input, b.templates.emailWithButton)
}

func (b Builder) PasswordIsChanged(recipient string, language lang.Language) ([]byte, error) {
	input, err := b.newChangedPasswordTemplateInput(language)
	if err != nil {
		return nil, err
	}

	return makeEmailFromTemplate(recipient, input.Subject, input, b.templates.messageEmail)
}

func (b Builder) OAuthAccountConnected(recipient string, language lang.Language, connectionType oauth.Type) ([]byte, error) {
	input, err := b.newOauthConnectionTemplateInputTemplateInput(language, connectionType)
	if err != nil {
		return nil, err
	}

	return makeEmailFromTemplate(recipient, input.Subject, input, b.templates.messageEmail)
}

func (b Builder) NewPassword(recipient, password string, language lang.Language) ([]byte, error) {
	input, err := b.newNewPasswordTemplateInput(password, language)
	if err != nil {
		return nil, err
	}

	return makeEmailFromTemplate(recipient, input.Subject, input, b.templates.newPasswordEmail)
}

func (b Builder) NewDeviceLogin(recipient string, data NotificationData, language lang.Language) ([]byte, error) {
	input, err := b.newNewDeviceTemplateInput(data, language)
	if err != nil {
		return nil, err
	}

	return makeEmailFromTemplate(recipient, input.Subject, input, b.templates.newDeviceEmail)
}

func makeEmailFromTemplate(recipient, subject string, data any, bodyTemplate *template.Template) ([]byte, error) {
	body, err := messageBody(data, bodyTemplate)
	if err != nil {
		return nil, err
	}

	return buildEmailMessage(recipient, subject, body)
}

func messageBody(data any, bodyTemplate *template.Template) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	if err := bodyTemplate.Execute(buffer, data); err != nil {
		return buffer, errors.Join(ErrFailedExecuteTemplate, err)
	}

	return buffer, nil
}

func buildEmailMessage(recipient, subject string, body *bytes.Buffer) ([]byte, error) {
	buffer := new(bytes.Buffer)
	buffer.WriteString("To: ")
	buffer.WriteString(recipient)
	buffer.WriteString("\r\nSubject: ")
	buffer.WriteString(subject)
	buffer.WriteString("\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
	_, err := buffer.ReadFrom(body)

	return buffer.Bytes(), err
}
