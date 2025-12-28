package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"sentul-golf-be/config"
)

// Cache TTL constants
const (
	CacheTTLHolesList   = 1 * time.Hour
	CacheTTLHoleDetail  = 24 * time.Hour
	CacheTTLNewsList    = 15 * time.Minute
	CacheTTLNewsDetail  = 1 * time.Hour
	CacheTTLEventsList  = 15 * time.Minute
	CacheTTLEventDetail = 1 * time.Hour
)

// IsRedisAvailable checks if Redis client is connected
func IsRedisAvailable() bool {
	return config.GetRedis() != nil
}

// CacheGet retrieves cached data and unmarshals it into dest
func CacheGet(ctx context.Context, key string, dest interface{}) error {
	if !IsRedisAvailable() {
		return fmt.Errorf("redis not available")
	}

	client := config.GetRedis()
	val, err := client.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

// CacheSet stores data in cache with TTL
func CacheSet(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !IsRedisAvailable() {
		return fmt.Errorf("redis not available")
	}

	client := config.GetRedis()
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return client.Set(ctx, key, jsonData, ttl).Err()
}

// CacheDelete removes a single cache key
func CacheDelete(ctx context.Context, key string) error {
	if !IsRedisAvailable() {
		return nil // Silently skip if Redis not available
	}

	client := config.GetRedis()
	return client.Del(ctx, key).Err()
}

// CacheDeletePattern removes all keys matching pattern (e.g., "news:*")
func CacheDeletePattern(ctx context.Context, pattern string) error {
	if !IsRedisAvailable() {
		return nil // Silently skip if Redis not available
	}

	client := config.GetRedis()
	
	// Use SCAN to find all matching keys
	var cursor uint64
	var keys []string
	
	for {
		var scanKeys []string
		var err error
		scanKeys, cursor, err = client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		
		keys = append(keys, scanKeys...)
		
		if cursor == 0 {
			break
		}
	}
	
	// Delete all found keys
	if len(keys) > 0 {
		return client.Del(ctx, keys...).Err()
	}
	
	return nil
}

// BuildCacheKey builds a cache key from parts
func BuildCacheKey(parts ...interface{}) string {
	key := ""
	for i, part := range parts {
		if i > 0 {
			key += ":"
		}
		key += fmt.Sprintf("%v", part)
	}
	return key
}
