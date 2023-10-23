package util

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/isrc-cas/gt/predef"
	"github.com/isrc-cas/gt/web/server/model/request"
	"time"
)

type JWT struct {
	SigningKey          []byte
	TokenExpireDuration time.Duration
}

func NewJWT(signingKey string, tokenExpireDuration time.Duration) *JWT {
	if tokenExpireDuration == 0 {
		tokenExpireDuration = predef.DefaultTokenDuration
	}
	return &JWT{
		SigningKey:          []byte(signingKey),
		TokenExpireDuration: tokenExpireDuration,
	}
}

type CustomClaims struct {
	Username string
	jwt.RegisteredClaims
}

func (j *JWT) CreateClaims(username string, issuer string) CustomClaims {
	return CustomClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.TokenExpireDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
			Subject:   "USER TOKEN",
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
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}

func GenerateToken(signingKey string, expireDuration time.Duration, issuer string, user request.User) (token string, err error) {
	j := NewJWT(signingKey, expireDuration)
	claims := j.CreateClaims(user.Username, issuer)
	token, err = j.CreateToken(claims)
	if err != nil {
		return "", err
	}
	return token, nil
}
