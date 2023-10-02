package request

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zhuboris/never-expires/internal/id/authservice"
	"github.com/zhuboris/never-expires/internal/id/session"
	"github.com/zhuboris/never-expires/internal/id/usr/oauth"
)

type AuthService interface {
	UserFromGoogleIDToken(ctx context.Context, idToken oauth.Token) (oauth.User, error)
	UserFromAppleTokenCode(ctx context.Context, idToken oauth.Token) (oauth.User, error)
	AuthorizedUser(ctx context.Context) authservice.GettingUserResult
	DeleteUser(ctx context.Context) error
	UpdateMail(ctx context.Context, data authservice.ChangeMailData) error
	ChangePassword(ctx context.Context, data authservice.ChangePasswordData) error
	ChangeUsername(ctx context.Context, input authservice.ChangeUsernameData) error
	Login(ctx context.Context, data authservice.LoginData) authservice.LoginResult
	CreateSession(ctx context.Context, userID pgtype.UUID, userDevice string) (authservice.AuthData, error)
	Logout(ctx context.Context, data authservice.LogoutData) error
	AllowRefreshingJWT(ctx context.Context, currentSession session.Session) error
	DeactivateSession(ctx context.Context, sessionID pgtype.UUID) error
	ConfirmEmail(ctx context.Context, token string) error
	ResetPassword(ctx context.Context, token string) authservice.ResetPasswordResult
	Register(ctx context.Context, data authservice.RegisterData) authservice.RegisterResult
	IsDeviceNewWhenUserHadSessionsBefore(ctx context.Context, currentSession session.Session) (bool, error)
	AddEmailConfirmationToken(ctx context.Context, email string) (string, error)
	AddPasswordResetToken(ctx context.Context, email string) (string, error)
	CheckIfUserExists(ctx context.Context, emailAddress string) error
	IsConfirmed(ctx context.Context, email string) (bool, error)
	LoginWithOAuth(ctx context.Context, data authservice.LoginWithOAuthData, option authservice.OAuthOption) authservice.LoginResult
	Status(ctx context.Context) error
	WithGoogle() authservice.OAuthOption
	WithApple() authservice.OAuthOption
}
