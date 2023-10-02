package authservice

import (
	"github.com/zhuboris/never-expires/internal/id/usr"
	"github.com/zhuboris/never-expires/internal/id/usr/oauth"
)

type LoginResult struct {
	user            *usr.User
	authData        AuthData
	oauthResultType oauth.LoginResultType
	err             error
}

func newLoginResult(user *usr.User, authData AuthData, err error) LoginResult {
	return LoginResult{
		user:     user,
		authData: authData,
		err:      err,
	}
}

func newLoginWithOAuthResult(user *usr.User, authData AuthData, resultType oauth.LoginResultType, err error) LoginResult {
	return LoginResult{
		user:            user,
		authData:        authData,
		oauthResultType: resultType,
		err:             err,
	}
}

func newErrorLoginResult(err error) LoginResult {
	return LoginResult{
		err: err,
	}
}

func (r LoginResult) AuthData() AuthData {
	return r.authData
}

func (r LoginResult) UserData() *usr.PublicData {
	return r.user.PublicData()
}

func (r LoginResult) OAuthResultType() oauth.LoginResultType {
	return r.oauthResultType
}

func (r LoginResult) Error() error {
	return r.err
}
