package database

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(url string) *redis.Client {
	opts, err := redis.ParseURL(url)
	if err != nil {
		// Fallback to simple host:port
		opts = &redis.Options{
			Addr: url,
			DB:   0,
		}
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		panic("Failed to connect to Redis: " + err.Error())
	}

	return client
}
