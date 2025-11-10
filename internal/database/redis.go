package database

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(url string) (*redis.Client, error) {
	if url == "" {
		url = "localhost:6379"
	}

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
		return nil, err
	}

	return client, nil
}
