package shortener

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jasoncheung94/url-shortener/internal/mocks"
	"github.com/jasoncheung94/url-shortener/internal/shortener/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRoutes(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	service := mocks.NewMockService(gomock.NewController(t))
	handler := NewHandler(service) // use a mock or dummy service
	handler.Routes(mux)
}

func makeJSONRequest(method, path string, payload any) *http.Request {
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(method, path, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestShortenURL(t *testing.T) {
	t.Parallel()
	mockService := mocks.NewMockService(gomock.NewController(t))
	handler := NewHandler(mockService)
	mockService.EXPECT().SaveURL(gomock.Any(), gomock.Any()).Return("test", nil)

	input := model.URL{
		OriginalURL: "https://example.com",
	}

	req := makeJSONRequest(http.MethodPost, "/shorten", input)
	rr := httptest.NewRecorder()

	handler.ShortenURL(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)
	var res model.URL
	err := json.NewDecoder(rr.Body).Decode(&res)
	assert.NoError(t, err)
	assert.Contains(t, res.ShortURL, "http://localhost:8080/")
}

func TestRedirectURL(t *testing.T) {
	t.Parallel()
	mockService := mocks.NewMockService(gomock.NewController(t))
	handler := NewHandler(mockService)
	data := model.URL{
		ShortURL:    "1",
		OriginalURL: "http://localhost:8080/google",
	}
	// Set up the router
	mux := http.NewServeMux()
	handler.Routes(mux) // This registers the route handlers in mux.
	mockService.EXPECT().GetURL(gomock.Any(), gomock.Any()).Return(&data, nil)
	// Create a GET request to the redirect route
	req := httptest.NewRequest(http.MethodGet, "/1234", nil)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	// Now use the mux to handle the request, simulating the routing logic
	mux.ServeHTTP(rr, req)

	// Assert the status code
	assert.Equal(t, http.StatusFound, rr.Code) // Should return HTTP 302 as per RedirectURL behavior

	// Check if the Location header has the correct redirect URL
	assert.Equal(t, "http://localhost:8080/google", rr.Header().Get("Location"))
}

func TestPreviewURL(t *testing.T) {
	t.Parallel()
	mockService := mocks.NewMockService(gomock.NewController(t))
	handler := NewHandler(mockService)
	data := model.URL{
		ShortURL:    "1",
		OriginalURL: "http://localhost:8080/google",
	}
	// Set up the router
	mux := http.NewServeMux()
	handler.Routes(mux) // This registers the route handlers in mux.
	mockService.EXPECT().GetURL(gomock.Any(), gomock.Any()).Return(&data, nil)
	// Create a GET request to the redirect route
	req := httptest.NewRequest(http.MethodGet, "/preview/1234", nil)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	// Now use the mux to handle the request, simulating the routing logic
	mux.ServeHTTP(rr, req)

	// Assert the status code
	assert.Equal(t, http.StatusOK, rr.Code) // Should return HTTP 302 as per RedirectURL behavior
	var res model.URL
	err := json.NewDecoder(rr.Body).Decode(&res)
	assert.NoError(t, err)
	assert.Contains(t, res.OriginalURL, "http://localhost:8080/")
	assert.Contains(t, res.ShortURL, "1")
}
