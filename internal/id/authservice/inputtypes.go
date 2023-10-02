package authservice

import (
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zhuboris/never-expires/internal/id/usr/oauth"
)

type (
	ChangeMailData struct {
		NewEmail string `json:"new_email"`
		Password string `json:"password"`
	}
	ChangePasswordData struct {
		Current string `json:"current_password"`
		New     string `json:"new_password"`
	}
	ChangeUsernameData struct {
		New string `json:"new_username"`
	}
	LoginData struct {
		Email      string `json:"email"`
		Password   string `json:"password"`
		userDevice string
	}
	LoginWithOAuthData struct {
		user       oauth.User
		userDevice string
	}
	RegisterData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Username string `json:"username"`
	}
	SendRestorePasswordEmailData struct {
		Email string `json:"email"`
	}
	RefreshJWTData struct {
		RefreshToken string `json:"refresh_token"`
		SessionID    string `json:"session_id"`
	}
	LogoutData struct {
		sessionID pgtype.UUID
	}
)

func NewLoginData(device string) LoginData {
	return LoginData{
		userDevice: device,
	}
}

func NewLoginWithOAuthData(user oauth.User, device string) LoginWithOAuthData {
	return LoginWithOAuthData{
		user:       user,
		userDevice: device,
	}
}

func NewLogoutData(sessionID pgtype.UUID) LogoutData {
	return LogoutData{
		sessionID: sessionID,
	}
}

func (d ChangeMailData) IsMissingRequiredField() bool {
	return d.NewEmail == "" || d.Password == ""
}

func (d ChangePasswordData) IsMissingRequiredField() bool {
	return d.New == ""
}

func (d ChangeUsernameData) IsMissingRequiredField() bool {
	return d.New == ""
}

func (d LoginData) IsMissingRequiredField() bool {
	return d.Email == "" || d.Password == ""
}

func (d RegisterData) IsMissingRequiredField() bool {
	return d.Email == "" || d.Password == ""
}
func (d SendRestorePasswordEmailData) IsMissingRequiredField() bool {
	return d.Email == ""
}
