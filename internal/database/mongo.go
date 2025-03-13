package database

import (
	"context"
	"fmt"
	"log"

	l "github.com/jasoncheung94/url-shortener/internal/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InitMongo initializes the MongoDB connection and sets up collection and indexes on start up every time!
func InitMongo(ctx context.Context, uri string) (*mongo.Client, error) {
	println("MongoDB starting up......")
	clientOpts := options.Client().ApplyURI(uri).SetMaxPoolSize(10)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	// Ping to confirm connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	databases, err := client.ListDatabases(ctx, bson.D{{Key: "name", Value: "url_shortener_db"}})
	if err != nil {
		l.Logger.Warn("failed to list mongodb databases")
	}

	// Check if database is already created. Return early otherwise create.
	l.Logger.Info("Checking if mongodb databases are setup")
	for _, db := range databases.Databases {
		if db.Name == "url_shortener_db" {
			return client, nil
		}
	}

	l.Logger.Info("Setting up indexes and database for url shortener")
	// Create/Use Database.
	db := client.Database("url_shortener_db")
	createCollectionWithValidation(ctx, db)

	// Use the collection.
	collection := db.Collection("urls")
	createIndexes(ctx, collection)

	return client, nil
}

func createCollectionWithValidation(ctx context.Context, db *mongo.Database) {
	// Add JSON Schema for validation
	validator := bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"required": []string{"original_url", "short_url", "created_at"},
			"properties": bson.M{
				"original_url": bson.M{
					"bsonType":    "string",
					"description": "must be a string and is required",
				},
				"short_url": bson.M{
					"bsonType":    "string",
					"description": "must be a string and is required",
				},
				"custom_url": bson.M{
					"bsonType":    []string{"string", "null"},
					"description": "optional string",
				},
				"expiration_date": bson.M{
					"bsonType":    bson.A{"date", "null"}, // example with bson.A
					"description": "optional expiration date",
				},
				"created_at": bson.M{
					"bsonType":    "date",
					"description": "must be a date and is required",
				},
			},
		},
	}

	opts := options.CreateCollection().SetValidator(validator)
	err := db.CreateCollection(ctx, "urls", opts)
	if err != nil {
		if cmdErr, ok := err.(mongo.CommandError); ok && cmdErr.Code == 48 {
			fmt.Println("Collection already exists. Skipping creation.")
		} else {
			log.Fatal("creating collection:", err)
		}
	} else {
		fmt.Println("Collection created with schema validation.")
	}
}

func createIndexes(ctx context.Context, collection *mongo.Collection) {
	// Create index on short_url.
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "short_url", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Fatal("Creating index", err)
	}
}
