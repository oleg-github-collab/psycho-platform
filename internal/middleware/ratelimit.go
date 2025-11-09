package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redis *redis.Client
	limit int
	window time.Duration
}

func NewRateLimiter(redis *redis.Client, requestsPerMinute int) *RateLimiter {
	return &RateLimiter{
		redis:  redis,
		limit:  requestsPerMinute,
		window: time.Minute,
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if rl.redis == nil {
			c.Next()
			return
		}

		userID := c.GetString("user_id")
		if userID == "" {
			userID = c.ClientIP()
		}

		key := fmt.Sprintf("ratelimit:%s", userID)
		ctx := context.Background()

		// Get current count
		count, err := rl.redis.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			c.Next()
			return
		}

		if count >= rl.limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		// Increment counter
		pipe := rl.redis.Pipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, rl.window)
		_, err = pipe.Exec(ctx)
		if err != nil {
			c.Next()
			return
		}

		c.Next()
	}
}
