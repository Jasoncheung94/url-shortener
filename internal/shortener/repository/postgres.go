package repository

import (
	"context"
	"database/sql"
	"errors"

	e "github.com/jasoncheung94/url-shortener/internal/errors"
	l "github.com/jasoncheung94/url-shortener/internal/logger"
	"github.com/jasoncheung94/url-shortener/internal/shortener/model"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// PostgresRepo is a repository that interacts with a PostgreSQL database for URL storage and retrieval.
type PostgresRepo struct {
	db *sqlx.DB
}

var _ URL = &PostgresRepo{}

// NewPostgres an instance of PostgresRepo.
func NewPostgres(db *sqlx.DB) *PostgresRepo {
	return &PostgresRepo{db}
}

// SaveURL inserts a new URL into the database and returns the ID of the newly created URL.
func (r *PostgresRepo) SaveURL(ctx context.Context, data *model.URL) error {
	query := `INSERT INTO urls
	(original_url, short_url, custom_url, expiration_date, created_at)
	VALUES
	($1, $2, $3, $4, $5)
	RETURNING id`

	// Use QueryRow to retrieve the auto-generated ID.
	err := r.db.QueryRowContext(ctx, query,
		data.OriginalURL,
		data.ShortURL,
		data.CustomURL,
		data.ExpirationDate,
		data.CreatedAt,
	).Scan(&data.ID) // Scanning the returned ID into the data struct
	if err != nil {
		if pq, ok := err.(*pq.Error); ok && pq.Code == "23505" {
			return e.NewConflictError("short url already exists: try again!")
		}
		return errors.New("failed to insert url:" + err.Error())
	}

	return nil
}

// GetURL retrieves a URL record by its short URL.
func (r *PostgresRepo) GetURL(c context.Context, shortURL string) (*model.URL, error) {
	query := `SELECT id, original_url, short_url, custom_url, expiration_date, created_at FROM urls WHERE short_url = $1`

	var data model.URL
	// Use Get since we expect at most one result (single row).
	err := r.db.GetContext(c, &data, query, shortURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("short URL not found")
		}
		return nil, errors.New("failed to find short URL")
	}

	return &data, nil
}

// IncrementCounter increments the counter and returns it's value.
func (r *PostgresRepo) IncrementCounter() (uint64, error) {
	var counter uint64
	err := r.db.Get(&counter, "SELECT nextval('url_shortener_seq');")
	if err != nil {
		l.Logger.Error("failed to get next counter", "repo", err)
		return counter, errors.New("failed to retrieve counter")
	}
	return counter, nil
}
