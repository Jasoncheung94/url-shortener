package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jasoncheung94/url-shortener/internal/mocks"
	"github.com/jasoncheung94/url-shortener/internal/shortener/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

//nolint:paralleltest
func TestCacheWrapper_SaveURL(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockURL(ctrl)
	mockCache := mocks.NewMockRedisInterface(ctrl)

	c := NewCache(mockRepo, mockCache)

	url := &model.URL{
		ShortURL:    "abc123",
		OriginalURL: "https://example.com",
	}

	t.Run("success - repo and cache", func(t *testing.T) {
		mockRepo.EXPECT().SaveURL(gomock.Any(), url).Return(nil)
		mockCache.EXPECT().Set(gomock.Any(), "shorturl:abc123", url, time.Hour).Return(nil)

		err := c.SaveURL(context.Background(), url)
		assert.NoError(t, err)
	})

	t.Run("repo failure - returns error", func(t *testing.T) {
		mockRepo.EXPECT().SaveURL(gomock.Any(), url).Return(errors.New("db error"))

		err := c.SaveURL(context.Background(), url)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})

	t.Run("cache failure - logs only", func(t *testing.T) {
		mockRepo.EXPECT().SaveURL(gomock.Any(), url).Return(nil)
		mockCache.EXPECT().Set(gomock.Any(), "shorturl:abc123", url, time.Hour).Return(errors.New("redis error"))

		err := c.SaveURL(context.Background(), url)
		assert.NoError(t, err)
	})
}

//nolint:paralleltest
func TestCacheWrapper_GetURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockURL(ctrl)
	mockCache := mocks.NewMockRedisInterface(ctrl)

	c := NewCache(mockRepo, mockCache)

	expectedURL := &model.URL{
		ShortURL:    "abc123",
		OriginalURL: "https://example.com",
	}

	t.Run("cache hit", func(t *testing.T) {
		mockCache.EXPECT().Get(gomock.Any(), "shorturl:abc123", gomock.Any()).
			DoAndReturn(func(_ context.Context, _ string, dest any) error {
				ptr := dest.(**model.URL)
				*ptr = expectedURL
				return nil
			})

		url, err := c.GetURL(context.Background(), "abc123")
		assert.NoError(t, err)
		assert.Equal(t, expectedURL, url)
	})

	t.Run("cache miss, db hit, cache set succeeds", func(t *testing.T) {
		mockCache.EXPECT().Get(gomock.Any(), "shorturl:abc123", gomock.Any()).Return(errors.New("cache miss"))
		mockRepo.EXPECT().GetURL(gomock.Any(), "abc123").Return(expectedURL, nil)
		mockCache.EXPECT().Set(gomock.Any(), "shorturl:abc123", expectedURL, time.Hour).Return(nil)

		url, err := c.GetURL(context.Background(), "abc123")
		assert.NoError(t, err)
		assert.Equal(t, expectedURL, url)
	})

	t.Run("cache miss, db hit, cache set fails", func(t *testing.T) {
		mockCache.EXPECT().Get(gomock.Any(), "shorturl:abc123", gomock.Any()).Return(errors.New("cache miss"))
		mockRepo.EXPECT().GetURL(gomock.Any(), "abc123").Return(expectedURL, nil)
		mockCache.EXPECT().Set(gomock.Any(), "shorturl:abc123", expectedURL, time.Hour).Return(errors.New("redis error"))

		url, err := c.GetURL(context.Background(), "abc123")
		assert.NoError(t, err)
		assert.Equal(t, expectedURL, url)
	})

	t.Run("cache miss, db error", func(t *testing.T) {
		mockCache.EXPECT().Get(gomock.Any(), "shorturl:abc123", gomock.Any()).Return(errors.New("cache miss"))
		mockRepo.EXPECT().GetURL(gomock.Any(), "abc123").Return(nil, errors.New("db error"))

		url, err := c.GetURL(context.Background(), "abc123")
		assert.Error(t, err)
		assert.Nil(t, url)
	})
}

//nolint:paralleltest
func TestCacheWrapper_IncrementCounter(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockURL(ctrl)
	mockCache := mocks.NewMockRedisInterface(ctrl)

	c := NewCache(mockRepo, mockCache)

	t.Run("cache hit", func(t *testing.T) {
		mockCache.EXPECT().Increment(gomock.Any(), "url_shortener_counter").Return(int64(100), nil)

		val, err := c.IncrementCounter()
		assert.NoError(t, err)
		assert.Equal(t, uint64(100), val)
	})

	t.Run("cache failure, fallback to repo", func(t *testing.T) {
		mockCache.EXPECT().Increment(gomock.Any(), "url_shortener_counter").Return(int64(0), errors.New("redis error"))
		mockRepo.EXPECT().IncrementCounter().Return(uint64(42), nil)

		val, err := c.IncrementCounter()
		assert.NoError(t, err)
		assert.Equal(t, uint64(42), val)
	})
}
