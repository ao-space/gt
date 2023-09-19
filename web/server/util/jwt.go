package util

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/isrc-cas/gt/web/server/model/request"
	"time"
)

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
	Username string
	jwt.RegisteredClaims
}

func (j *JWT) CreateClaims(username string, issuer string) CustomClaims {
	return CustomClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpireDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
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
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}

func GenerateToken(signingKey string, issuer string, user request.User) (token string, err error) {
	j := NewJWT(signingKey)
	claims := j.CreateClaims(user.Username, issuer)
	token, err = j.CreateToken(claims)
	if err != nil {
		return "", err
	}
	return token, nil
}
