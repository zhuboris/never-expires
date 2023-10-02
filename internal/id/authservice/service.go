package authservice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zhuboris/never-expires/internal/id/pw"
	"github.com/zhuboris/never-expires/internal/id/session"
	"github.com/zhuboris/never-expires/internal/id/tkn"
	"github.com/zhuboris/never-expires/internal/id/usr"
	"github.com/zhuboris/never-expires/internal/id/usr/oauth"
	"github.com/zhuboris/never-expires/internal/shared/postgresql"
)

type (
	UserService interface {
		Register(ctx context.Context, user usr.User) (*usr.User, error)
		PublicDataByUserCtx(ctx context.Context) (*usr.PublicData, error)
		UserByEmail(ctx context.Context, email string) (*usr.User, error)
		Delete(ctx context.Context) error
		CheckPassword(ctx context.Context, toCheck string) error
		Contains(ctx context.Context, email string) error
		UpdatePassword(ctx context.Context, new string) error
		UpdateEmail(ctx context.Context, new string) error
		UpdateUsername(ctx context.Context, new string) error
		ConfirmEmail(ctx context.Context, token string) error
		ResetPassword(ctx context.Context, validationToken string) (user *usr.User, newPassword string, err error)
		IsConfirmed(ctx context.Context, email string) (bool, error)
		UserFromGoogleIDToken(ctx context.Context, idToken oauth.Token) (oauth.User, error)
		UserFromAppleTokenCode(ctx context.Context, idToken oauth.Token) (oauth.User, error)
		LoginWithOAuthGoogle(ctx context.Context, user oauth.User) (*usr.User, oauth.LoginResultType, error)
		LoginWithApple(ctx context.Context, user oauth.User) (*usr.User, oauth.LoginResultType, error)
		TryRevokeAppleAccount(ctx context.Context) error
		AddEmailConfirmationToken(ctx context.Context, email string, tokenLifetime time.Duration) (string, error)
		AddPasswordResetToken(ctx context.Context, email string, tokenLifetime time.Duration) (string, error)
		Status(ctx context.Context) error
	}
	SessionService interface {
		Add(ctx context.Context, session session.Session) (pgtype.UUID, error)
		Deactivate(ctx context.Context, sessionID pgtype.UUID) error
		DeactivateAll(ctx context.Context) error
		Contains(ctx context.Context, session session.Session) error
		IsDeviceNewWhenUserHadSessionsBefore(ctx context.Context, session session.Session) (bool, error)
		Status(ctx context.Context) error
	}
)

type AuthService struct {
	userService    UserService
	sessionService SessionService
}

func New(userService UserService, sessionService SessionService) *AuthService {
	return &AuthService{
		userService:    userService,
		sessionService: sessionService,
	}
}

func (s AuthService) UserFromGoogleIDToken(ctx context.Context, idToken oauth.Token) (oauth.User, error) {
	return s.userService.UserFromGoogleIDToken(ctx, idToken)
}

func (s AuthService) UserFromAppleTokenCode(ctx context.Context, idToken oauth.Token) (oauth.User, error) {
	return s.userService.UserFromAppleTokenCode(ctx, idToken)
}

func (s AuthService) AuthorizedUser(ctx context.Context) GettingUserResult {
	user, err := s.userService.PublicDataByUserCtx(ctx)
	return newGettingUserResult(user, err)
}

func (s AuthService) DeleteUser(ctx context.Context) error {
	err := s.userService.TryRevokeAppleAccount(ctx)
	if err != nil {
		return err
	}

	return s.userService.Delete(ctx)
}

func (s AuthService) UpdateMail(ctx context.Context, data ChangeMailData) error {
	if err := s.userService.CheckPassword(ctx, data.Password); err != nil {
		return err
	}

	if err := s.sessionService.DeactivateAll(ctx); err != nil {
		return err
	}

	err := s.userService.UpdateEmail(ctx, data.NewEmail)
	if errors.Is(err, postgresql.ErrAddedDuplicateOfUnique) {
		err = errors.Join(ErrAlreadyRegistered, err)
	}

	return err
}

func (s AuthService) ChangePassword(ctx context.Context, data ChangePasswordData) error {
	if err := s.userService.CheckPassword(ctx, data.Current); err != nil {
		return err
	}

	if err := s.sessionService.DeactivateAll(ctx); err != nil {
		return err
	}

	return s.userService.UpdatePassword(ctx, data.New)
}

func (s AuthService) ChangeUsername(ctx context.Context, input ChangeUsernameData) error {
	return s.userService.UpdateUsername(ctx, input.New)
}

func (s AuthService) Login(ctx context.Context, data LoginData) LoginResult {
	if data.userDevice == "" {
		return newErrorLoginResult(errMissingUserDevice)
	}

	user, err := s.userService.UserByEmail(ctx, data.Email)
	if err != nil {
		return newErrorLoginResult(errors.Join(ErrWrongLoginData, err))
	}

	if err := pw.Check(data.Password, user.Password); err != nil {
		return newErrorLoginResult(errors.Join(ErrWrongLoginData, err))
	}

	authData, err := s.CreateSession(ctx, user.ID, data.userDevice)
	return newLoginResult(user, authData, err)
}

