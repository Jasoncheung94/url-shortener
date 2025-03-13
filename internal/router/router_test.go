package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jasoncheung94/url-shortener/internal/mocks"
	"github.com/jasoncheung94/url-shortener/internal/shortener"
	"github.com/stretchr/testify/require"
)

func TestRouter(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := shortener.NewHandler(mockService)
	router := New(handler)

	// Example test: GET /health should return 200 OK
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Test: GET /panic (if this route is defined and meant to panic/test recovery)
	req = httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Adjust expectations depending on how /panic behaves.
	require.Equal(t, http.StatusInternalServerError, rec.Code)      // or whatever it should return
	require.Contains(t, rec.Body.String(), "Internal Server Error") // or actual response
}
