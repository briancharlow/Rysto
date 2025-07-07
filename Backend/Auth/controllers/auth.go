package controllers

import (
    "context"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "golang.org/x/crypto/bcrypt"

    "authService.com/auth/models" 
    "authService.com/auth/utils" 
)


var userCollection *mongo.Collection

// SetUserCollection injects the MongoDB collection for user operations.
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


func Register(c *gin.Context) {
    var input RegisterInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    
    count, err := userCollection.CountDocuments(ctx, bson.M{"email": input.Email})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking email availability"})
        return
    }
    if count > 0 {
        c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
        return
    }

   
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
        
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error during login"})
        }
        return
    }

    
    err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        return
    }
 
    
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
