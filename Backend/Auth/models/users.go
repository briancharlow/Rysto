package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// User represents the structure of a user document in MongoDB.
type User struct {
    ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Email    string             `bson:"email" json:"email" binding:"required,email"`
    Password string             `bson:"password" json:"-" binding:"required,min=6"` // `json:"-"` hides it from JSON output
}
