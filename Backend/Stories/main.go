
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
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, relying on system environment variables.")
	}
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = gin.DebugMode
	}
	gin.SetMode(ginMode)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("MONGODB_URI is not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set")
	}
	utils.SetJWTSecret([]byte(jwtSecret))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("MongoDB connection error:", err)
	}

	db := client.Database("RystoDB")
	models.InitCollections(db)

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

	r.Run(":" + port)
}