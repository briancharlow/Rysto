package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"authService.com/auth/controllers"
	"authService.com/auth/middleware"
	"authService.com/auth/redis"
	"authService.com/auth/utils"
)

// --- Business Metrics ---
var (
	registrationCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_registrations_total",
			Help: "Total number of user registrations",
		},
	)

	loginCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_logins_total",
			Help: "Total number of successful logins",
		},
	)

	failedLoginCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_failed_logins_total",
			Help: "Total number of failed login attempts",
		},
	)

	activeSessionsGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "auth_active_sessions",
			Help: "Current number of active sessions (tokens in Redis)",
		},
	)
)

func init() {
	// Register metrics with Prometheus
	prometheus.MustRegister(registrationCounter, loginCounter, failedLoginCounter, activeSessionsGauge)
}

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
	os.Setenv("REDIS_ADDR", redisAddr)
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

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Public
	r.POST("/register", func(c *gin.Context) {
		controllers.Register(c)
		if c.Writer.Status() == http.StatusCreated {
			registrationCounter.Inc()
		}
	})

	r.POST("/login", func(c *gin.Context) {
		controllers.Login(c)
		if c.Writer.Status() == http.StatusOK {
			loginCounter.Inc()
			activeSessionsGauge.Inc()
		} else if c.Writer.Status() == http.StatusUnauthorized {
			failedLoginCounter.Inc()
		}
	})

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

		protected.POST("/logout", func(c *gin.Context) {
			controllers.Logout(c)
			if c.Writer.Status() == http.StatusOK {
				activeSessionsGauge.Dec()
			}
		})
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
