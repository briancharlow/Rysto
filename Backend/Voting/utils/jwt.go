package utils

import (
	"github.com/golang-jwt/jwt/v4"
)

var jwtKey []byte

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func SetJWTSecret(secret []byte) {
	jwtKey = secret
}

func ValidateToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return claims, nil
}
