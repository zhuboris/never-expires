package oauth

import (
	"strings"
)

type User struct {
	id              string
	email           string
	isEmailVerified bool
	name            string
	refreshToken    string
}

func NewUser(id, email, name string, isEmailVerified bool) User {
	return User{
		id:              id,
		email:           email,
		isEmailVerified: isEmailVerified,
		name:            name,
	}
}

func (u *User) SetRefreshToken(token string) {
	u.refreshToken = token
}

func (u *User) UpdateEmailToLower() {
	u.email = strings.ToLower(u.email)
}

func (u *User) ID() string {
	return u.id
}

func (u *User) Email() string {
	return u.email
}

func (u *User) IsEmailVerified() bool {
	return u.isEmailVerified
}

func (u *User) Name() string {
	return u.name
}

func (u *User) RefreshToken() string {
	return u.refreshToken
}
