package repository

import (
	"context"

	"github.com/jasoncheung94/url-shortener/internal/shortener/model"
)

//go:generate mockgen -source=repository.go -destination=../../mocks/mock_repo.go -package=mocks

// URL represents the methods for interacting with URL storage.
type URL interface {
	SaveURL(ctx context.Context, data *model.URL) error
	GetURL(ctx context.Context, shortURL string) (*model.URL, error)
	IncrementCounter() (uint64, error)
}
