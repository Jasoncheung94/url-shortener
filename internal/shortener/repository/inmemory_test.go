package repository

import (
	"context"
	"log/slog"
	"os"
	"testing"

	l "github.com/jasoncheung94/url-shortener/internal/logger"
	"github.com/jasoncheung94/url-shortener/internal/shortener/model"
	"github.com/stretchr/testify/assert"
)

func TestSaveURL(t *testing.T) {
	t.Parallel()
	l.SetLogger(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	// common.SetLogger(slog.New(slog.NewTextHandler(io.Discard, nil)))

	repo := NewInMemory()
	ctx := context.Background()
	data := model.URL{
		OriginalURL: "https://loooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo.ng/",
		ShortURL:    "Lu8S545",
	}
	err := repo.SaveURL(ctx, &data)

	assert.NoError(t, err)

	url, err := repo.GetURL(ctx, data.ShortURL)
	assert.NoError(t, err)
	assert.Equal(t, data.OriginalURL, url.OriginalURL)
}
