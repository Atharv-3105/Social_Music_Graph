package handlers

import (
	"net/http"

	"github.com/atharv-3105/Social_Music_Graph/db"
	"github.com/atharv-3105/Social_Music_Graph/logger"
	"github.com/atharv-3105/Social_Music_Graph/models"
	"github.com/gin-gonic/gin"
)


func AddListen(c *gin.Context) {
	var req models.ListenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H {
			"error" : "invalid request body",
		})
		return 
	}

	err := db.AddListen(req.UserID, req.TrackID, req.Title)
	if err != nil {
		logger.Log.WithError(err).Error("failed to add listen")
		c.JSON(http.StatusInternalServerError, gin.H {
			"error" : "internal server error",
		})
		return 
	}

	c.JSON(http.StatusCreated, gin.H {
		"message" : "listen recorded",
	})
}

func GetMusicSimilarity(c *gin.Context) {
	userID := c.Param("user_id")

	results, err := db.GetMusicSimilarity(userID)

	if err != nil {
		logger.Log.WithError(err).Error("failed to compute music similarity")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error" : "internal server error",
		})
		return 
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id" : userID,
		"similarity" : results,
	})
}

func GetCosineSimilarity(c *gin.Context) {
	userID := c.Param("user_id")

	results, err := db.GetCosineSimilarity(userID)
	if err != nil {
		logger.Log.WithError(err).Error("failed to compute cosine similarity")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error" : "internal server error",
		})
		return 
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id" : userID,
		"similarity" : results, 
	})
}