package ratelimiter

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Note: In memory solutions for rate limiting! Ideally, use store like Redis.

// TokenBucket is a basic implementation of the token bucket rate limiter algorithm.
type TokenBucket struct {
	capacity   int        // The maximum number of tokens in the bucket (burst size).
	tokens     int        // Current number of tokens in the bucket.
	rate       int        // Tokens added per second.
	lastRefill time.Time  // Timestamp of the last refill.
	mu         sync.Mutex // Mutex to protect concurrent access.
}

// NewTokenBucket creates a new token bucket limiter with the specified rate and capacity.
// rate: number of tokens to add per second
// capacity: max burst (bucket size)
func NewTokenBucket(rate int, capacity int) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity, // Start full
		rate:       rate,
		lastRefill: time.Now(),
	}
}

// refill adds tokens based on elapsed time and the refill rate.
// It refills the bucket only when called and ensures tokens don't exceed the capacity.
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.lastRefill = now

	// Calculate how many tokens to add based on elapsed time and rate
	newTokens := int(elapsed * float64(tb.rate))
	if newTokens > 0 {
		tb.tokens = min(tb.capacity, tb.tokens+newTokens)
	}
}

// Allow returns true if a request is allowed (token available), false otherwise.
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill the bucket before checking the token count.
	tb.refill()

	// If there are tokens available, consume one and allow the request
	if tb.tokens > 0 {
		tb.tokens-- // Consume a token
		return true
	}

	// If no tokens available, deny the request
	return false
}

// IPRateLimiter represents the methods and data for IPRateLimiting.
type IPRateLimiter struct {
	rate     int
	capacity int
	limiters map[string]*TokenBucket
	mu       sync.Mutex
	cleanup  time.Duration
	lastSeen map[string]time.Time
}

// NewIPRateLimiter initializes a new IPRateLimiter.
func NewIPRateLimiter(rate int, capacity int, cleanupInterval time.Duration) *IPRateLimiter {
	ipLimiter := &IPRateLimiter{
		rate:     rate,
		capacity: capacity,
		limiters: make(map[string]*TokenBucket),
		lastSeen: make(map[string]time.Time),
		cleanup:  cleanupInterval,
	}

	// Optional: clean up stale IPs periodically
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		for range ticker.C {
			ipLimiter.cleanupOldEntries()
		}
	}()

	return ipLimiter
}

// GetLimiter returns a token bucket for the given IP or creates one.
func (l *IPRateLimiter) GetLimiter(ip string) *TokenBucket {
	l.mu.Lock()
	defer l.mu.Unlock()

	if limiter, exists := l.limiters[ip]; exists {
		l.lastSeen[ip] = time.Now()
		return limiter
	}

	tb := NewTokenBucket(l.rate, l.capacity)
	l.limiters[ip] = tb
	l.lastSeen[ip] = time.Now()
	return tb
}

// cleanupOldEntries removes entries that haven’t been used recently
func (l *IPRateLimiter) cleanupOldEntries() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	for ip, lastSeen := range l.lastSeen {
		if now.Sub(lastSeen) > l.cleanup {
			delete(l.limiters, ip)
			delete(l.lastSeen, ip)
		}
	}
}

// Middleware applies per-IP rate limiting to incoming requests.
func (l *IPRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		limiter := l.GetLimiter(ip)

		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getIP tries to extract a user IP, accounting for proxies.
func getIP(r *http.Request) string {
	// Check X-Forwarded-For (for reverse proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
