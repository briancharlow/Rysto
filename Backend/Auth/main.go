package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"authService.com/auth/controllers"
	"authService.com/auth/middleware"
	"authService.com/auth/redis"
	"authService.com/auth/utils"
)

func main() {
	// --- Load environment variables ---
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, relying on system environment variables.")
	}

	// --- Set Gin mode ---
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = gin.DebugMode
	}
	gin.SetMode(ginMode)

	// --- JWT Secret ---
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("Error: JWT_SECRET not set")
	}
	utils.SetJWTSecret([]byte(jwtSecret))

	// --- Redis ---
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatal("Error: REDIS_ADDR not set")
	}
	os.Setenv("REDIS_ADDR", redisAddr) // ensure redis sees it
	log.Println("Connecting to Redis...")
	redis.InitRedis()
	log.Println("Redis connected successfully!")

	// --- MongoDB ---
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("Error: MONGODB_URI not set")
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to create MongoDB client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
		log.Println("Disconnected from MongoDB.")
	}()

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("MongoDB ping failed: %v", err)
	}
	log.Println("MongoDB connected and ping successful!")

	// --- Inject collection ---
	userCollection := client.Database("RystoDB").Collection("users")
	controllers.SetUserCollection(userCollection)

	// --- Setup Gin routes ---
	r := gin.Default()

	// Public
	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)
	

	// Protected
	projectURL := os.Getenv("PROJECT_URL")
	if projectURL == "" {
		projectURL = "authService.com/auth"
	}

	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/profile", func(c *gin.Context) {
			email, exists := c.Get("email")
			if !exists {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Email not found in context"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message":      "Welcome to your protected profile!",
				"user_email":   email,
				"project_url":  projectURL,
				"access_level": "authenticated",
			})
		})

		protected.POST("/logout", controllers.Logout)
	}

	// --- Run server ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Auth service running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
