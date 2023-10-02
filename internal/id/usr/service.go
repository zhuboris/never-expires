package usr

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zhuboris/never-expires/internal/id/pw"
	"github.com/zhuboris/never-expires/internal/id/usr/oauth"
	"github.com/zhuboris/never-expires/internal/shared/postgresql"
	"github.com/zhuboris/never-expires/internal/shared/servicechecker"
)

type (
	repository interface {
		addByPassword(ctx context.Context, user User) (*User, error)
		byID(ctx context.Context, id pgtype.UUID) (*User, error)
		byEmail(ctx context.Context, email string) (*User, error)
		delete(ctx context.Context, id pgtype.UUID) error
		encryptedPassword(ctx context.Context, userID pgtype.UUID) (string, error)
		updateColumn(ctx context.Context, userID pgtype.UUID, toUpdate column, new string) error
		validateEmail(ctx context.Context, validationToken string) error
		restorePassword(ctx context.Context, validationToken, newPassword string) (userEmail string, err error)
		isConfirmed(ctx context.Context, email string) (bool, error)
		byOAuth(ctx context.Context, userInputted oauth.User, oAuthServiceOption oAuthMethod) (*User, oauth.LoginResultType, error)
		saveAppleRefreshToken(ctx context.Context, userID pgtype.UUID, token string) error
		allAppleRefreshTokens(ctx context.Context, userID pgtype.UUID) ([]string, error)
		addEmailConfirmationToken(ctx context.Context, email string, tempToken ConfirmationToken) error
		addPasswordResetToken(ctx context.Context, email string, tempToken ConfirmationToken) error
		Ping(ctx context.Context) error
	}
	GoogleOAuthService interface {
		UserFromToken(ctx context.Context, idToken oauth.Token) (oauth.User, error)
	}
	AppleOAuthService interface {
		UserFromToken(ctx context.Context, idToken oauth.Token) (oauth.User, error)
		RevokeRefreshToken(ctx context.Context, refreshToken string) error
	}
)

type Service struct {
	googleOauth  GoogleOAuthService
	appleSignIn  AppleOAuthService
	repo         repository
	statusMetric servicechecker.StatusDisplay
}

func NewService(repo repository, googleOauth GoogleOAuthService, appleSignIn AppleOAuthService, statusDisplay servicechecker.StatusDisplay) *Service {
	return &Service{
		repo:         repo,
		googleOauth:  googleOauth,
		appleSignIn:  appleSignIn,
		statusMetric: statusDisplay,
	}
}

func (s Service) Register(ctx context.Context, user User) (*User, error) {
	address, err := parseAddress(user.Email)
	if err != nil {
		return nil, errors.Join(ErrInvalidEmail, err)
	}

	user.Email = strings.ToLower(address.Address)
	if err := pw.Validate(user.Password); err != nil {
		return nil, fmt.Errorf("invalid password: %w", err)
	}

	encryptedPassword, err := pw.Hash(user.Password)
	if err != nil {
		return nil, err
	}

	user.Password = encryptedPassword

	if user.Username == "" {
		user.Username = address.Name
	}

	return s.repo.addByPassword(ctx, user)
}

func (s Service) UserByEmail(ctx context.Context, email string) (*User, error) {
	email = strings.ToLower(email)
	return s.repo.byEmail(ctx, email)
}

func (s Service) PublicDataByUserCtx(ctx context.Context) (*PublicData, error) {
	id, err := ID(ctx)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.byID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &PublicData{
		Username:         user.Username,
		Email:            user.Email,
		IsEmailConfirmed: user.IsEmailConfirmed,
	}, nil
}

func (s Service) Delete(ctx context.Context) error {
	id, err := ID(ctx)
	if err != nil {
		return err
	}

	return s.repo.delete(ctx, id)
}

func (s Service) CheckPassword(ctx context.Context, toCheck string) error {
	id, err := ID(ctx)
	if err != nil {
		return err
	}

	storedPassword, err := s.repo.encryptedPassword(ctx, id)
	if toCheck == "" && errors.Is(err, pgx.ErrNoRows) {
		return nil
	}

	if err != nil {
		return err
	}

	return pw.Check(toCheck, storedPassword)
}

func (s Service) Contains(ctx context.Context, email string) error {
	email = strings.ToLower(email)
	_, err := s.repo.byEmail(ctx, email)
	return err
}

func (s Service) UpdatePassword(ctx context.Context, new string) error {
	userID, err := ID(ctx)
	if err != nil {
		return err
	}

	if err := pw.Validate(new); err != nil {
		return err
	}
	encryptedPassword, err := pw.Hash(new)
	if err != nil {
		return err
	}

	new = encryptedPassword
	err = s.repo.updateColumn(ctx, userID, password, new)
	if errors.Is(err, postgresql.ErrNoMatches) {
		err = errors.Join(ErrNotFound, err)
	}

	return err
}

