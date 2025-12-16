package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RateLimitMiddleware(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := "rl:api:" + clientIP

		// Pipeline for atomicity/performance
		pipe := rdb.Pipeline()
		incr := pipe.Incr(c, key)
		pipe.Expire(c, key, window)
		_, err := pipe.Exec(c)

		if err == nil && incr.Val() > int64(limit) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			return
		}

		c.Next()
	}
}
