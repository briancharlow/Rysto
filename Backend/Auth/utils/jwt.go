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


type Claims struct {
    Email string `json:"email"`
    jwt.RegisteredClaims
}


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
