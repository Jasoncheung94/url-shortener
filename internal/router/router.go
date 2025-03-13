package router

import (
	"encoding/json"
	"net/http"

	l "github.com/jasoncheung94/url-shortener/internal/logger"
	"github.com/jasoncheung94/url-shortener/internal/middleware"
	"github.com/jasoncheung94/url-shortener/internal/shortener"
	httpSwagger "github.com/swaggo/http-swagger"
	"golang.org/x/time/rate"
)

// New return the http handler with routes init and middleware.
func New(handler *shortener.Handler) http.Handler {
	mux := http.NewServeMux()
	// Setup handler routes.
	handler.Routes(mux)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		l.Logger.Info("Health check request received", "method", r.Method, "path", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "OK"})
	})
	mux.HandleFunc("GET /panic", func(_ http.ResponseWriter, _ *http.Request) {
		panic("something went wrong!") // This simulates a panic
	})

	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// middleware chain.
	rateLimiter := rate.NewLimiter(2, 5) // Small rate limit for my app! :D
	middlewareMux := middleware.Chain(mux,
		middleware.Logger(l.Logger), // Logs every request
		middleware.Recovery,
		middleware.RateLimiter(rateLimiter),
	)

	return middlewareMux
}
