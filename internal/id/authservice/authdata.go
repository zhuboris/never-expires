package authservice

import (
	"github.com/google/uuid"

	"github.com/zhuboris/never-expires/internal/id/session"
)

type (
	AuthData struct {
		accessJWT  string
		refreshJWT string
		newSession session.Session
	}
	AuthResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		SessionID    string `json:"session_id"`
	}
)

func (d AuthData) ToResponse() AuthResponse {
	id := uuid.UUID(d.newSession.ID.Bytes)

	return AuthResponse{
		AccessToken:  d.accessJWT,
		RefreshToken: d.refreshJWT,
		SessionID:    id.String(),
	}
}

func (d AuthData) AccessToken() string {
	return d.accessJWT
}

func (d AuthData) RefreshToken() string {
	return d.refreshJWT
}

func (d AuthData) Session() session.Session {
	return d.newSession
}
