package middleware

import (
	"log/slog"
	"net/http"
)

// Logger logs details about each incoming request.
func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Log incoming requests
			logger.Info("Incoming request",
				"method", r.Method,
				"path", r.URL.Path,
				"user_agent", r.UserAgent(),
			)
			next.ServeHTTP(w, r)
		})
	}
}
