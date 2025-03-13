package repository

import (
	"context"
	"sync"

	e "github.com/jasoncheung94/url-shortener/internal/errors"
	"github.com/jasoncheung94/url-shortener/internal/logger"
	"github.com/jasoncheung94/url-shortener/internal/shortener/model"
)

// InMemoryRepo is a repo that handles the URL.
type InMemoryRepo struct {
	mu      sync.RWMutex
	store   map[string]model.URL
	counter uint64 // not a good solution if scaled.
}

var _ URL = &InMemoryRepo{}

// NewInMemory returns an instance of the in memory repo.
func NewInMemory() *InMemoryRepo {
	return &InMemoryRepo{
		mu:      sync.RWMutex{},
		store:   make(map[string]model.URL),
		counter: 1,
	}
}

// SaveURL saves the URL to memory.
func (r *InMemoryRepo) SaveURL(_ context.Context, data *model.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	lookupURL := data.ShortURL
	if data.CustomURL != nil {
		lookupURL = *data.CustomURL
	}

	if _, ok := r.store[lookupURL]; ok {
		logger.Logger.Info("short url already exists:", "url", lookupURL)
		return e.NewConflictError("short url already exists")
	}

	data.ID = int64(len(r.store) + 1)
	r.store[data.ShortURL] = *data
	return nil
}

// GetURL retrieves the URL from memory.
func (r *InMemoryRepo) GetURL(_ context.Context, shortURL string) (*model.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if data, ok := r.store[shortURL]; ok {
		return &data, nil
	}
	return nil, e.NewNotFoundError("failed to get original url")
}

// IncrementCounter returns the next counter value and increments it.
func (r *InMemoryRepo) IncrementCounter() (uint64, error) {
	r.mu.Lock() // Lock to ensure only one goroutine can increment the counter at a time.
	defer r.mu.Unlock()
	// Extra safety check for in memory solution.
	if r.counter < uint64(len(r.store)) {
		r.counter = uint64(len(r.store) + 1)
	}

	counter := r.counter
	r.counter++ // Increment the counter for the next URL.

	return counter, nil
}
