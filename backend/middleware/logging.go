package middleware

import (
	"time"

	"github.com/atharv-3105/Social_Music_Graph/logger"
	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return func(c * gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		logger.Log.WithFields(map[string]interface{}{
			"method": c.Request.Method,
			"path":	  c.Request.URL.Path,
			"status": c.Writer.Status(),
			"latency": duration.String(),
			"clientIP": c.ClientIP(),
		}).Info("incoming request")
	}
}