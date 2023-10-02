package googleoauthios

import (
	"context"
	"errors"
	"os"

	"google.golang.org/api/idtoken"

	"github.com/zhuboris/never-expires/internal/id/tkn"
	"github.com/zhuboris/never-expires/internal/id/usr/oauth"
)

type validator interface {
	Validate(ctx context.Context, idToken string, audience string) (*idtoken.Payload, error)
}

type Service struct {
	clientID  string
	validator validator
}

func NewService() (*Service, error) {
	const clientIDEnvKey = "IOS_OAUTH_CLIENT_ID"

	clientID := os.Getenv(clientIDEnvKey)
	if clientID == "" {
		return nil, errors.New("clientID env not set")
	}

	return &Service{
		clientID:  clientID,
		validator: idTokenValidator{},
	}, nil
}

func (s Service) UserFromToken(ctx context.Context, idToken oauth.Token) (oauth.User, error) {
	payload, err := s.validator.Validate(ctx, idToken.IDToken.TokenString, s.clientID)
	if err != nil {
		return oauth.User{}, errors.Join(tkn.ErrUnauthorized, oauth.ErrInvalidToken, err)
	}

	data, err := oauth.NewClaimsData(payload.Claims)
	if err != nil {
		return oauth.User{}, errors.Join(tkn.ErrUnauthorized, err)
	}

	return data.CastToUser()
}

type idTokenValidator struct{}

func (v idTokenValidator) Validate(ctx context.Context, idToken string, audience string) (*idtoken.Payload, error) {
	return idtoken.Validate(ctx, idToken, audience)
}
