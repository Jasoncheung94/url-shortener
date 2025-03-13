package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Simulated handler that panics
func panicHandler(_ http.ResponseWriter, _ *http.Request) {
	panic("test panic")
}

func TestRecoveryMiddleware(t *testing.T) {
	t.Parallel()
	// Wrap the panic-inducing handler with the recovery middleware - called immediately
	handler := Recovery(http.HandlerFunc(panicHandler))

	// Create a test request
	req, err := http.NewRequest("GET", "/panic", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a test response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	handler.ServeHTTP(rr, req)

	// Check if response is 500 Internal Server Error
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Expected status code 500, got %d", status)
	}

	// Check response body
	expectedBody := "Internal Server Error\n"
	if rr.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, rr.Body.String())
	}
}
