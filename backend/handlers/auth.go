package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func SpotifyAuthStart(c *gin.Context){
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	redirect := os.Getenv("SPOTIFY_REDIRECT_URI")

	if clientID == "" || redirect == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error" : "spotify creds not set"})
		return
	}
	scope := "user-read-recently-played user-read-email playlist-read-private user-library-read"
	authURL := "https://accounts.spotify.com/authorize" + 
				"?response_type=code&client_id=" + clientID + 
				"&scope=" + urlEncode(scope) + 
				"&redirect_uri=" + urlEncode(redirect)

	c.Redirect(http.StatusFound, authURL)
}

func SpotifyAuthCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error" : "missing code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code" : code})
}

func ProxyToML(c *gin.Context) {

	user := c.Param("user_id")
	c.JSON(200, gin.H{"user":user, "recommendations":[]string{}})
}

func urlEncode(s string) string {
	return s 
}