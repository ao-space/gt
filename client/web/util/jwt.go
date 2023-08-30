package util

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// TODO: change
const TokenExpireDuration = 6 * time.Hour

type JWT struct {
	SigningKey []byte
}

func NewJWT(signingKey string) *JWT {
	return &JWT{
		[]byte(signingKey),
	}
}

type CustomClaims struct {
	username string
	jwt.RegisteredClaims
}

func (j *JWT) CreateClaims(username string) CustomClaims {
	return CustomClaims{
		username: username,
		//BufferTime: int64(24 * time.Hour),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpireDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "gt",
			Subject:   "user token",
		},
	}
}

func (j *JWT) CreateToken(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}
