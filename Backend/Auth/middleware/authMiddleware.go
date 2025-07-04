package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"

    "authService.com/auth/utils" // Updated import
)

// AuthMiddleware validates the JWT token from the Authorization header.
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
            c.Abort() // Stop processing the request
            return
        }

        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
            c.Abort()
            return
        }

        token := parts[1]
        claims, err := utils.ValidateToken(token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
            c.Abort()
            return
        }

        // Store user email in context for subsequent handlers to access
        c.Set("email", claims.Email)
        c.Next() // Continue to the next handler
    }
}
