package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Story struct {
	ID        primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	AuthorID  string              `bson:"authorId" json:"authorId"`
	Content   string              `bson:"content" json:"content"`
	CreatedAt time.Time           `bson:"createdAt" json:"createdAt"`
	Accepted  *primitive.ObjectID `bson:"accepted,omitempty" json:"accepted,omitempty"`
}

type Continuation struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	StoryID   primitive.ObjectID `bson:"storyId" json:"storyId"`
	AuthorID  string             `bson:"authorId" json:"authorId"`
	Content   string             `bson:"content" json:"content"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	Accepted  bool               `bson:"accepted" json:"accepted"`
}

var StoryCollection *mongo.Collection
var ContinuationCollection *mongo.Collection

func InitCollections(db *mongo.Database) {
	StoryCollection = db.Collection("stories")
	ContinuationCollection = db.Collection("continuations")
}

func DeleteContinuationsByStoryID(ctx context.Context, storyID primitive.ObjectID) error {
	_, err := ContinuationCollection.DeleteMany(ctx, bson.M{"storyId": storyID})
	return err
}
