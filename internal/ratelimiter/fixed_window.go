package ratelimiter

import (
	"sync"
	"time"
)

// FixedWindowGlobalLimiter applies a fixed-window rate limit across the entire application.
// Every request, regardless of user/IP, shares the same counter and window.
type FixedWindowGlobalLimiter struct {
	mu         sync.Mutex    // Mutex to make the limiter safe for concurrent use
	limit      int           // Max number of requests allowed during each window
	window     time.Duration // Duration of each window (e.g., 1 second, 1 minute)
	count      int           // Current number of requests processed in the active window
	windowEnds time.Time     // Timestamp marking the end of the current window
}

// NewFixedWindowGlobalLimiter initializes a new global fixed-window limiter.
// `limit` = max requests per window. `window` = size of each time window.
func NewFixedWindowGlobalLimiter(limit int, window time.Duration) *FixedWindowGlobalLimiter {
	return &FixedWindowGlobalLimiter{
		limit:      limit,
		window:     window,
		windowEnds: time.Now().Add(window), // Start the first window from now
	}
}

// Allow returns true if a request is allowed under the current rate limit.
// If the current window has expired, it resets the counter and starts a new window.
func (l *FixedWindowGlobalLimiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	// If current time is after the window, reset the count and start a new window.
	if now.After(l.windowEnds) {
		l.count = 1
		l.windowEnds = now.Add(l.window)
		return true
	}

	// If still within window and under limit, allow the request.
	if l.count < l.limit {
		l.count++
		return true
	}

	// Otherwise, limit has been hit for this window.
	return false
}

// FixedWindowKeyedLimiter applies a fixed-window rate limit per key (e.g., per IP or user ID).
// Each key has its own window and counter, enabling independent rate limits.
type FixedWindowKeyedLimiter struct {
	mu     sync.Mutex    // Mutex for thread safety when updating the map
	limit  int           // Max number of requests per window per key
	window time.Duration // Duration of the window

	visits map[string]*visitData // Stores request count and window info per key
}

// visitData tracks the number of requests and window expiry for a single key.
type visitData struct {
	count     int       // Number of requests during the current window
	windowEnd time.Time // End time for this key's current window
}

// NewFixedWindowKeyedLimiter creates a new rate limiter that applies limits per key.
// `limit` = max requests per window per key. `window` = size of each time window.
func NewFixedWindowKeyedLimiter(limit int, window time.Duration) *FixedWindowKeyedLimiter {
	return &FixedWindowKeyedLimiter{
		limit:  limit,
		window: window,
		visits: make(map[string]*visitData),
	}
}

// Allow checks whether a request from the given key (e.g., IP address) is allowed.
// It resets the window and counter if the current window has expired.
// Also performs lazy cleanup of expired keys.
func (l *FixedWindowKeyedLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	// Lazy cleanup: remove expired keys on access
	for k, v := range l.visits {
		if now.After(v.windowEnd) {
			delete(l.visits, k)
		}
	}

	v, exists := l.visits[key]

	// If the key doesn't exist, or the window has expired, start a new window
	if !exists || now.After(v.windowEnd) {
		l.visits[key] = &visitData{
			count:     1,
			windowEnd: now.Add(l.window),
		}
		return true
	}

	// If still within the window and under the limit
	if v.count < l.limit {
		v.count++
		return true
	}

	// Rate limit hit
	return false
}
