package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"storyService.com/story/models"
)

func getUserEmail(c *gin.Context) string {
	email, _ := c.Get("email")
	return email.(string)
}

// create story
func CreateStory(c *gin.Context) {
	var req struct {
		Content string `json:"content" binding:"required"`
		Title   string `json:"title" binding:"required"`
		Tags    []string `json:"tags,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create story"})
		return
	}

	story.ID = res.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, story)
}

// AddContinuation
func AddContinuation(c *gin.Context) {
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	storyID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit continuation"})
		return
	}

	cont.ID = res.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, cont)
}

// EditStory allows the author to edit their story
func EditStory(c *gin.Context) {
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
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
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized or story not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Story updated"})
}

// EditContinuation allows the original submitter to edit if not accepted
func EditContinuation(c *gin.Context) {
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
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
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized or continuation locked"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Continuation updated"})
}

// DeleteStory deletes a story and all its continuations
func DeleteStory(c *gin.Context) {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	authorID := getUserEmail(c)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := models.StoryCollection.DeleteOne(ctx, bson.M{"_id": id, "authorId": authorID})
	if err != nil || res.DeletedCount == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized or story not found"})
		return
	}

	_ = models.DeleteContinuationsByStoryID(ctx, id)
	c.JSON(http.StatusOK, gin.H{"message": "Story and its continuations deleted"})
}

// DeleteContinuation deletes a continuation if not accepted
func DeleteContinuation(c *gin.Context) {
	cid, _ := primitive.ObjectIDFromHex(c.Param("cid"))
	authorID := getUserEmail(c)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": cid, "authorId": authorID, "accepted": false}
	res, err := models.ContinuationCollection.DeleteOne(ctx, filter)
	if err != nil || res.DeletedCount == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized or continuation locked"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Continuation deleted"})
}

// AcceptContinuation marks a continuation as accepted in a story
func AcceptContinuation(c *gin.Context) {
	storyID, _ := primitive.ObjectIDFromHex(c.Param("id"))
	cid, _ := primitive.ObjectIDFromHex(c.Param("cid"))
	authorID := getUserEmail(c)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var story models.Story
	err := models.StoryCollection.FindOne(ctx, bson.M{"_id": storyID, "authorId": authorID}).Decode(&story)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You cannot accept for this story"})
		return
	}

	_, err = models.StoryCollection.UpdateOne(ctx, bson.M{"_id": storyID}, bson.M{"$set": bson.M{"accepted": cid}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept continuation"})
		return
	}

	_, _ = models.ContinuationCollection.UpdateOne(ctx, bson.M{"_id": cid}, bson.M{"$set": bson.M{"accepted": true}})
	c.JSON(http.StatusOK, gin.H{"message": "Continuation accepted"})
}

// GetAllStoriesWithContinuations returns all stories and their associated continuations
func GetAllStoriesWithContinuations(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var stories []models.Story
	storiesCollections, err := models.StoryCollection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stories"})
		return
	}
	defer storiesCollections.Close(ctx)
	if err := storiesCollections.All(ctx, &stories); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode stories"})
		return
	}

	var storiesWithContinuations []gin.H
	for _, story := range stories {
		var continuations []models.Continuation
		continuationsCollections, err := models.ContinuationCollection.Find(ctx, bson.M{"storyId": story.ID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch continuations"})
			return
		}
		if err := continuationsCollections.All(ctx, &continuations); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode continuations"})
			continuationsCollections.Close(ctx)
			return
		}
		continuationsCollections.Close(ctx)
		storiesWithContinuations = append(storiesWithContinuations, gin.H{
			"story":         story,
			"continuations": continuations,
		})
	}
	c.JSON(http.StatusOK, storiesWithContinuations)
}

// GetStoryByID returns a story and its continuations by story ID
func GetStoryByID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	storyIDHex := c.Param("id")
	storyID, err := primitive.ObjectIDFromHex(storyIDHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid story ID"})
		return
	}

	var story models.Story
	if err := models.StoryCollection.FindOne(ctx, bson.M{"_id": storyID}).Decode(&story); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Story not found"})
		return
	}

	var continuations []models.Continuation
	continuationsCollections, err := models.ContinuationCollection.Find(ctx, bson.M{"storyId": story.ID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch continuations"})
		return
	}
	if err := continuationsCollections.All(ctx, &continuations); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode continuations"})
		continuationsCollections.Close(ctx)
		return
	}
	continuationsCollections.Close(ctx)

	c.JSON(http.StatusOK, gin.H{
		"story":         story,
		"continuations": continuations,
	})
}

// GetStoriesByTitle returns stories and their continuations by title
func GetStoriesByTitle(c *gin.Context) {
	title := c.Query("title")
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title query parameter is required"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var stories []models.Story
	storiesCollections, err := models.StoryCollection.Find(ctx, bson.M{"title": title})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stories"})
		return
	}
	defer storiesCollections.Close(ctx)
	if err := storiesCollections.All(ctx, &stories); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode stories"})
		return
	}

	var storiesWithContinuations []gin.H
	for _, story := range stories {
		var continuations []models.Continuation
		continuationsCollections, err := models.ContinuationCollection.Find(ctx, bson.M{"storyId": story.ID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch continuations"})
			return
		}
		if err := continuationsCollections.All(ctx, &continuations); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode continuations"})
			continuationsCollections.Close(ctx)
			return
		}
		continuationsCollections.Close(ctx)
		storiesWithContinuations = append(storiesWithContinuations, gin.H{
			"story":         story,
			"continuations": continuations,
		})
	}
	c.JSON(http.StatusOK, storiesWithContinuations)
}

// GetStoriesByAuthor returns stories and their continuations by authorId
func GetStoriesByAuthor(c *gin.Context) {
	authorID := c.Query("authorId")
	if authorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "authorId query parameter is required"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var stories []models.Story
	storiesCollections, err := models.StoryCollection.Find(ctx, bson.M{"authorId": authorID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stories"})
		return
	}
	defer storiesCollections.Close(ctx)
	if err := storiesCollections.All(ctx, &stories); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode stories"})
		return
	}

	var storiesWithContinuations []gin.H
	for _, story := range stories {
		var continuations []models.Continuation
		continuationsCollections, err := models.ContinuationCollection.Find(ctx, bson.M{"storyId": story.ID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch continuations"})
			return
		}
		if err := continuationsCollections.All(ctx, &continuations); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode continuations"})
			continuationsCollections.Close(ctx)
			return
		}
		continuationsCollections.Close(ctx)
		storiesWithContinuations = append(storiesWithContinuations, gin.H{
			"story":         story,
			"continuations": continuations,
		})
	}
	c.JSON(http.StatusOK, storiesWithContinuations)
}
