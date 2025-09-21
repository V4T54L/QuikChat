package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
)

func NewClient(redisURL string) *redis.Client {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Could not parse Redis URL: %v", err)
	}

	client := redis.NewClient(opts)

	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	log.Println("Successfully connected to Redis")
	return client
}

