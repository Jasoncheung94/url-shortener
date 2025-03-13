package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisInterface represents the methods for interacting with redis.
type RedisInterface interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string, dest any) error
	Increment(ctx context.Context, key string) (int64, error)
}

// RedisCache represents the redis client.
type RedisCache struct {
	client *redis.Client
}

// NewRedis returns new instance of RedisCache.
func NewRedis(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

// Set stores a key-value pair in Redis with a TTL (time-to-live)
func (r *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	// Serialize the object to JSON
	serializedValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// Store the serialized data in Redis
	err = r.client.Set(ctx, key, serializedValue, ttl).Err()
	return err
}

// Get retrieves the value associated with a given key from Redis
func (r *RedisCache) Get(ctx context.Context, key string, dest any) error {
	val, err := r.client.GetEx(ctx, key, time.Hour*2).Result() // Get key, refresh ttl on each call.
	if err != nil {
		return err
	}

	if val != "" {
		err := json.Unmarshal([]byte(val), &dest)
		if err != nil {
			return err
		}
		// Parse val into your object, for example using json.Unmarshal or another method
	}
	return nil
}

// Increment increases the counter value for a given key and returns the updated value
func (r *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	// Use Redis' INCR command to atomically increment the value of a key
	val, err := r.client.Incr(ctx, key).Result()
	return val, err
}
