package repository

import (
	"context"
	"time"

	l "github.com/jasoncheung94/url-shortener/internal/logger"
	"github.com/jasoncheung94/url-shortener/internal/shortener/cache"
	"github.com/jasoncheung94/url-shortener/internal/shortener/model"
)

// CacheWrapper implements the repository interface.
type CacheWrapper struct {
	repo  URL
	cache cache.RedisInterface
}

// NewCache returns a new instance of the cache wrapper.
func NewCache(repo URL, cache cache.RedisInterface) *CacheWrapper {
	return &CacheWrapper{repo, cache}
}

var ttl = time.Hour // short ttl, every time a cache hit, redis will set a new ttl of 2 hours. See Redis code.

// SaveURL saves the URL to redis using a cache key.
func (c *CacheWrapper) SaveURL(ctx context.Context, data *model.URL) error {
	err := c.repo.SaveURL(ctx, data)
	if err != nil {
		return err
	}

	cacheKey := "shorturl:" + data.ShortURL
	err = c.cache.Set(ctx, cacheKey, data, ttl)
	if err != nil {
		// Cache doesn't cause hard failure. DB still worked.
		l.Logger.Error("failed to set key", "cache", cacheKey, "error", err.Error())
	}
	return nil
}

// GetURL gets the URL from redis using a cache key.
func (c *CacheWrapper) GetURL(ctx context.Context, shortURL string) (*model.URL, error) {
	var data *model.URL
	cacheKey := "shorturl:" + shortURL

	if err := c.cache.Get(ctx, "shorturl:"+shortURL, &data); err == nil {
		return data, nil
	}

	data, err := c.repo.GetURL(ctx, shortURL)
	if err != nil {
		return nil, err
	}

	err = c.cache.Set(ctx, cacheKey, data, ttl)
	if err != nil {
		// Cache doesn't cause hard failure. DB still worked.
		l.Logger.Error("failed to set key", "cache", cacheKey, "error", err.Error())
	}
	return data, nil
}

// IncrementCounter increments counter and fetches latest value from redis
func (c *CacheWrapper) IncrementCounter() (uint64, error) {
	if counterValue, err := c.cache.Increment(context.Background(), "url_shortener_counter"); err == nil {
		return uint64(counterValue), nil
	} else {
		l.Logger.Error(err.Error())
	}

	// Check redis otherwise fallback to repository solution.
	return c.repo.IncrementCounter()
}
