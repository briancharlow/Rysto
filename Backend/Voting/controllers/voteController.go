package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"votingService.com/voting/models"
	"votingService.com/voting/metrics"
)

func CreateVote(c *gin.Context) {
	start := time.Now()
	email, _ := c.Get("email")
	continuationId := c.Param("continuationId")
	objID, err := primitive.ObjectIDFromHex(continuationId)
	if err != nil {
		metrics.HttpRequests.WithLabelValues("/api/votes/:continuationId", "400").Inc()
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
		metrics.HttpRequests.WithLabelValues("/api/votes/:continuationId", "500").Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record vote"})
		return
	}

	metrics.HttpRequests.WithLabelValues("/api/votes/:continuationId", "201").Inc()
	metrics.VotesCast.Inc()
	metrics.ActiveVotes.Inc()
	metrics.HttpRequestDuration.WithLabelValues("/api/votes/:continuationId").Observe(time.Since(start).Seconds())

	c.JSON(http.StatusCreated, gin.H{"message": "Vote recorded"})
}

func GetVotesByContinuation(c *gin.Context) {
	start := time.Now()
	continuationId := c.Param("continuationId")
	objID, err := primitive.ObjectIDFromHex(continuationId)
	if err != nil {
		metrics.HttpRequests.WithLabelValues("/api/votes/:continuationId", "400").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid continuation ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	votes, err := models.GetVotesByContinuation(ctx, objID)
	if err != nil {
		metrics.HttpRequests.WithLabelValues("/api/votes/:continuationId", "500").Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch votes"})
		return
	}

	metrics.HttpRequests.WithLabelValues("/api/votes/:continuationId", "200").Inc()
	metrics.HttpRequestDuration.WithLabelValues("/api/votes/:continuationId").Observe(time.Since(start).Seconds())
	c.JSON(http.StatusOK, votes)
}

func DeleteVote(c *gin.Context) {
	start := time.Now()
	email, _ := c.Get("email")
	continuationId := c.Param("continuationId")
	objID, err := primitive.ObjectIDFromHex(continuationId)
	if err != nil {
		metrics.HttpRequests.WithLabelValues("/api/votes/:continuationId", "400").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid continuation ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := models.DeleteVote(ctx, objID, email.(string)); err != nil {
		metrics.HttpRequests.WithLabelValues("/api/votes/:continuationId", "500").Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete vote"})
		return
	}

	metrics.HttpRequests.WithLabelValues("/api/votes/:continuationId", "200").Inc()
	metrics.VotesDeleted.Inc()
	metrics.ActiveVotes.Dec()
	metrics.HttpRequestDuration.WithLabelValues("/api/votes/:continuationId").Observe(time.Since(start).Seconds())

	c.JSON(http.StatusOK, gin.H{"message": "Vote deleted"})
}
