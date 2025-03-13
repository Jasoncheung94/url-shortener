package shortener

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	l "github.com/jasoncheung94/url-shortener/internal/logger"
	"github.com/jasoncheung94/url-shortener/internal/shortener/model"
	"github.com/jasoncheung94/url-shortener/internal/shortener/repository"
)

//go:generate mockgen -source=service.go -destination=../mocks/mock_service.go -package=mocks

// Service defines the methods for interacting with the URL service.
// Run go generate ./... to generate mocks or run manually the command.
type Service interface {
	SaveURL(ctx context.Context, data *model.URL) (string, error)
	GetURL(ctx context.Context, shortURL string) (*model.URL, error)
}

// NewService returns an instance of Service.
func NewService(repo repository.URL) Service {
	return &shortenerService{repo: repo}
}

type shortenerService struct {
	repo repository.URL
}

// isValidShortURL ensures the short URL is only alphanumeric
func isValidShortURL(shortURL string) bool {
	match, _ := regexp.MatchString(`^[a-zA-Z0-9]{1,10}$`, shortURL)
	return match
}

// ValidateURL checks if the provided URL is valid and has a proper scheme.
func ValidateURL(rawURL string) error {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		l.Logger.Info("validation failed for url:", "service", err)
		return errors.New("invalid URL format")
	}

	// Ensure the scheme is either http or https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("URL must start with http:// or https://")
	}

	// Ensure there is a host (not just a path)
	if parsedURL.Host == "" {
		return errors.New("URL must contain a valid domain")
	}

	// Prevent self-referencing redirects
	if strings.HasPrefix(rawURL, "http://localhost:8080/") { // Change this to your actual domain
		return errors.New("URL must contain a valid domain")
	}

	return nil
}

func (s *shortenerService) SaveURL(ctx context.Context, data *model.URL) (string, error) {
	err := ValidateURL(data.OriginalURL)
	if err != nil {
		return "", errors.New("invalid url")
	}

	counter, err := s.repo.IncrementCounter()
	if err != nil {
		return "", err
	}

	var shortURL string
	// base62HashedCounter := hashCounter(counter)
	if data.CustomURL == nil || *data.CustomURL == "" {
		shortURL = EncodeBase62(counter)
	} else {
		shortURL = *data.CustomURL
	}

	log.Println("Hashed url with counter:", data.OriginalURL, counter, shortURL)
	data.ShortURL = shortURL
	data.CreatedAt = time.Now().UTC()

	err = s.repo.SaveURL(ctx, data)
	if err != nil {
		return "", fmt.Errorf("shortener/service: failed to create url: %w", err)
	}
	return shortURL, err
}

func (s *shortenerService) GetURL(ctx context.Context, shortURL string) (*model.URL, error) {
	if !isValidShortURL(shortURL) {
		return nil, errors.New("invalid url")
	}

	data, err := s.repo.GetURL(ctx, shortURL)
	if err != nil {
		return nil, fmt.Errorf("shortener/service: failed to get url: %w", err)
	}
	return data, nil
}
