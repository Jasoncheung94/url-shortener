package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jasoncheung94/url-shortener/internal/errors"
	e "github.com/jasoncheung94/url-shortener/internal/errors"
	"github.com/jasoncheung94/url-shortener/internal/shortener/model"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestSaveURL_Success(t *testing.T) {
	// Create a new MongoDB test environment with proper client options
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	// URL data for the test
	urlData := &model.URL{
		ShortURL:    "short123",
		OriginalURL: "http://example.com",
		CreatedAt:   time.Now(),
	}

	mt.Run("Test SaveURL Success", func(mt *mtest.T) {
		// Mock the InsertOne call to succeed.
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		// Define MongoRepo using the mocked client and collection.
		repo := NewMongoDB(mt.Coll)

		// Call the SaveURL method.
		err := repo.SaveURL(context.Background(), urlData)

		// Assert no errors occurred.
		assert.Nil(t, err)
	})

	// The mt object will be automatically cleaned up after the test completes.
}

func TestSaveURL_DuplicateError(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	urlData := &model.URL{
		ShortURL:    "short123",
		OriginalURL: "http://example.com",
		CreatedAt:   time.Now(),
	}

	mt.Run("Test SaveURL Duplicate Error", func(mt *mtest.T) {
		// Mock the InsertOne call to simulate a duplicate key error (code 11000).
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index:   1,
			Code:    11000,
			Message: "duplicate error",
		}))

		// Define MongoRepo using the mocked client and collection.
		repo := NewMongoDB(mt.Coll)

		// Call the SaveURL method.
		err := repo.SaveURL(context.Background(), urlData)

		// Assert the expected duplicate error occurs.
		assert.Equal(t, errors.NewConflictError("short url already exists"), err)
	})
}

func TestGetURL_Success(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Test GetURL Success", func(mt *mtest.T) {
		// URL data to return for the mock query
		urlData := model.URL{
			ShortURL:    "short123",
			OriginalURL: "http://example.com",
			CreatedAt:   time.Now(),
		}

		// Mock the FindOne call to return a successful response with a cursor-like result
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.collection", mtest.FirstBatch,
			bson.D{
				{Key: "short_url", Value: urlData.ShortURL},
				{Key: "original_url", Value: urlData.OriginalURL},
				{Key: "created_at", Value: urlData.CreatedAt},
			},
		))

		// Define MongoRepo using the mocked client and collection.
		repo := NewMongoDB(mt.Coll)

		// Call the GetURL method
		result, err := repo.GetURL(context.Background(), "short123")

		// Assert that the result is correct and there is no error
		assert.Nil(t, err)
		assert.Equal(t, urlData.ShortURL, result.ShortURL)
		assert.Equal(t, urlData.OriginalURL, result.OriginalURL)
	})

	// The mt object will be automatically cleaned up after the test completes.
}

func TestGetURL_NotFound(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Test GetURL Not Found", func(mt *mtest.T) {
		// Mock the FindOne call to simulate no document found (mongo.ErrNoDocuments).
		mt.AddMockResponses(
			mtest.CreateCursorResponse(0, "test.collection", "firstBatch"), // Empty response to simulate "not found"
		)
		// Define MongoRepo using the mocked client and collection.
		repo := NewMongoDB(mt.Coll)

		// Call the GetURL method.
		result, err := repo.GetURL(context.Background(), "nonexistent")

		// Assert that the error is the "not found" error
		assert.Nil(t, result)
		assert.Equal(t, e.NewNotFoundError("url with short_url 'nonexistent' not found"), err)
	})
}

func TestIncrementCounter_Success(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Test IncrementCounter Success", func(mt *mtest.T) {
		// Mock the CountDocuments call to return a count of 5 documents in the collection.

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.collection", mtest.FirstBatch,
			bson.D{{Key: "n", Value: int32(5)}}, // Simulating 5 documents in the collection.
		))

		// Define MongoRepo using the mocked client and collection.
		repo := NewMongoDB(mt.Coll)

		// Initially, counter is set to 1.
		// The next count after 5 should be 6, so it will increment it.
		repo.counter = 1

		// Call the IncrementCounter method.
		count, err := repo.IncrementCounter()

		// Assert that the incremented counter value is 6.
		assert.Nil(t, err)
		assert.Equal(t, uint64(6), count)
	})
}

func TestIncrementCounter_Error(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Test IncrementCounter Error", func(mt *mtest.T) {
		// Mock the CountDocuments call to return an error.
		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "test.collection", "batch1",
				// Simulate an error by not returning a valid "count"
				bson.D{
					{Key: "error", Value: "failed to count documents"},
				}),
		)

		// Define MongoRepo using the mocked client and collection.
		repo := NewMongoDB(mt.Coll)

		// Initially, counter is set to 1.
		repo.counter = 1

		// Call the IncrementCounter method.
		count, err := repo.IncrementCounter()

		// Assert that the error is returned and count is 0.
		assert.NotNil(t, err)
		assert.Equal(t, uint64(0), count)
	})
}
