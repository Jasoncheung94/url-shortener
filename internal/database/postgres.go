package database

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	maxOpenConns = 20
	maxIdleConns = 5
	maxIdleTime  = 10 * time.Minute
)

// NewPostgres returns a postgres db connection using the given dsn. If empty: uses defaults.
func NewPostgres(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxIdleTime(maxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
