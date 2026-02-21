package main

import (
	"log"
	// "os"

	"github.com/atharv-3105/Social_Music_Graph/config"
	"github.com/atharv-3105/Social_Music_Graph/db"
	"github.com/atharv-3105/Social_Music_Graph/handlers"
	"github.com/atharv-3105/Social_Music_Graph/logger"
	"github.com/atharv-3105/Social_Music_Graph/middleware"

	"github.com/gin-gonic/gin"
)

func main(){
	gin.SetMode(gin.ReleaseMode)
	logger.Init()
	logger.Log.Info("starting Social_Music_Graph backend")

	cfg := config.Load()

	if err := db.Init(cfg.Neo4jURI, cfg.Neo4jUser, cfg.Neo4jPassword); err != nil {
		log.Fatal("failed to connect to Neo4j")
	}


	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger())
	//Simple Health
	router.GET("/health", handlers.Health)

	//Route to Add users
	router.POST("/users", handlers.CreateUer)

	//Route to Fetch Users
	router.GET("/users/:user_id", handlers.GetUser)

	router.POST("/follow", handlers.FollowUser)

	router.GET("/users/:user_id/followers", handlers.GetFollowers)
	
	router.GET("/users/:user_id/following", handlers.GetFollowing)

	router.GET("/users/:user_id/suggestions", handlers.SuggestUsers)

	router.GET("/users/:user_id/mutual/:other_id", handlers.GetMutualFollowers)

	router.GET("/users/:user_id/metrics", handlers.GetUserMetrics)

	router.GET("/users/:user_id/music-similarity", handlers.GetMusicSimilarity)

	router.GET("/users/:user_id/music-cosine", handlers.GetCosineSimilarity)

	router.POST("/listen", handlers.AddListen)
	logger.Log.Info("Server running on PORT:8080")
	router.Run(":8080")
	
}