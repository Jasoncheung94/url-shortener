package ratelimiter

import (
	"sync"
	"time"
)

// SlidingWindowGlobalLimiter applies a sliding window rate limit globally across the application.
// Every request, regardless of user/IP, shares the same sliding window and count.
type SlidingWindowGlobalLimiter struct {
	sync.Mutex
	timestamps []time.Time   // Stores timestamps for all requests in the window
	limit      int           // Maximum number of allowed requests in the sliding window
	window     time.Duration // Duration of the sliding window (e.g., 1 minute)
}

// NewSlidingWindowGlobalLimiter creates a new global sliding window rate limiter.
// - limit: the maximum number of requests allowed within the window.
// - window: the size of the sliding time window (e.g., 1 minute, 10 seconds).
func NewSlidingWindowGlobalLimiter(limit int, window time.Duration) *SlidingWindowGlobalLimiter {
	return &SlidingWindowGlobalLimiter{
		timestamps: make([]time.Time, 0),
		limit:      limit,
		window:     window,
	}
}

// Allow checks if the request is allowed within the sliding window.
// Returns true if the request is allowed, otherwise false.
func (rl *SlidingWindowGlobalLimiter) Allow() bool {
	rl.Lock()
	defer rl.Unlock()

	now := time.Now()

	// Remove timestamps older than the sliding window
	windowStart := now.Add(-rl.window)
	rl.timestamps = filterOldRequests(rl.timestamps, windowStart)

	// Check if the request limit has been exceeded
	if len(rl.timestamps) < rl.limit {
		// Add the current request timestamp
		rl.timestamps = append(rl.timestamps, now)
		return true
	}

	// If the limit is exceeded, deny the request
	return false
}

// SlidingWindowKeyedLimiter applies a sliding window rate limit per key (e.g., per IP or user ID).
// Each key (IP) has its own sliding window and count.
type SlidingWindowKeyedLimiter struct {
	sync.Mutex
	limit   int                    // Maximum number of allowed requests per window for each key
	window  time.Duration          // Duration of the sliding window (e.g., 1 minute, 10 seconds)
	clients map[string][]time.Time // Stores request timestamps per key (e.g., IP)
}

// NewSlidingWindowKeyedLimiter creates a new sliding window limiter for specific keys (e.g., IP or user ID).
// - limit: the maximum number of requests allowed within the window for each key.
// - window: the size of the sliding time window (e.g., 1 minute, 10 seconds).
func NewSlidingWindowKeyedLimiter(limit int, window time.Duration) *SlidingWindowKeyedLimiter {
	return &SlidingWindowKeyedLimiter{
		limit:   limit,
		window:  window,
		clients: make(map[string][]time.Time),
	}
}

// Allow checks if the request from the specified key (e.g., IP) is allowed.
// Returns true if the request is allowed, otherwise false.
func (rl *SlidingWindowKeyedLimiter) Allow(key string) bool {
	rl.Lock()
	defer rl.Unlock()

	now := time.Now()

	// Get the request timestamps for the key (IP)
	timestamps, exists := rl.clients[key]
	if !exists {
		timestamps = []time.Time{}
	}

	// Remove timestamps older than the sliding window
	windowStart := now.Add(-rl.window)
	timestamps = filterOldRequests(timestamps, windowStart)

	// Check if the request limit for the key has been exceeded
	if len(timestamps) < rl.limit {
		// Add the current request timestamp
		timestamps = append(timestamps, now)
		rl.clients[key] = timestamps
		return true
	}

	// If the limit has been exceeded, deny the request
	return false
}

// filterOldRequests filters out timestamps that are older than the provided time threshold.
func filterOldRequests(timestamps []time.Time, threshold time.Time) []time.Time {
	var filtered []time.Time
	for _, ts := range timestamps {
		if ts.After(threshold) {
			filtered = append(filtered, ts)
		}
	}
	return filtered
}
