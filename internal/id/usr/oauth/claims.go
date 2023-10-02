package oauth

import (
	"errors"
	"strconv"
)

type ClaimsData struct {
	id            any
	email         any
	emailVerified any
	name          any
}

func NewClaimsData(claims map[string]any) (ClaimsData, error) {
	const (
		subKey           = "sub"
		emailKey         = "email"
		emailVerifiedKey = "email_verified"
		nameKey          = "name"
	)

	id, ok := claims[subKey]
	if !ok {
		return ClaimsData{}, errNoUserProfileData
	}

	email, ok := claims[emailKey]
	if !ok {
		return ClaimsData{}, errNoUserProfileData
	}

	emailVerified, ok := claims[emailVerifiedKey]
	if !ok {
		emailVerified = false
	}

	name, ok := claims[nameKey]
	if !ok {
		name = ""
	}

	return ClaimsData{
		id:            id,
		email:         email,
		emailVerified: emailVerified,
		name:          name,
	}, nil
}

func (c ClaimsData) CastToUser() (User, error) {
	id, ok := c.id.(string)
	if !ok {
		return User{}, errWrongTypeCasting
	}
	email, ok := c.email.(string)
	if !ok {
		return User{}, errWrongTypeCasting
	}

	isEmailVerified, err := c.isEmailVerified()
	if err != nil {
		return User{}, err
	}

	name, ok := c.name.(string)
	if !ok {
		return User{}, errWrongTypeCasting
	}

	return NewUser(id, email, name, isEmailVerified), nil
}

func (c ClaimsData) isEmailVerified() (bool, error) {
	isEmailVerified, ok := c.emailVerified.(bool)
	if ok {
		return isEmailVerified, nil
	}

	verified, ok := c.emailVerified.(string)
	if !ok {
		return false, errWrongTypeCasting
	}

	isEmailVerified, err := strconv.ParseBool(verified)
	if err != nil {
		return false, errors.Join(errWrongTypeCasting, err)
	}

	return isEmailVerified, nil
}
