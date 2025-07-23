package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Vote struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ContinuationID primitive.ObjectID `bson:"continuationId" json:"continuationId"`
	VoterEmail     string             `bson:"voterEmail" json:"voterEmail"`
	VotedAt        time.Time          `bson:"votedAt" json:"votedAt"`
}

var voteCollection *mongo.Collection

func InitCollection(db *mongo.Database) {
	voteCollection = db.Collection("votes")
}

func CreateVote(ctx context.Context, vote Vote) error {
	_, err := voteCollection.InsertOne(ctx, vote)
	return err
}

func GetVotesByContinuation(ctx context.Context, continuationID primitive.ObjectID) ([]Vote, error) {
	cursor, err := voteCollection.Find(ctx, bson.M{"continuationId": continuationID})
	if err != nil {
		return nil, err
	}
	var votes []Vote
	if err = cursor.All(ctx, &votes); err != nil {
		return nil, err
	}
	return votes, nil
}

func DeleteVote(ctx context.Context, continuationID primitive.ObjectID, email string) error {
	_, err := voteCollection.DeleteOne(ctx, bson.M{
		"continuationId": continuationID,
		"voterEmail":     email,
	})
	return err
}
