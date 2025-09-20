package redis

import (
	"chat-app/backend/config"

	"github.com/go-redis/redis/v8"
)

func NewRedisClient(cfg *config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0, // use default DB
	})
	return rdb
}
