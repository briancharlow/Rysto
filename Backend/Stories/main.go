package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"storyService.com/story/controllers"
	"storyService.com/story/middleware"
	"storyService.com/story/models"
	"storyService.com/story/utils"
	"storyService.com/story/redis"
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

	// --- Inject story collections ---
	db := client.Database("RystoDB")
	models.InitCollections(db)

	// --- Setup Gin routes ---
	r := gin.Default()

	auth := r.Group("/api")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.POST("/stories", controllers.CreateStory)
		auth.POST("/stories/:id/continuations", controllers.AddContinuation)
		auth.PUT("/stories/:id", controllers.EditStory)
		auth.PUT("/stories/:id/continuations/:cid", controllers.EditContinuation)
		auth.DELETE("/stories/:id", controllers.DeleteStory)
		auth.DELETE("/stories/:id/continuations/:cid", controllers.DeleteContinuation)
		auth.POST("/stories/:id/accept/:cid", controllers.AcceptContinuation)
	}

	// --- Run server ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("Story service running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start story service: %v", err)
	}
}
