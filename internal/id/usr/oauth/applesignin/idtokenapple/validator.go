package idtokenapple

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type (
	keysResponse struct {
		Keys []key `json:"keys"`
	}
	key struct {
		Kid string `json:"kid"`
		N   string `json:"n"`
		E   string `json:"e"`
	}
)

const (
	issuer  = "https://appleid.apple.com"
	keysURL = "https://appleid.apple.com/auth/keys"
)

type Validator struct {
	audience string
}

func NewValidator(audience string) Validator {
	return Validator{
		audience: audience,
	}
}

func (v Validator) Validate(idToken string) (jwt.MapClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(idToken, jwt.MapClaims{})
	if err != nil {
		return jwt.MapClaims{}, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return jwt.MapClaims{}, ErrInvalidRequiredField
	}

	if err := validateClaims(claims, v.audience); err != nil {
		return jwt.MapClaims{}, err
	}

	kid, ok := token.Header["kid"].(string)
	if !ok {
		return jwt.MapClaims{}, ErrInvalidRequiredField
	}

	appleKey, err := fetchApplePublicKey(kid)
	if err != nil {
		return jwt.MapClaims{}, err
	}

	keyFunc := func(token *jwt.Token) (any, error) {
		return appleKey, nil
	}

	token, err = jwt.Parse(idToken, keyFunc)
	if err != nil {
		return jwt.MapClaims{}, errors.Join(ErrWrongKey, err)
	}

	claims, ok = token.Claims.(jwt.MapClaims)
	if !ok {
		return jwt.MapClaims{}, ErrInvalidRequiredField
	}

	return claims, nil
}

func validateClaims(claims jwt.MapClaims, audience string) error {
	if claims["iss"] != issuer {
		return ErrWrongIssuer
	}

	if claims["aud"] != audience {
		return ErrWrongAudience
	}

	if err := checkExpiration(claims); err != nil {
		return err
	}

	return nil
}

func fetchApplePublicKey(kid string) (*rsa.PublicKey, error) {
	resp, err := http.Get(keysURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var keys keysResponse
	err = json.NewDecoder(resp.Body).Decode(&keys)
	if err != nil {
		return nil, err
	}

	return matchPublicKey(kid, keys)
}

func checkExpiration(claims jwt.MapClaims) error {
	expRaw, ok := claims["exp"]
	if !ok {
		return ErrInvalidRequiredField
	}

	exp := expRaw.(float64)
	if exp <= 0 {
		return ErrInvalidRequiredField
	}

	if int64(exp) < time.Now().Unix() {
		return ErrExpired
	}

	return nil
}

func matchPublicKey(kid string, keys keysResponse) (*rsa.PublicKey, error) {
	for _, k := range keys.Keys {
		if k.Kid == kid {
			return parse(k)
		}
	}

	return nil, ErrKeyNotFound
}

func parse(key key) (*rsa.PublicKey, error) {
	n, err := decode(key.N)
	if err != nil {
		return nil, err
	}

	e, err := decode(key.E)
	if err != nil {
		return nil, err
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(n),
		E: int(new(big.Int).SetBytes(e).Int64()),
	}, nil
}

func decode(input string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(input)
}
