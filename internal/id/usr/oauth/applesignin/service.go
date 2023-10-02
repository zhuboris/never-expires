package applesignin

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/Timothylock/go-signin-with-apple/apple"
	"github.com/golang-jwt/jwt/v5"

	"github.com/zhuboris/never-expires/internal/id/tkn"
	"github.com/zhuboris/never-expires/internal/id/usr/oauth"
	"github.com/zhuboris/never-expires/internal/id/usr/oauth/applesignin/idtokenapple"
)

type validator interface {
	Validate(idToken string) (jwt.MapClaims, error)
}

type Service struct {
	client       *apple.Client
	clientSecret string
	clientID     string
	validator    validator
}

func NewService() (*Service, error) {
	envs, err := newConfig()
	if err != nil {
		return nil, err
	}

	secret, err := apple.GenerateClientSecret(envs.privateKey, envs.teamID, envs.bundleID, envs.keyID)
	if err != nil {
		return nil, fmt.Errorf("error generating secret: %w", err)
	}

	return &Service{
		client:       apple.New(),
		clientSecret: secret,
		clientID:     envs.bundleID,
		validator:    idtokenapple.NewValidator(envs.bundleID),
	}, nil
}

func (s Service) UserFromToken(ctx context.Context, token oauth.Token) (oauth.User, error) {
	claims, err := s.validator.Validate(token.IDToken.TokenString)
	if err != nil {
		return oauth.User{}, errors.Join(tkn.ErrUnauthorized, oauth.ErrInvalidToken, err)
	}

	data, err := oauth.NewClaimsData(claims)
	if err != nil {
		return oauth.User{}, errors.Join(tkn.ErrUnauthorized, err)
	}

	user, err := data.CastToUser()
	if err != nil {
		return oauth.User{}, err
	}

	refreshToken, err := s.refreshTokenFromAuthCode(ctx, token.AuthCode)
	if err != nil {
		return oauth.User{}, err
	}

	user.SetRefreshToken(refreshToken)
	return user, nil
}

func (s Service) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	req := apple.RevokeRefreshTokenRequest{
		ClientID:     s.clientID,
		ClientSecret: s.clientSecret,
		RefreshToken: refreshToken,
	}

	var resp apple.RevokeResponse
	err := s.client.RevokeRefreshToken(ctx, req, &resp)

	return checkRevokeRespForErrors(err, resp)
}

func (s Service) refreshTokenFromAuthCode(ctx context.Context, authCode string) (string, error) {
	req := apple.AppValidationTokenRequest{
		ClientID:     s.clientID,
		ClientSecret: s.clientSecret,
		Code:         authCode,
	}

	var resp apple.ValidationResponse
	err := s.client.VerifyAppToken(ctx, req, &resp)
	if err := checkCodeValidationRespForErrors(err, resp); err != nil {
		return "", fmt.Errorf("error validating")
	}

	return resp.RefreshToken, nil
}

func checkCodeValidationRespForErrors(err error, resp apple.ValidationResponse) error {
	switch {
	case resp.Error == "" && errors.Is(err, io.EOF):
		return nil
	case err != nil || resp.Error != "":
		return fmt.Errorf(" apple returned an error: error : %w, response error: %s, description: %s", err, resp.Error, resp.ErrorDescription)
	case resp.RefreshToken == "":
		return errors.New("apple provided empty refresh token without error")
	default:
		return nil
	}
}

func checkRevokeRespForErrors(err error, resp apple.RevokeResponse) error {
	switch {
	case resp.Error == "" && errors.Is(err, io.EOF):
		return nil
	case err != nil || resp.Error != "":
		return fmt.Errorf(" apple returned an error: error : %w, response error: %s, description: %s", err, resp.Error, resp.ErrorDescription)
	default:
		return nil
	}
}
