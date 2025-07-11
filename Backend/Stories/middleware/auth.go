package middleware

import (
	"net/http"
	

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

		token := utils.ExtractToken(authHeader)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be Bearer <token>"})
			c.Abort()
			return
		}

		// log.Printf("AuthMiddleware: token extracted: %s", token)

		claims, err := utils.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token verification failed"})
			c.Abort()
			return
		}

		email, err := redis.Client.Get(redis.Ctx, token).Result()
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		if claims.Email != email {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token mismatch"})
			c.Abort()
			return
		}

		c.Set("email", email)
		c.Set("token", token)

		c.Next()
	}
}
