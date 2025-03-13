package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jasoncheung94/url-shortener/internal/ptr"
	"github.com/jasoncheung94/url-shortener/internal/shortener/model"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestPostgresSaveURL(t *testing.T) {
	t.Parallel()
	// Create a mock database and a mock sqlx.DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")

	// Create a repository
	repo := NewPostgres(sqlxDB)

	// Define test data
	data := model.URL{
		OriginalURL:    "https://example.com",
		ShortURL:       "short123",
		CustomURL:      ptr.Of("custom123"),
		ExpirationDate: ptr.Of(time.Now().Add(24 * time.Hour)),
		CreatedAt:      time.Now(),
	}

	// Set up the expected query and mock behavior
	mock.ExpectQuery(`INSERT INTO urls`).
		WithArgs(data.OriginalURL, data.ShortURL, data.CustomURL, data.ExpirationDate, data.CreatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Call the method
	err = repo.SaveURL(context.Background(), &data)

	// Assert the expectations
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetURL(t *testing.T) {
	t.Parallel()
	// Create a mock database and a mock sqlx.DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")

	// Create a repository
	repo := NewPostgres(sqlxDB)

	// Define test data
	shortURL := "short123"
	expectedURL := &model.URL{
		ID:             1,
		OriginalURL:    "https://example.com",
		ShortURL:       shortURL,
		CustomURL:      ptr.Of("custom123"),
		ExpirationDate: ptr.Of(time.Now().Add(24 * time.Hour)),
		CreatedAt:      time.Now(),
	}

	// Set up the expected query and mock behavior
	mock.ExpectQuery(
		`SELECT id, original_url, short_url, custom_url, expiration_date, created_at FROM urls WHERE short_url =`,
	).WithArgs(shortURL).
		WillReturnRows(sqlmock.NewRows(
			[]string{"id", "original_url", "short_url", "custom_url", "expiration_date", "created_at"},
		).AddRow(
			expectedURL.ID,
			expectedURL.OriginalURL,
			expectedURL.ShortURL,
			expectedURL.CustomURL,
			expectedURL.ExpirationDate,
			expectedURL.CreatedAt,
		))

	// Call the method
	url, err := repo.GetURL(context.Background(), shortURL)

	// Assert the expectations
	assert.NoError(t, err)
	assert.Equal(t, expectedURL, url)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIncrementCounter(t *testing.T) {
	t.Parallel()
	// Create a mock database and a mock sqlx.DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")

	// Create a repository
	repo := NewPostgres(sqlxDB)

	// Define the expected behavior for the counter increment
	mock.ExpectQuery(`SELECT nextval\('url_shortener_seq'\);`).
		WillReturnRows(sqlmock.NewRows([]string{"nextval"}).AddRow(123))

	// Call the method
	counter, _ := repo.IncrementCounter()

	// Assert the expectations
	assert.Equal(t, uint64(123), counter)
	assert.NoError(t, mock.ExpectationsWereMet())
}
