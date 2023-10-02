package session

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Session struct {
	ID         pgtype.UUID `json:"id"`
	UserID     pgtype.UUID `json:"user_id"`
	Device     string      `json:"device"`
	RefreshJWT string      `json:"refresh_jwt"`
}
