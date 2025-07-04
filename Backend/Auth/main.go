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
	"authService.com/auth/utils"       
)

func main() {
    // Load environment variables from .env file
    err := godotenv.Load()
    if err != nil {
        // Log a warning if .env is not found, as in production, env vars might be set directly
        log.Println("Warning: .env file not found, relying on system environment variables.")
    }

    // --- Configure Gin Mode ---
    ginMode := os.Getenv("GIN_MODE")
    if ginMode == "" {
        ginMode = gin.DebugMode // Default to debug mode if not set
    }
    gin.SetMode(ginMode)

    // --- Get MongoDB URI from environment ---
    mongoURI := os.Getenv("MONGODB_URI")
    if mongoURI == "" {
        log.Fatal("Error: MONGODB_URI environment variable not set.")
    }

    // --- Get JWT Secret from environment ---
    jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == "" {
        log.Fatal("Error: JWT_SECRET environment variable not set.")
    }
    utils.SetJWTSecret([]byte(jwtSecret)) // Pass the secret to the JWT utility

    // --- Get Server Port from environment ---
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080" // Default port
    }

    // --- Get Project URL from environment (for example usage in protected route) ---
    projectURL := os.Getenv("PROJECT_URL")
    if projectURL == "" {
        projectURL = "authService.com/auth" // Default if not set
    }

    // --- Connect to MongoDB ---
    log.Println("Attempting to connect to MongoDB...")
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

    // Ping the primary to ensure a connection has been established
    err = client.Ping(ctx, nil)
    if err != nil {
        log.Fatalf("Failed to ping MongoDB: %v", err)
    }
    log.Println("Successfully connected and pinged MongoDB!")

    // Inject the user collection into the controllers
    userCollection := client.Database("RystoDB").Collection("users")
    controllers.SetUserCollection(userCollection)

    // --- Initialize Gin Router ---
    r := gin.Default()

    // --- Define Routes ---

    // Public routes (no authentication required)
    r.POST("/register", controllers.Register)
    r.POST("/login", controllers.Login)

    // Protected routes (require JWT authentication)
    protected := r.Group("/api")
    protected.Use(middleware.AuthMiddleware()) // Apply the authentication middleware
    {
        protected.GET("/profile", func(c *gin.Context) {
            // Retrieve email from context (set by AuthMiddleware)
            email, exists := c.Get("email")
            if !exists {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "User email not found in context"})
                return
            }
            c.JSON(http.StatusOK, gin.H{
                "message":      "Welcome to your protected profile!",
                "user_email":   email,
                "project_url":  projectURL, // Example: Using the PROJECT_URL from env
                "access_level": "authenticated",
            })
        })
        // Add more protected routes here
    }

    // --- Run the server ---
    log.Printf("Server starting on port %s...", port)
    if err := r.Run(":" + port); err != nil {
        log.Fatalf("Failed to run server: %v", err)
    }
}
