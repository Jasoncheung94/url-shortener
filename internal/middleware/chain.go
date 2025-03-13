package middleware

import "net/http"

// Chain applies multiple middleware to an http.Handler
func Chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, m := range middlewares {
		handler = m(handler)
	}
	return handler
}
