package config

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

// ConnectRedis initializes Redis connection
func ConnectRedis() {
	addr := GetEnv("REDIS_HOST", "localhost") + ":" + GetEnv("REDIS_PORT", "6379")
	password := GetEnv("REDIS_PASSWORD", "")
	db := 0 // default DB

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx := context.Background()
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
		log.Println("Application will continue without caching")
		RedisClient = nil // Set to nil so we can check availability
	} else {
		log.Println("Redis connected successfully")
	}
}

// GetRedis returns the Redis client instance
func GetRedis() *redis.Client {
	return RedisClient
}
