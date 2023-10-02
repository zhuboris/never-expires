package usr

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Entity struct {
	ID               pgtype.UUID `json:"id"`
	Username         string      `json:"username"`
	Email            string      `json:"email"`
	Password         *string     `json:"password"`
	IsEmailConfirmed bool        `json:"is_email_confirmed"`
}

func (e Entity) ToUser() *User {
	user := new(User)
	user.ID = e.ID
	user.Username = e.Username
	user.Email = e.Email
	user.IsEmailConfirmed = e.IsEmailConfirmed
	if e.Password != nil {
		user.Password = *e.Password
	}

	return user
}

type User struct {
	ID               pgtype.UUID `json:"id"`
	Username         string      `json:"username"`
	Email            string      `json:"email"`
	Password         string      `json:"password"`
	IsEmailConfirmed bool        `json:"is_email_confirmed"`
}

type PublicData struct {
	Username         string `json:"username"`
	Email            string `json:"email"`
	IsEmailConfirmed bool   `json:"is_email_confirmed"`
}

func (u User) PublicData() *PublicData {
	return &PublicData{
		Username:         u.Username,
		Email:            u.Email,
		IsEmailConfirmed: u.IsEmailConfirmed,
	}
}
