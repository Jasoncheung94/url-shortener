package repository

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-migrate/migrate/v4"
	postm "github.com/golang-migrate/migrate/v4/database/postgres" // golang-migrate postgres driver
	_ "github.com/golang-migrate/migrate/v4/source/file"           // Import the file driver here
	"github.com/jasoncheung94/url-shortener/internal/ptr"
	"github.com/jasoncheung94/url-shortener/internal/shortener/model"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func applyMigrations(db *sqlx.DB, t *testing.T) {
	// Get the absolute path of your migrations directory
	migrationsPath, err := filepath.Abs("../../database/migrations")
	fmt.Println("Path:", migrationsPath)
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}
	fmt.Println("Migrations path:", migrationsPath) // For debugging

	driver, err := postm.WithInstance(db.DB, &postm.Config{})
	if err != nil {
		t.Fatalf("could not create migrate driver: %v", err)
	}

	migrationsPath = "../../database/migrations"
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres", // Your database driver
		driver,     // The database driver instance
	)
	if err != nil {
		t.Fatalf("failed to create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("failed to run migrations: %v", err)
	}
}

func setupTestDB(t *testing.T) (*sqlx.DB, func()) {
	t.Helper()

	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:latest",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("user"),
		postgres.WithPassword("pass"),
		// postgres.BasicWaitStrategies(),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	db := sqlx.MustConnect("postgres", connStr)

	applyMigrations(db, t)

	cleanup := func() {
		db.Close()
		_ = container.Terminate(ctx)
	}

	return db, cleanup
}

func TestIntegrationPostgresRepo_SaveAndGetURL(t *testing.T) {
	t.Parallel()
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgres(db)

	url := &model.URL{
		OriginalURL: "https://example.com",
		ShortURL:    "abc123",
		CustomURL:   nil,
		CreatedAt:   time.Now(),
	}

	err := repo.SaveURL(context.Background(), url)
	require.NoError(t, err)
	require.NotZero(t, url.ID)

	result, err := repo.GetURL(context.Background(), "abc123")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, url.OriginalURL, result.OriginalURL)
	require.Equal(t, url.ShortURL, result.ShortURL)
	result, err = repo.GetURL(context.Background(), "fake-doesnt-exist")
	require.Error(t, err)
	require.Nil(t, result)
}

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