func (s AuthService) LoginWithOAuth(ctx context.Context, data LoginWithOAuthData, loginOptionFunc OAuthOption) LoginResult {
	if data.userDevice == "" {
		return newErrorLoginResult(errMissingUserDevice)
	}

	user, resultType, err := loginOptionFunc(ctx, data.user)
	if err != nil {
		return newErrorLoginResult(err)
	}

	authData, err := s.CreateSession(ctx, user.ID, data.userDevice)
	return newLoginWithOAuthResult(user, authData, resultType, err)
}

func (s AuthService) CreateSession(ctx context.Context, userID pgtype.UUID, userDevice string) (AuthData, error) {
	if !userID.Valid {
		return AuthData{}, ErrWrongLoginData
	}

	if userDevice == "" {
		return AuthData{}, errMissingUserDevice
	}

	authToken, err := tkn.CreateJWT(userID, AuthLifetime)
	if err != nil {
		return AuthData{}, err
	}

	refreshToken, err := tkn.CreateJWT(userID, RefreshLifetime)
	if err != nil {
		return AuthData{}, err
	}

	newSession := session.Session{
		UserID:     userID,
		Device:     userDevice,
		RefreshJWT: refreshToken,
	}

	newSessionID, err := s.sessionService.Add(ctx, newSession)
	if err != nil {
		return AuthData{}, err
	}

	newSession.ID = newSessionID
	return AuthData{
		accessJWT:  authToken,
		refreshJWT: refreshToken,
		newSession: newSession,
	}, nil
}

func (s AuthService) Logout(ctx context.Context, data LogoutData) error {
	return s.sessionService.Deactivate(ctx, data.sessionID)
}

func (s AuthService) AllowRefreshingJWT(ctx context.Context, currentSession session.Session) error {
	if err := s.sessionService.Contains(ctx, currentSession); err != nil {
		return fmt.Errorf("%w: session is not exist: %w", tkn.ErrUnauthorized, err)
	}

	return nil
}

func (s AuthService) DeactivateSession(ctx context.Context, sessionID pgtype.UUID) error {
	return s.sessionService.Deactivate(ctx, sessionID)
}

func (s AuthService) ConfirmEmail(ctx context.Context, token string) error {
	return s.userService.ConfirmEmail(ctx, token)
}

func (s AuthService) ResetPassword(ctx context.Context, token string) ResetPasswordResult {
	user, newPassword, err := s.userService.ResetPassword(ctx, token)
	if err != nil {
		return newResetPasswordResult("", "", err)
	}

	ctx = usr.WithUserID(ctx, user.ID)
	if err := s.sessionService.DeactivateAll(ctx); err != nil {
		return newResetPasswordResult("", "", err)
	}

	return newResetPasswordResult(user.Email, newPassword, err)
}

func (s AuthService) IsConfirmed(ctx context.Context, email string) (bool, error) {
	return s.userService.IsConfirmed(ctx, email)
}

func (s AuthService) Register(ctx context.Context, data RegisterData) RegisterResult {
	if err := s.checkIfEmailAlreadyRegistered(ctx, data.Email); err != nil {
		return newRegisterResult(nil, err)
	}

	newUser := usr.User{
		Email:    data.Email,
		Password: data.Password,
		Username: data.Username,
	}

	user, err := s.userService.Register(ctx, newUser)
	return newRegisterResult(user, err)
}

func (s AuthService) IsDeviceNewWhenUserHadSessionsBefore(ctx context.Context, currentSession session.Session) (bool, error) {
	return s.sessionService.IsDeviceNewWhenUserHadSessionsBefore(ctx, currentSession)
}

func (s AuthService) CheckIfUserExists(ctx context.Context, emailAddress string) error {
	if err := s.userService.Contains(ctx, emailAddress); err != nil {
		return ErrMissingEmailAddress
	}

	return nil
}

func (s AuthService) AddEmailConfirmationToken(ctx context.Context, email string) (string, error) {
	const lifetime = 24 * time.Hour

	return s.userService.AddEmailConfirmationToken(ctx, email, lifetime)
}

func (s AuthService) AddPasswordResetToken(ctx context.Context, email string) (string, error) {
	const lifetime = time.Hour

	return s.userService.AddPasswordResetToken(ctx, email, lifetime)
}

func (s AuthService) Status(ctx context.Context) error {
	userErr := s.userService.Status(ctx)
	sessionsErr := s.sessionService.Status(ctx)
	return errors.Join(userErr, sessionsErr)
}

func (s AuthService) checkIfEmailAlreadyRegistered(ctx context.Context, email string) error {
	err := s.userService.Contains(ctx, email)

	switch {
	case errors.Is(err, postgresql.ErrNoMatches):
		return nil
	case err != nil:
		return err
	default:
		return ErrAlreadyRegistered
	}
}
