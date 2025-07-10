package utils

import (
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var jwtKey []byte // This will be set from an environment variable

// SetJWTSecret initializes the JWT secret key.
func SetJWTSecret(secret []byte) {
	jwtKey = secret
}

// Claims defines the structure of the JWT payload.
type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT for the given email.
func GenerateToken(email string) (string, error) {
	expirationTime := time.Now().Add(72 * time.Hour)

	claims := &Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "rysto-auth-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// ValidateToken parses and validates a JWT string.
func ValidateToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return claims, nil
}

// ExtractToken safely parses the token from the Authorization header.
func ExtractToken(authHeader string) string {
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
