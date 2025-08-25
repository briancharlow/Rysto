package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"authService.com/auth/models"
	"authService.com/auth/redis"
	"authService.com/auth/utils"
	"authService.com/auth/metrics" // <-- new import for Prometheus metrics
)

var userCollection *mongo.Collection

func SetUserCollection(collection *mongo.Collection) {
	userCollection = collection
}

type RegisterInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

// Register handles user registration
func Register(c *gin.Context) {
	start := time.Now()

	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		metrics.HttpRequests.WithLabelValues("/register", "400").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := userCollection.CountDocuments(ctx, bson.M{"email": input.Email})
	if err != nil {
		metrics.HttpRequests.WithLabelValues("/register", "500").Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking email availability"})
		return
	}
	if count > 0 {
		metrics.HttpRequests.WithLabelValues("/register", "409").Inc()
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		metrics.HttpRequests.WithLabelValues("/register", "500").Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Email:    input.Email,
		Password: string(hashedPassword),
	}

	_, err = userCollection.InsertOne(ctx, user)
	if err != nil {
		metrics.HttpRequests.WithLabelValues("/register", "500").Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	metrics.HttpRequests.WithLabelValues("/register", "201").Inc()
	metrics.HttpRequestDuration.WithLabelValues("/register").Observe(time.Since(start).Seconds())

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// Login authenticates a user and stores the token in Redis
func Login(c *gin.Context) {
	start := time.Now()

	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		metrics.FailedLogins.Inc()
		metrics.HttpRequests.WithLabelValues("/login", "400").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			metrics.FailedLogins.Inc()
			metrics.HttpRequests.WithLabelValues("/login", "401").Inc()
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		} else {
			metrics.HttpRequests.WithLabelValues("/login", "500").Inc()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error during login"})
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		metrics.FailedLogins.Inc()
		metrics.HttpRequests.WithLabelValues("/login", "401").Inc()
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	token, err := utils.GenerateToken(user.Email)
	if err != nil {
		metrics.HttpRequests.WithLabelValues("/login", "500").Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Store token in Redis with expiration
	tokenTTL := 24 * time.Hour
	err = redis.Client.Set(redis.Ctx, token, user.Email, tokenTTL).Err()
	if err != nil {
		metrics.HttpRequests.WithLabelValues("/login", "500").Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store token"})
		return
	}

	metrics.SuccessfulLogins.Inc()
	metrics.ActiveSessions.Inc()
	metrics.HttpRequests.WithLabelValues("/login", "200").Inc()
	metrics.HttpRequestDuration.WithLabelValues("/login").Observe(time.Since(start).Seconds())

	c.JSON(http.StatusOK, LoginResponse{
		Token:   token,
		Message: "Login successful",
	})
}

func Logout(c *gin.Context) {
	start := time.Now()

	token, exists := c.Get("token")
	if !exists {
		log.Println("Logout: token not found in context")
		metrics.HttpRequests.WithLabelValues("/logout", "401").Inc()
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token not found in context"})
		return
	}

	log.Printf("Logout: token to delete: %s", token.(string))

	err := redis.Client.Del(redis.Ctx, token.(string)).Err()
	if err != nil {
		metrics.HttpRequests.WithLabelValues("/logout", "500").Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to invalidate token"})
		return
	}

	metrics.ActiveSessions.Dec()
	metrics.HttpRequests.WithLabelValues("/logout", "200").Inc()
	metrics.HttpRequestDuration.WithLabelValues("/logout").Observe(time.Since(start).Seconds())

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
