package tkn

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const secretEnvKey = "JWT_SECRET_KEY"

func CreateJWT(userID pgtype.UUID, expires time.Duration) (string, error) {
	claims := &UserClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: expiresIn(expires),
		},
	}

	var (
		token            = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		jwtSecretKey, ok = os.LookupEnv(secretEnvKey)
	)

	if !ok {
		return "", errEnvMissingKey
	}

	return token.SignedString([]byte(jwtSecretKey))
}

func expiresIn(timeLeft time.Duration) *jwt.NumericDate {
	return jwt.NewNumericDate(time.Now().Add(timeLeft))
}
