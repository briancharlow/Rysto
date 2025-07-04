package controllers

import (
    "context"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "golang.org/x/crypto/bcrypt"

    "authService.com/auth/models" // Updated import
    "authService.com/auth/utils" // Updated import
)

// userCollection is a package-level variable to hold the MongoDB collection.
var userCollection *mongo.Collection

// SetUserCollection injects the MongoDB collection for user operations.
func SetUserCollection(collection *mongo.Collection) {
    userCollection = collection
}

// RegisterInput defines the expected input for user registration.
type RegisterInput struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

// LoginInput defines the expected input for user login.
type LoginInput struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

// Register handles new user registration.
func Register(c *gin.Context) {
    var input RegisterInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Check if user with this email already exists
    count, err := userCollection.CountDocuments(ctx, bson.M{"email": input.Email})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking email availability"})
        return
    }
    if count > 0 {
        c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
        return
    }

    // Hash the password before storing
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }

    user := models.User{
        Email:    input.Email,
        Password: string(hashedPassword),
    }

    _, err = userCollection.InsertOne(ctx, user)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// Login handles user authentication and JWT token generation.
func Login(c *gin.Context) {
    var input LoginInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var user models.User
    err := userCollection.FindOne(ctx, bson.M{"email": input.Email}).Decode(&user)
    if err != nil {
        // User not found or other database error
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error during login"})
        }
        return
    }

    // Compare provided password with hashed password
    err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        return
    }

    // Generate JWT token
    token, err := utils.GenerateToken(user.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token": token,
	})
}
