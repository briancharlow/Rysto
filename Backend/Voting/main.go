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

	"votingService.com/voting/controllers"
	"votingService.com/voting/middleware"
	"votingService.com/voting/models"
	"votingService.com/voting/utils"
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI not set")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET not set")
	}
	utils.SetJWTSecret([]byte(jwtSecret))

	log.Println("Connecting to MongoDB...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("MongoDB connection failed: %v", err)
	}
	db := client.Database("RystoDB")
	models.InitCollection(db)

	r := gin.Default()
	api := r.Group("/api/votes")
	api.Use(middleware.AuthMiddleware())
	{
		api.POST("/:continuationId", controllers.CreateVote)
		api.GET("/:continuationId", controllers.GetVotesByContinuation)
		api.DELETE("/:continuationId", controllers.DeleteVote)
	}

	log.Printf("Voting service running on port %s", port)
	r.Run(":" + port)
}
