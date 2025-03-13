package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"

	e "github.com/jasoncheung94/url-shortener/internal/errors"
	"github.com/jasoncheung94/url-shortener/internal/shortener/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoRepo represents the methods for handling a URL.
type MongoRepo struct {
	client  *mongo.Collection
	counter uint64
	sync.RWMutex
}

var _ URL = &MongoRepo{}

// NewMongoDB returns an instance of MongoRepo.
func NewMongoDB(client *mongo.Collection) *MongoRepo {
	return &MongoRepo{
		client:  client,
		counter: 1,
		RWMutex: sync.RWMutex{},
	}
}

// SaveURL saves URL data to MongoDB.
func (m *MongoRepo) SaveURL(ctx context.Context, data *model.URL) error {
	result, err := m.client.InsertOne(ctx, map[string]interface{}{
		"short_url":       data.ShortURL,
		"original_url":    data.OriginalURL,
		"created_at":      data.CreatedAt,
		"expiration_date": data.ExpirationDate,
		"custom_url":      data.CustomURL,
	})
	if err != nil {
		if isDuplicateError(err) {
			return e.NewConflictError("short url already exists")
		}
		return errors.New("failed to insert url:" + err.Error())
	}
	fmt.Println("Created!", result.InsertedID)
	return nil
}

func isDuplicateError(err error) bool {
	var e mongo.WriteException
	if errors.As(err, &e) {
		for _, we := range e.WriteErrors {
			if we.Code == 11000 {
				return true
			}
		}
	}
	return false
}

// GetURL retrieves the URL from MongoDB.
// Example to Define projection: map MongoDB field names to Go struct fields
//
//	projection := bson.M{
//		"short_url":       1, // MongoDB field "short_url"
//		"original_url":    1, // MongoDB field "original_url"
//		"custom_url":      1, // MongoDB field "custom_url"
//		"expiration_date": 1, // MongoDB field "expiration_date"
//		"created_at":      1, // MongoDB field "created_at"
//	}
func (m *MongoRepo) GetURL(ctx context.Context, shortURL string) (*model.URL, error) {
	var result model.URL
	err := m.client.FindOne(ctx, bson.D{
		{Key: "short_url", Value: shortURL},
	}).Decode(&result)
	// }, options.FindOne().SetProjection(projection)).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			// No document found
			return nil, e.NewNotFoundError("url with short_url '%s' not found", shortURL)
		}
		// Other errors
		return nil, fmt.Errorf("error while retrieving URL: %v", err)
	}

	return &result, nil
}

// IncrementCounter increments the counter and returns it's value.
// Hacky solution if redis + replicas fail. Not expecting to reach this code but safety net.
func (m *MongoRepo) IncrementCounter() (uint64, error) {
	m.Lock()
	defer m.Unlock()
	count, err := m.client.CountDocuments(context.Background(), bson.D{})
	if err != nil {
		return 0, err
	}

	if m.counter < uint64(count) {
		m.counter = uint64(count + 1)
	}

	return m.counter, nil
}