func (s Service) UpdateEmail(ctx context.Context, new string) error {
	userID, err := ID(ctx)
	if err != nil {
		return err
	}
	newAddress, err := parseAddress(new)
	if err != nil {
		return errors.Join(ErrInvalidEmail, err)
	}

	new = strings.ToLower(newAddress.Address)
	err = s.repo.updateColumn(ctx, userID, email, new)
	if errors.Is(err, postgresql.ErrNoMatches) {
		err = errors.Join(ErrNotFound, err)
	}

	return err
}

func (s Service) UpdateUsername(ctx context.Context, new string) error {
	userID, err := ID(ctx)
	if err != nil {
		return err
	}

	err = s.repo.updateColumn(ctx, userID, username, new)
	if errors.Is(err, postgresql.ErrNoMatches) {
		err = errors.Join(ErrNotFound, err)
	}

	return err
}

func (s Service) ConfirmEmail(ctx context.Context, validationToken string) error {
	if err := checkIfTokenNotEmpty(validationToken); err != nil {
		return errors.Join(ErrValidationRefused, err)
	}

	if err := s.repo.validateEmail(ctx, validationToken); err != nil {
		return errors.Join(ErrValidationRefused, err)
	}

	return nil
}

func (s Service) ResetPassword(ctx context.Context, validationToken string) (user *User, newPassword string, err error) {
	const newPasswordLength = 12

	if err := checkIfTokenNotEmpty(validationToken); err != nil {
		return nil, "", errors.Join(ErrPasswordResetRefused, err)
	}

	newPassword = pw.Generate(newPasswordLength)
	encryptedPassword, err := pw.Hash(newPassword)
	if err != nil {
		return nil, "", err
	}

	email, err := s.repo.restorePassword(ctx, validationToken, encryptedPassword)
	if err != nil {
		return nil, "", errors.Join(ErrPasswordResetRefused, err)
	}

	user, err = s.UserByEmail(ctx, email)
	return user, newPassword, err

}

func (s Service) IsConfirmed(ctx context.Context, email string) (bool, error) {
	email = strings.ToLower(email)
	return s.repo.isConfirmed(ctx, email)
}

func (s Service) UserFromGoogleIDToken(ctx context.Context, idToken oauth.Token) (oauth.User, error) {
	return s.googleOauth.UserFromToken(ctx, idToken)
}

func (s Service) UserFromAppleTokenCode(ctx context.Context, idToken oauth.Token) (oauth.User, error) {
	return s.appleSignIn.UserFromToken(ctx, idToken)
}

func (s Service) LoginWithOAuthGoogle(ctx context.Context, user oauth.User) (*User, oauth.LoginResultType, error) {
	user.UpdateEmailToLower()
	return s.repo.byOAuth(ctx, user, withGoogle())
}

func (s Service) LoginWithApple(ctx context.Context, user oauth.User) (*User, oauth.LoginResultType, error) {
	user.UpdateEmailToLower()
	loggedUser, resultType, err := s.repo.byOAuth(ctx, user, withApple())
	if err != nil {
		return nil, resultType, err
	}

	err = s.repo.saveAppleRefreshToken(ctx, loggedUser.ID, user.RefreshToken())
	return loggedUser, resultType, err
}

func (s Service) TryRevokeAppleAccount(ctx context.Context) error {
	id, err := ID(ctx)
	if err != nil {
		return err
	}

	tokens, err := s.repo.allAppleRefreshTokens(ctx, id)
	if len(tokens) == 0 {
		return err
	}

	for _, token := range tokens {
		revokeErr := s.appleSignIn.RevokeRefreshToken(ctx, token)
		if revokeErr != nil {
			err = errors.Join(revokeErr, err)
		}
	}

	return err
}

func (s Service) AddEmailConfirmationToken(ctx context.Context, email string, tokenLifetime time.Duration) (string, error) {
	token, err := newConfirmationToken(tokenLifetime)
	if err != nil {
		return "", err
	}

	return token.Value, s.repo.addEmailConfirmationToken(ctx, email, token)
}

func (s Service) AddPasswordResetToken(ctx context.Context, email string, tokenLifetime time.Duration) (string, error) {
	token, err := newConfirmationToken(tokenLifetime)
	if err != nil {
		return "", err
	}

	return token.Value, s.repo.addPasswordResetToken(ctx, email, token)
}

func (s Service) Status(ctx context.Context) error {
	return servicechecker.Ping(ctx, s.repo, s.statusMetric, "userRepository")
}

func checkIfTokenNotEmpty(input string) error {
	if input == "" {
		return ErrMissingConfirmToken
	}

	return nil
}
