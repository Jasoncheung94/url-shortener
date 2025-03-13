package middleware

import (
	"net/http"
	"strings"

	"github.com/jasoncheung94/url-shortener/internal/errors"
	"golang.org/x/time/rate"
)

// RateLimiter returns false if rate limited otherwise allows connection to continue.
func RateLimiter(rl *rate.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip rate limiting for swagger paths
			if strings.HasPrefix(r.URL.Path, "/swagger/") {
				next.ServeHTTP(w, r)
				return
			}
			if !rl.Allow() {
				errors.WriteJSONError(w, http.StatusTooManyRequests,
					errors.NewErrorResponse(http.StatusTooManyRequests, "Too Many Requests", ""),
				)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
