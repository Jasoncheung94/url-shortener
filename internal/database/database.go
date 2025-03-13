package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jasoncheung94/url-shortener/internal/shortener/repository"
	"github.com/spf13/viper"
)

// DBType represents the type of database being used.
type DBType string

// Constants for different types of databases.
const (
	// DBMongo represents a MongoDB database.
	DBMongo DBType = "mongodb"
	// DBSQL represents a SQL database (e.g., MySQL, PostgreSQL).
	DBSQL DBType = "sql"
	// DBMem represents an in-memory database.
	DBMem DBType = "memory"
)

// SetupDB sets up the DB client/connection and returns the repo and a cleanup function for graceful shutdown.
func SetupDB(ctx context.Context, dbType DBType) (repository.URL, func(), error) {
	switch dbType {
	case DBMongo:
		// SH to container. Run "mongosh --username=admin" to access shell.
		mongoURI := viper.GetString("mongo_uri")
		client, err := InitMongo(ctx, mongoURI)
		if err != nil {
			return nil, nil, fmt.Errorf("mongo connection failed: %w", err)
		}
		db := client.Database("url_shortener_db")
		collection := db.Collection("urls")
		cleanup := func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			fmt.Println("Shutting down mongo!")
			if err := client.Disconnect(shutdownCtx); err != nil {
				log.Printf("Mongo disconnect failed: %v", err)
			}
		}
		return repository.NewMongoDB(collection), cleanup, nil

	case DBSQL:
		// Can check the disconnection manually by running the following query:
		// SELECT pid, usename, application_name, client_addr, state, query
		// FROM pg_stat_activity ;
		db, err := NewPostgres(viper.GetString("POSTGRES_URI"))
		if err != nil {
			return nil, nil, fmt.Errorf("sql connection failed: %w", err)
		}
		cleanup := func() {
			fmt.Println("Shutting down Postgres!")
			if err := db.Close(); err != nil {
				log.Printf("sql disconnect failed: %v", err)
			}
		}
		return repository.NewPostgres(db), cleanup, nil

	default:
		return repository.NewInMemory(), nil, nil
	}
}
