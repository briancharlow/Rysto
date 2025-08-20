package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"votingService.com/voting/models"
)

func CreateVote(c *gin.Context) {
	email, _ := c.Get("email")
	continuationId := c.Param("continuationId")
	objID, err := primitive.ObjectIDFromHex(continuationId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid continuation ID"})
		return
	}

	vote := models.Vote{
		ContinuationID: objID,
		VoterEmail:     email.(string),
		VotedAt:        time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := models.CreateVote(ctx, vote); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record vote"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Vote recorded"})
}

func GetVotesByContinuation(c *gin.Context) {
	continuationId := c.Param("continuationId")
	objID, err := primitive.ObjectIDFromHex(continuationId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid continuation ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	votes, err := models.GetVotesByContinuation(ctx, objID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch votes"})
		return
	}
	c.JSON(http.StatusOK, votes)
}

func DeleteVote(c *gin.Context) {
	email, _ := c.Get("email")
	continuationId := c.Param("continuationId")
	objID, err := primitive.ObjectIDFromHex(continuationId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid continuation ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := models.DeleteVote(ctx, objID, email.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete vote"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Vote deleted"})
}


func RemoveVote(c *gin.Context){
	
}