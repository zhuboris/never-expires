package authservice

import (
	"github.com/zhuboris/never-expires/internal/id/usr"
)

type GettingUserResult struct {
	user *usr.PublicData
	err  error
}

func newGettingUserResult(user *usr.PublicData, err error) GettingUserResult {
	return GettingUserResult{
		user: user,
		err:  err,
	}
}

func (r GettingUserResult) UserData() *usr.PublicData {
	return r.user
}

func (r GettingUserResult) Error() error {
	return r.err
}

type RegisterResult struct {
	user *usr.User
	err  error
}

func newRegisterResult(user *usr.User, err error) RegisterResult {
	return RegisterResult{
		user: user,
		err:  err,
	}
}

func (r RegisterResult) User() *usr.User {
	return r.user
}

func (r RegisterResult) Error() error {
	return r.err
}

type ResetPasswordResult struct {
	emailAddress string
	newPassword  string
	err          error
}

func newResetPasswordResult(emailAddress, newPassword string, err error) ResetPasswordResult {
	return ResetPasswordResult{
		emailAddress: emailAddress,
		newPassword:  newPassword,
		err:          err,
	}
}

func (r ResetPasswordResult) Address() string {
	return r.emailAddress
}

func (r ResetPasswordResult) NewPassword() string {
	return r.newPassword
}

func (r ResetPasswordResult) Error() error {
	return r.err
}
