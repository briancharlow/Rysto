package utils

import (
    "time"

    "github.com/golang-jwt/jwt/v4"
)

var jwtKey []byte // This will be set from an environment variable

// SetJWTSecret initializes the JWT secret key.
func SetJWTSecret(secret []byte) {
    jwtKey = secret
}

// Claims defines the JWT claims structure.
type Claims struct {
    Email string `json:"email"`
    jwt.RegisteredClaims
}

// GenerateToken creates a new JWT for the given email.
func GenerateToken(email string) (string, error) {
    // Set token expiration to 72 hours from now
    expirationTime := time.Now().Add(72 * time.Hour)
    claims := &Claims{
        Email: email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "rysto-auth-service", // Can also be an env var
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtKey)
}

// ValidateToken parses and validates a JWT string.
func ValidateToken(tokenStr string) (*Claims, error) {
    claims := &Claims{}

    token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
        // Ensure the token's signing method is HMAC.
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
