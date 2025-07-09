package middleware

import (
	"net/http"

	"authService.com/auth/redis"
	"authService.com/auth/utils"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the JWT and checks Redis for token validity
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		token := utils.ExtractToken(authHeader)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization format must be Bearer {token}"})
			c.Abort()
			return
		}

		// Validate JWT token
		claims, err := utils.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Check Redis to ensure token is active
		exists, err := redis.Client.Exists(redis.Ctx, token).Result()
		if err != nil || exists == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is no longer valid"})
			c.Abort()
			return
		}

		// Token is valid and present in Redis
		c.Set("email", claims.Email)
		c.Next()
	}
}
