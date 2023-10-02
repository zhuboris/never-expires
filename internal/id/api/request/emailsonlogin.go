package request

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/mileusna/useragent"

	"github.com/zhuboris/never-expires/internal/id/mailing/mailbuilder"
	"github.com/zhuboris/never-expires/internal/id/session"
	"github.com/zhuboris/never-expires/internal/id/usr"
	"github.com/zhuboris/never-expires/internal/id/usr/oauth"
)

type successLoginData struct {
	user       *usr.PublicData
	newSession session.Session
	loginTime  time.Time
	request    *http.Request
	service    AuthService
}

func sendEmailOnOAuthLogin(resultType oauth.LoginResultType, loginInfo successLoginData, oAuthType oauth.Type) error {
	switch resultType {
	case oauth.Register:
		return sendRegisterWithOAuthNotification(loginInfo)
	case oauth.Login:
		return sendNewDeviceNotifyIfNeeded(loginInfo)
	case oauth.Connect:
		return sendOAuthConnectionNotification(loginInfo, oAuthType)
	default:
		return nil
	}
}

func sendRegisterWithOAuthNotification(loginData successLoginData) error {
	sendingCtx, cancel := ctxWithTimeoutToSendMail()

	option, err := oAuthRegisterTemplateOption(sendingCtx, loginData)
	if err != nil {
		cancel()
		return err
	}

	msg := emailSender.registerMessage(loginData.user.Email, option)
	go emailSender.addToQueue(sendingCtx, cancel, loginData.request, loginData.user.Email, msg)
	return nil
}

func sendNewDeviceNotifyIfNeeded(info successLoginData) error {
	sendingCtx, cancel := ctxWithTimeoutToSendMail()

	if needToNotify, err := needToNotifyAboutLoginFromNewDevice(sendingCtx, info, info.service); !needToNotify || err != nil {
		cancel()
		return err
	}

	var (
		userAgent    = useragent.Parse(info.request.UserAgent())
		notification = mailbuilder.NewNotificationData(userAgent, tryFindIP(info.request), info.loginTime)
		msg          = emailSender.newDeviceLoginMessage(info.user.Email, notification)
	)

	go emailSender.addToQueue(sendingCtx, cancel, info.request, info.user.Email, msg)
	return nil
}

func sendOAuthConnectionNotification(loginData successLoginData, oAuthType oauth.Type) error {
	sendingCtx, cancel := ctxWithTimeoutToSendMail()

	msg := emailSender.oAuthConnectionMessage(loginData.user.Email, oAuthType)
	go emailSender.addToQueue(sendingCtx, cancel, loginData.request, loginData.user.Email, msg)
	return nil
}

func oAuthRegisterTemplateOption(ctx context.Context, loginData successLoginData) (mailbuilder.RegisterTemplateOption, error) {
	if loginData.user.IsEmailConfirmed {
		return emailSender.registerWithoutConfirmationButton(), nil
	}

	url, err := confirmEmailURL(ctx, loginData.request, loginData.service, loginData.user.Email)
	if err != nil {
		return nil, err
	}

	return emailSender.registerWithConfirmationButton(url), nil
}

func needToNotifyAboutLoginFromNewDevice(ctx context.Context, info successLoginData, authService AuthService) (bool, error) {
	if !info.user.IsEmailConfirmed {
		return false, SendingError{"email address is not confirmed"}
	}

	isDeviceNew, err := authService.IsDeviceNewWhenUserHadSessionsBefore(ctx, info.newSession)
	if err != nil {
		sendingError := SendingError{"failed to detect if device is new or already used after successful login"}
		return false, errors.Join(sendingError, err)
	}

	if !isDeviceNew {
		return false, nil
	}

	return true, nil
}

func tryFindIP(r *http.Request) string {
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return ""
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return ""
	}

	return parsedIP.String()
}
