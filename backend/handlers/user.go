package handlers

import (
	"net/http"

	"github.com/atharv-3105/Social_Music_Graph/db"
	"github.com/atharv-3105/Social_Music_Graph/logger"
	"github.com/atharv-3105/Social_Music_Graph/models"
	"github.com/gin-gonic/gin"
	
)



func CreateUer(c *gin.Context){
	var req models.CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.WithError(err).Warn("invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return 
	}


	err := db.CreateUser(req.UserID, req.Name)

	if err != nil {

		if err == db.ErrUserAlreadyExists{
			c.JSON(http.StatusConflict, gin.H{
				"error": "user already exists",
			})
			return 
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return 
	}

	c.JSON(http.StatusCreated, models.UserResponse{
		UserID: req.UserID,
		Name:	req.Name,
	})
}


func GetUser(c *gin.Context){
	userID := c.Param("user_id")

	if userID == ""{
		c.JSON(http.StatusBadRequest, gin.H{
			"error":"user_id is required",
		})
		return 
	}

	id, name, err := db.GetUser(userID)

	if err != nil {
		if err == db.ErrUserNotFound{
			c.JSON(http.StatusNotFound, gin.H{
				"error":"user not found",
			})
			return 
		}

		logger.Log.WithError(err).Error("failed to fetch user")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return 
	}

	c.JSON(http.StatusOK, models.UserResponse{
		UserID: id,
		Name: name,
	})
}


func FollowUser(c *gin.Context) {
	var req models.FollowRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return 
	}

	err := db.FollowUser(req.FollowerID, req.FolloweeID)

	if err != nil {

		switch err {
			case db.ErrUserNotFound:
				c.JSON(http.StatusNotFound, gin.H{
					"error": "one or both users not found",
				})
				return 
		
			case db.ErrAlreadyFollowing:
				c.JSON(http.StatusConflict, gin.H {
					"error": "already following",
				})
				return 
			case db.ErrCannotFollowSelf:
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "cannot follow yourself",
				})
				return
		}

		logger.Log.WithError(err).Error("follow operation failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return 
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "follow successful",
	})
}

func GetFollowers(c *gin.Context) {
	userID := c.Param("user_id")

	followers, err := db.GetFollowers(userID)

	if err != nil {

		if err == db.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "user not found",
			})
			return 
		}

		logger.Log.WithError(err).Error("failed to fetch followers")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return 
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id" : userID,
		"followers": followers,
	})
}


func GetFollowing(c *gin.Context) {
	userID := c.Param("user_id")

	following, err := db.GetFollowing(userID)
	if err != nil {

		if err == db.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H {
				"error" : "user not found",
			})
			return 
		}

		logger.Log.WithError(err).Error("failed to fetch following")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return 
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id" : userID,
		"following": following,
	})
}


func SuggestUsers(c *gin.Context) {
	userID := c.Param("user_id")

	suggestions, err := db.SuggestUsers(userID)
	if err != nil {
		logger.Log.WithError(err).Error("failed to fetch suggestions")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error" : "internal server error",
		})
		return 
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id" : userID, 
		"suggestions" : suggestions,
	})
}
