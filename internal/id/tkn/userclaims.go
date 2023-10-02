package tkn

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserClaims struct {
	UserID pgtype.UUID `json:"user_id"`

	jwt.RegisteredClaims
}
