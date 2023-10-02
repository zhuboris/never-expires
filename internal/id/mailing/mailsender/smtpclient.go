package mailsender

import (
	"context"
	"crypto/tls"
	"errors"
	"net/smtp"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/zhuboris/never-expires/internal/id/mailing/mailmsg"
	"github.com/zhuboris/never-expires/internal/shared/str"
	"github.com/zhuboris/never-expires/internal/shared/try"
)

const attemptsDelay = 30 * time.Second

const timeLogKey = "elapsedTime"

var errSMPTConnectionRefused = errors.New("cannot connect to given SMPT server")

type SmtpClient struct {
	client    *smtp.Client
	config    Config
	tlsConfig *tls.Config
	from      string
	logger    *zap.Logger
}

func NewSMTPClient(ctx context.Context, config Config, logger *zap.Logger) (*SmtpClient, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         config.host,
	}

	smtpClient := &SmtpClient{
		tlsConfig: tlsConfig,
		config:    config,
		from:      config.from,
		logger:    logger,
	}

	if err := try.DoWithAttempts(ctx, smtpClient.connect, attemptsDelay); err != nil {
		return nil, err
	}

	return smtpClient, nil
}

func (c *SmtpClient) Quit() error {
	return c.client.Quit()
}

func (c *SmtpClient) sendEmail(ctx context.Context, messageID string, message mailmsg.Message) error {
	sendingFunc := func() error {
		var err error
		defer func(startTime time.Time) {
			c.logSending(startTime, messageID, message.Recipient, err)
		}(time.Now())

		if err = c.reconnectIfNeeded(); err != nil {
			return err
		}

		err = c.sendMessage(message.Recipient, message.Email)
		return err
	}

	return try.DoWithAttempts(ctx, sendingFunc, attemptsDelay)
}

func (c *SmtpClient) connect() error {
	var err error
	defer func(startTime time.Time) {
		c.logConnection(startTime, err)
	}(time.Now())

	conn, err := tls.Dial("tcp", c.config.serverName(), c.tlsConfig)
	if err != nil {
		return errors.Join(errSMPTConnectionRefused, err)
	}

	client, err := smtp.NewClient(conn, c.config.host)
	if err != nil {
		return errors.Join(errSMPTConnectionRefused, err)
	}

	if err = client.Auth(c.config.auth()); err != nil {
		return errors.Join(errSMPTConnectionRefused, err)
	}

	c.client = client
	return nil
}

func (c *SmtpClient) reconnectIfNeeded() error {
	if err := c.client.Noop(); err != nil {
		return c.connect()
	}

	return nil
}

func (c *SmtpClient) sendMessage(recipient string, msg []byte) error {
	if err := c.client.Mail(c.from); err != nil {
		return err
	}
	if err := c.client.Rcpt(recipient); err != nil {
		return err
	}

	w, err := c.client.Data()
	if err != nil {
		return err
	}

	if _, err := w.Write(msg); err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	return c.client.Reset()
}

func (c *SmtpClient) logConnection(startTime time.Time, err error) {
	elapsedTime := time.Since(startTime)

	if err != nil {
		c.logger.Error(str.Capitalized(errSMPTConnectionRefused.Error()), zap.Error(err), zap.Duration(timeLogKey, elapsedTime))
	}

	c.logger.Info("Successfully connected to smtp server", zap.Duration(timeLogKey, elapsedTime))
}

func (c *SmtpClient) logSending(startTime time.Time, messageID, recipient string, err error) {
	const messageIDLogKey = "messageID"
	const recipientLogKey = "recipient"

	elapsedTime := time.Since(startTime)

	msg := "Successfully sent email"
	lvl := zapcore.InfoLevel
	if err != nil {
		msg = "Email was not send, an error occurred"
		lvl = zapcore.ErrorLevel
	}

	c.logger.Log(lvl, msg, zap.String(messageIDLogKey, messageID), zap.String(recipientLogKey, recipient), zap.Error(err), zap.Duration(timeLogKey, elapsedTime))
}
