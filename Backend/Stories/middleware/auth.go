package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"storyService.com/story/utils"
	"storyService.com/story/redis"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be Bearer <token>"})
			c.Abort()
			return
		}

		token := parts[1]

		// Check if token exists in Redis
		email, err := redis.Client.Get(redis.Ctx, token).Result()
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Validate token signature & expiry
		claims, err := utils.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token verification failed"})
			c.Abort()
			return
		}

		if claims.Email != email {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token mismatch"})
			c.Abort()
			return
		}

		// Save data for later use
		c.Set("email", email)
		c.Set("token", token)

		c.Next()
	}
}
