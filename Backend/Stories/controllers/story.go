package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"storyService.com/story/models"
	"storyService.com/story/metrics"
)

func getUserEmail(c *gin.Context) string {
	email, _ := c.Get("email")
	return email.(string)
}

// CreateStory
func CreateStory(c *gin.Context) {
	start := time.Now()
	var req struct {
		Content string   `json:"content" binding:"required"`
		Title   string   `json:"title" binding:"required"`
		Tags    []string `json:"tags,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		metrics.HttpRequests.WithLabelValues("/stories", "400").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	story := models.Story{
		AuthorID:  getUserEmail(c),
		Content:   req.Content,
		Title:     req.Title,
		Tags:      req.Tags,
		CreatedAt: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := models.StoryCollection.InsertOne(ctx, story)
	if err != nil {
		metrics.HttpRequests.WithLabelValues("/stories", "500").Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create story"})
		return
	}

	story.ID = res.InsertedID.(primitive.ObjectID)

	metrics.HttpRequests.WithLabelValues("/stories", "201").Inc()
	metrics.StoriesCreated.Inc()
	metrics.HttpRequestDuration.WithLabelValues("/stories").Observe(time.Since(start).Seconds())

	c.JSON(http.StatusCreated, story)
}

// AddContinuation
func AddContinuation(c *gin.Context) {
	start := time.Now()
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		metrics.HttpRequests.WithLabelValues("/continuations", "400").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	storyID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		metrics.HttpRequests.WithLabelValues("/continuations", "400").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid story ID"})
		return
	}

	cont := models.Continuation{
		StoryID:   storyID,
		AuthorID:  getUserEmail(c),
		Content:   req.Content,
		CreatedAt: time.Now(),
		Accepted:  false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := models.ContinuationCollection.InsertOne(ctx, cont)
	if err != nil {
		metrics.HttpRequests.WithLabelValues("/continuations", "500").Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit continuation"})
		return
	}

	cont.ID = res.InsertedID.(primitive.ObjectID)

	metrics.HttpRequests.WithLabelValues("/continuations", "201").Inc()
	metrics.ContinuationsSubmitted.Inc()
	metrics.HttpRequestDuration.WithLabelValues("/continuations").Observe(time.Since(start).Seconds())

	c.JSON(http.StatusCreated, cont)
}

// EditStory
func EditStory(c *gin.Context) {
	start := time.Now()
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		metrics.HttpRequests.WithLabelValues("/stories/edit", "400").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	authorID := getUserEmail(c)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": id, "authorId": authorID}
	update := bson.M{"$set": bson.M{"content": req.Content}}

	res := models.StoryCollection.FindOneAndUpdate(ctx, filter, update)
	if res.Err() != nil {
		metrics.HttpRequests.WithLabelValues("/stories/edit", "403").Inc()
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized or story not found"})
		return
	}

	metrics.HttpRequests.WithLabelValues("/stories/edit", "200").Inc()
	metrics.HttpRequestDuration.WithLabelValues("/stories/edit").Observe(time.Since(start).Seconds())
	c.JSON(http.StatusOK, gin.H{"message": "Story updated"})
}

// EditContinuation
func EditContinuation(c *gin.Context) {
	start := time.Now()
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		metrics.HttpRequests.WithLabelValues("/continuations/edit", "400").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cid, _ := primitive.ObjectIDFromHex(c.Param("cid"))
	authorID := getUserEmail(c)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": cid, "authorId": authorID, "accepted": false}
	update := bson.M{"$set": bson.M{"content": req.Content}}

	res := models.ContinuationCollection.FindOneAndUpdate(ctx, filter, update)
	if res.Err() != nil {
		metrics.HttpRequests.WithLabelValues("/continuations/edit", "403").Inc()
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized or continuation locked"})
		return
	}

	metrics.HttpRequests.WithLabelValues("/continuations/edit", "200").Inc()
	metrics.HttpRequestDuration.WithLabelValues("/continuations/edit").Observe(time.Since(start).Seconds())
	c.JSON(http.StatusOK, gin.H{"message": "Continuation updated"})
}

// DeleteStory
func DeleteStory(c *gin.Context) {
	start := time.Now()
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	authorID := getUserEmail(c)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := models.StoryCollection.DeleteOne(ctx, bson.M{"_id": id, "authorId": authorID})
	if err != nil || res.DeletedCount == 0 {
		metrics.HttpRequests.WithLabelValues("/stories/delete", "403").Inc()
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized or story not found"})
		return
	}

	_ = models.DeleteContinuationsByStoryID(ctx, id)

	metrics.HttpRequests.WithLabelValues("/stories/delete", "200").Inc()
	metrics.HttpRequestDuration.WithLabelValues("/stories/delete").Observe(time.Since(start).Seconds())
	c.JSON(http.StatusOK, gin.H{"message": "Story and its continuations deleted"})
}

// DeleteContinuation
func DeleteContinuation(c *gin.Context) {
	start := time.Now()
	cid, _ := primitive.ObjectIDFromHex(c.Param("cid"))
	authorID := getUserEmail(c)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": cid, "authorId": authorID, "accepted": false}
	res, err := models.ContinuationCollection.DeleteOne(ctx, filter)
	if err != nil || res.DeletedCount == 0 {
		metrics.HttpRequests.WithLabelValues("/continuations/delete", "403").Inc()
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized or continuation locked"})
		return
	}

	metrics.HttpRequests.WithLabelValues("/continuations/delete", "200").Inc()
	metrics.HttpRequestDuration.WithLabelValues("/continuations/delete").Observe(time.Since(start).Seconds())
	c.JSON(http.StatusOK, gin.H{"message": "Continuation deleted"})
}

// AcceptContinuation
func AcceptContinuation(c *gin.Context) {
	start := time.Now()
	storyID, _ := primitive.ObjectIDFromHex(c.Param("id"))
	cid, _ := primitive.ObjectIDFromHex(c.Param("cid"))
	authorID := getUserEmail(c)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var story models.Story
	err := models.StoryCollection.FindOne(ctx, bson.M{"_id": storyID, "authorId": authorID}).Decode(&story)
	if err != nil {
		metrics.HttpRequests.WithLabelValues("/continuations/accept", "403").Inc()
		c.JSON(http.StatusForbidden, gin.H{"error": "You cannot accept for this story"})
		return
	}

	_, err = models.StoryCollection.UpdateOne(ctx, bson.M{"_id": storyID}, bson.M{"$set": bson.M{"accepted": cid}})
	if err != nil {
		metrics.HttpRequests.WithLabelValues("/continuations/accept", "500").Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept continuation"})
		return
	}

	_, _ = models.ContinuationCollection.UpdateOne(ctx, bson.M{"_id": cid}, bson.M{"$set": bson.M{"accepted": true}})

	metrics.HttpRequests.WithLabelValues("/continuations/accept", "200").Inc()
	metrics.ContinuationsAccepted.Inc()
	metrics.HttpRequestDuration.WithLabelValues("/continuations/accept").Observe(time.Since(start).Seconds())
	c.JSON(http.StatusOK, gin.H{"message": "Continuation accepted"})
}

// GetAllStoriesWithContinuations
func GetAllStoriesWithContinuations(c *gin.Context) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var stories []models.Story
	storiesCollections, err := models.StoryCollection.Find(ctx, bson.M{})
	if err != nil {
		metrics.HttpRequests.WithLabelValues("/stories/all", "500").Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stories"})
		return
	}
	defer storiesCollections.Close(ctx)
	if err := storiesCollections.All(ctx, &stories); err != nil {
		metrics.HttpRequests.WithLabelValues("/stories/all", "500").Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode stories"})
		return
	}

	var storiesWithContinuations []gin.H
	for _, story := range stories {
		var continuations []models.Continuation
		continuationsCollections, err := models.ContinuationCollection.Find(ctx, bson.M{"storyId": story.ID})
		if err != nil {
			metrics.HttpRequests.WithLabelValues("/stories/all", "500").Inc()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch continuations"})
			return
		}
		if err := continuationsCollections.All(ctx, &continuations); err != nil {
			metrics.HttpRequests.WithLabelValues("/stories/all", "500").Inc()
			continuationsCollections.Close(ctx)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode continuations"})
			return
		}
		continuationsCollections.Close(ctx)
		storiesWithContinuations = append(storiesWithContinuations, gin.H{
			"story":         story,
			"continuations": continuations,
		})
	}

	metrics.HttpRequests.WithLabelValues("/stories/all", "200").Inc()
	metrics.HttpRequestDuration.WithLabelValues("/stories/all").Observe(time.Since(start).Seconds())
	c.JSON(http.StatusOK, storiesWithContinuations)
}

// GetStoryByID
func GetStoryByID(c *gin.Context) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	storyIDHex := c.Param("id")
	storyID, err := primitive.ObjectIDFromHex(storyIDHex)
	if err != nil {
		metrics.HttpRequests.WithLabelValues("/stories/id", "400").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid story ID"})
		return
	}

	var story models.Story
	if err := models.StoryCollection.FindOne(ctx, bson.M{"_id": storyID}).Decode(&story); err != nil {
		metrics.HttpRequests.WithLabelValues("/stories/id", "404").Inc()
		c.JSON(http.StatusNotFound, gin.H{"error": "Story not found"})
		return
	}

	var continuations []models.Continuation
	continuationsCollections, err := models.ContinuationCollection.Find(ctx, bson.M{"storyId": story.ID})
	if err != nil {
		metrics.HttpRequests.WithLabelValues("/stories/id", "500").Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch continuations"})
		return
	}
	if err := continuationsCollections.All(ctx, &continuations); err != nil {
		metrics.HttpRequests.WithLabelValues("/stories/id", "500").Inc()
		continuationsCollections.Close(ctx)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode continuations"})
		return
	}
	continuationsCollections.Close(ctx)

	metrics.HttpRequests.WithLabelValues("/stories/id", "200").Inc()
	metrics.HttpRequestDuration.WithLabelValues("/stories/id").Observe(time.Since(start).Seconds())
	c.JSON(http.StatusOK, gin.H{
		"story":         story,
		"continuations": continuations,
	})
}

// GetStoriesByTitle
func GetStoriesByTitle(c *gin.Context) {
	start := time.Now()
	title := c.Query("title")
	if title == "" {
		metrics.HttpRequests.WithLabelValues("/stories/title", "400").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title query parameter is required"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var stories []models.Story
	storiesCollections, err := models.StoryCollection.Find(ctx, bson.M{"title": title})
	if err != nil {
		metrics.HttpRequests.WithLabelValues("/stories/title", "500").Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stories"})
		return
	}
	defer storiesCollections.Close(ctx)
	if err := storiesCollections.All(ctx, &stories); err != nil {
		metrics.HttpRequests.WithLabelValues("/stories/title", "500").Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode stories"})
		return
	}

	var storiesWithContinuations []gin.H
	for _, story := range stories {
		var continuations []models.Continuation
		continuationsCollections, err := models.ContinuationCollection.Find(ctx, bson.M{"storyId": story.ID})
		if err != nil {
			metrics.HttpRequests.WithLabelValues("/stories/title", "500").Inc()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch continuations"})
			return
		}
		if err := continuationsCollections.All(ctx, &continuations); err != nil {
			metrics.HttpRequests.WithLabelValues("/stories/title", "500").Inc()
			continuationsCollections.Close(ctx)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode continuations"})
			return
		}
		continuationsCollections.Close(ctx)
		storiesWithContinuations = append(storiesWithContinuations, gin.H{
			"story":         story,
			"continuations": continuations,
		})
	}

	metrics.HttpRequests.WithLabelValues("/stories/title", "200").Inc()
	metrics.HttpRequestDuration.WithLabelValues("/stories/title").Observe(time.Since(start).Seconds())
	c.JSON(http.StatusOK, storiesWithContinuations)
}

// GetStoriesByAuthor
func GetStoriesByAuthor(c *gin.Context) {
	start := time.Now()
	authorID := c.Query("authorId")
	if authorID == "" {
		metrics.HttpRequests.WithLabelValues("/stories/author", "400").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": "authorId query parameter is required"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var stories []models.Story
	storiesCollections, err := models.StoryCollection.Find(ctx, bson.M{"authorId": authorID})
	if err != nil {
		metrics.HttpRequests.WithLabelValues("/stories/author", "500").Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stories"})
		return
	}
	defer storiesCollections.Close(ctx)
	if err := storiesCollections.All(ctx, &stories); err != nil {
		metrics.HttpRequests.WithLabelValues("/stories/author", "500").Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode stories"})
		return
	}

	var storiesWithContinuations []gin.H
	for _, story := range stories {
		var continuations []models.Continuation
		continuationsCollections, err := models.ContinuationCollection.Find(ctx, bson.M{"storyId": story.ID})
		if err != nil {
			metrics.HttpRequests.WithLabelValues("/stories/author", "500").Inc()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch continuations"})
			return
		}
		if err := continuationsCollections.All(ctx, &continuations); err != nil {
			metrics.HttpRequests.WithLabelValues("/stories/author", "500").Inc()
			continuationsCollections.Close(ctx)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode continuations"})
			return
		}
		continuationsCollections.Close(ctx)
		storiesWithContinuations = append(storiesWithContinuations, gin.H{
			"story":         story,
			"continuations": continuations,
		})
	}

	metrics.HttpRequests.WithLabelValues("/stories/author", "200").Inc()
	metrics.HttpRequestDuration.WithLabelValues("/stories/author").Observe(time.Since(start).Seconds())
	c.JSON(http.StatusOK, storiesWithContinuations)
}
