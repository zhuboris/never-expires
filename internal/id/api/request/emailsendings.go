package request

import (
	"context"
	"net/http"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/zhuboris/never-expires/internal/id/lang"
	"github.com/zhuboris/never-expires/internal/id/mailing/mailbuilder"
	"github.com/zhuboris/never-expires/internal/id/usr/oauth"
)

type (
	emailMessages interface {
		ConfirmEmail(recipient, url string, language lang.Language) ([]byte, error)
		ConfirmEmailOnChange(recipient, url string, language lang.Language) ([]byte, error)
		ResetPassword(recipient, url string, language lang.Language) ([]byte, error)
		PasswordIsChanged(recipient string, language lang.Language) ([]byte, error)
		OAuthAccountConnected(recipient string, language lang.Language, connectionType oauth.Type) ([]byte, error)
		NewPassword(recipient, password string, language lang.Language) ([]byte, error)
		NewDeviceLogin(recipient string, data mailbuilder.NotificationData, language lang.Language) ([]byte, error)
		Register(recipient string, language lang.Language, option mailbuilder.RegisterTemplateOption) ([]byte, error)
		WithConfirmationButton(buttonURL string) mailbuilder.RegisterTemplateOption
		WithoutConfirmationButton() mailbuilder.RegisterTemplateOption
	}
	emailQueueAdder interface {
		Add(ctx context.Context, recipient string, msg []byte) error
	}
)

type messageFunc func(lang.Language) ([]byte, error)

type EmailSender struct {
	messages emailMessages
	queue    emailQueueAdder
	logger   *zap.Logger
}

func (s EmailSender) addToQueue(ctx context.Context, cancel context.CancelFunc, r *http.Request, recipient string, msgFunc messageFunc) {
	defer cancel()

	var err error
	defer func() {
		s.logEnqueueResult(err)

	}()

	language := s.emailResponseLanguage(r)
	msg, err := msgFunc(language)
	if err != nil {
		return
	}

	err = s.queue.Add(ctx, recipient, msg)
}

func (s EmailSender) registerMessage(recipient string, templateOption mailbuilder.RegisterTemplateOption) messageFunc {
	return func(language lang.Language) ([]byte, error) {
		return s.messages.Register(recipient, language, templateOption)
	}
}

func (s EmailSender) confirmEmailMessage(recipient, url string) messageFunc {
	return func(language lang.Language) ([]byte, error) {
		return s.messages.ConfirmEmail(recipient, url, language)
	}
}

func (s EmailSender) confirmEmailOnChangeMessage(recipient, url string) messageFunc {
	return func(language lang.Language) ([]byte, error) {
		return s.messages.ConfirmEmailOnChange(recipient, url, language)
	}
}

func (s EmailSender) resetPasswordMessage(recipient, url string) messageFunc {
	return func(language lang.Language) ([]byte, error) {
		return s.messages.ResetPassword(recipient, url, language)
	}
}

func (s EmailSender) newPasswordMessage(recipient, password string) messageFunc {
	return func(language lang.Language) ([]byte, error) {
		return s.messages.NewPassword(recipient, password, language)
	}
}

func (s EmailSender) passwordIsChangedMessage(recipient string) messageFunc {
	return func(language lang.Language) ([]byte, error) {
		return s.messages.PasswordIsChanged(recipient, language)
	}
}

func (s EmailSender) oAuthConnectionMessage(recipient string, connectionType oauth.Type) messageFunc {
	return func(language lang.Language) ([]byte, error) {
		return s.messages.OAuthAccountConnected(recipient, language, connectionType)
	}
}

func (s EmailSender) newDeviceLoginMessage(recipient string, data mailbuilder.NotificationData) messageFunc {
	return func(language lang.Language) ([]byte, error) {
		return s.messages.NewDeviceLogin(recipient, data, language)
	}
}

func (s EmailSender) emailResponseLanguage(r *http.Request) lang.Language {
	const headerWithLocaleName = "Accept-Language"

	localeIdentifier := r.Header.Get(headerWithLocaleName)
	return lang.FromLocaleIdentifier(localeIdentifier)
}

func (s EmailSender) logEnqueueResult(err error) {
	msg := "Added to queue successfully"
	logLvl := zapcore.InfoLevel
	if err != nil {
		msg = "Failed to add to queue"
		logLvl = zapcore.ErrorLevel
	}

	s.logger.Log(logLvl, msg, zap.Error(err))
}

func (s EmailSender) registerWithConfirmationButton(url string) mailbuilder.RegisterTemplateOption {
	return s.messages.WithConfirmationButton(url)
}

func (s EmailSender) registerWithoutConfirmationButton() mailbuilder.RegisterTemplateOption {
	return s.messages.WithoutConfirmationButton()
}
