package database

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

// NewRedis returns a new redis client.
func NewRedis() (*redis.Client, error) {
	ctx := context.Background()
	password := "masterpassword"

	addr := viper.GetString("REDIS_ADDR")
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Printf("Primary Redis connection failed: %v", err)
	}
	return rdb, nil
}
