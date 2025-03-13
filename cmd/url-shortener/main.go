package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jasoncheung94/url-shortener/config"
	_ "github.com/jasoncheung94/url-shortener/docs" // swagger docs required import
	"github.com/jasoncheung94/url-shortener/internal/database"
	"github.com/jasoncheung94/url-shortener/internal/logger"
	"github.com/jasoncheung94/url-shortener/internal/router"
	"github.com/jasoncheung94/url-shortener/internal/server"
	"github.com/jasoncheung94/url-shortener/internal/shortener"
	"github.com/jasoncheung94/url-shortener/internal/shortener/cache"
	"github.com/jasoncheung94/url-shortener/internal/shortener/repository"
	"github.com/jasoncheung94/url-shortener/internal/validator"
	_ "github.com/lib/pq" // PostgreSQL driver for database/sql
	"github.com/spf13/viper"
)

func main() {
	fmt.Println("Starting URL Shortener!")
	logger.SetupLogger()
	validator.SetupValidator()
	config.Setup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	repo, cleanup, err := database.SetupDB(ctx, database.DBMongo)
	if err != nil {
		log.Println("Failed to get database repo", err)
	}

	// Create a new Redis client
	rdb, err := database.NewRedis()
	if err != nil {
		log.Panic("Failed to get redis client", err)
		return
	} else {
		defer rdb.Close()
	}

	redis := cache.NewRedis(rdb)

	cachedRepo := repository.NewCache(repo, redis)
	service := shortener.NewService(cachedRepo)
	handler := shortener.NewHandler(service)
	router := router.New(handler)

	server.Start(router, viper.GetString("port"), cleanup)
}
