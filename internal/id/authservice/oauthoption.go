package authservice

import (
	"context"

	"github.com/zhuboris/never-expires/internal/id/usr"
	"github.com/zhuboris/never-expires/internal/id/usr/oauth"
)

type OAuthOption func(ctx context.Context, user oauth.User) (*usr.User, oauth.LoginResultType, error)

func (s AuthService) WithGoogle() OAuthOption {
	return func(ctx context.Context, user oauth.User) (*usr.User, oauth.LoginResultType, error) {
		return s.userService.LoginWithOAuthGoogle(ctx, user)
	}
}

func (s AuthService) WithApple() OAuthOption {
	return func(ctx context.Context, user oauth.User) (*usr.User, oauth.LoginResultType, error) {
		return s.userService.LoginWithApple(ctx, user)
	}
}
