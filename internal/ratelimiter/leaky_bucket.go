package ratelimiter

import (
	"sync"
	"time"
)

// LeakyBucketGlobalLimiter applies the leaky bucket rate limit across the entire application.
type LeakyBucketGlobalLimiter struct {
	sync.Mutex
	water      float64   // Current water level in the bucket
	lastLeakTs time.Time // Timestamp of last leak calculation
	rate       float64   // Leak rate in requests per second
	cap        float64   // Max capacity of the bucket
}

// NewLeakyBucketGlobalLimiter creates a new global leaky bucket limiter.
// - ratePerSec: how many requests are allowed to leak per second (i.e., the processing rate).
// - capacity: how many requests can be held in the bucket before it overflows.
func NewLeakyBucketGlobalLimiter(ratePerSec float64, capacity float64) *LeakyBucketGlobalLimiter {
	return &LeakyBucketGlobalLimiter{
		rate:       ratePerSec,
		cap:        capacity,
		lastLeakTs: time.Now(),
	}
}

// Allow checks if a global request is allowed.
func (lb *LeakyBucketGlobalLimiter) Allow() bool {
	lb.Lock()
	defer lb.Unlock()

	now := time.Now()
	elapsed := now.Sub(lb.lastLeakTs).Seconds()

	// Calculate the amount of water to leak
	leaked := elapsed * lb.rate
	lb.water = max(0, lb.water-leaked) // Leak the water, never below 0 eg 0, 0-3=0

	// Update the timestamp of last leak
	lb.lastLeakTs = now

	// Check if there's space in the bucket
	if lb.water < lb.cap {
		lb.water++ // Add the request to the bucket
		return true
	}

	// If bucket is full, deny the request
	return false
}

// LeakyBucketKeyedLimiter applies the leaky bucket algorithm on a per-IP basis.
// It allows for rate limiting on individual IPs.
type LeakyBucketKeyedLimiter struct {
	sync.Mutex
	clients map[string]*leakyBucketState // Stores rate-limiting state per client
	rate    float64                      // Leak rate in requests per second
	cap     float64                      // Max capacity of the bucket
}

// leakyBucketState holds the current "water" level (i.e., how many requests are in the bucket)
// and the last timestamp when the bucket was leaked.
type leakyBucketState struct {
	water      float64   // Current water level in the bucket
	lastLeakTs time.Time // Timestamp of last leak calculation
}

// NewLeakyBucketKeyedLimiter creates a new leaky bucket limiter that is specific to keys (IPs or users).
// - ratePerSec: the number of requests allowed to leak per second (processing rate).
// - capacity: the number of requests that can be held in the bucket before it overflows.
func NewLeakyBucketKeyedLimiter(ratePerSec float64, capacity float64) *LeakyBucketKeyedLimiter {
	return &LeakyBucketKeyedLimiter{
		clients: make(map[string]*leakyBucketState),
		rate:    ratePerSec,
		cap:     capacity,
	}
}

// Allow checks if the request from the specified key (e.g., IP) is allowed.
func (lb *LeakyBucketKeyedLimiter) Allow(ip string) bool {
	lb.Lock()
	defer lb.Unlock()

	now := time.Now()

	// Get current client state or initialize it
	state, exists := lb.clients[ip]
	if !exists {
		state = &leakyBucketState{
			water:      0,
			lastLeakTs: now,
		}
		lb.clients[ip] = state
	}

	// Time since last leak
	elapsed := now.Sub(state.lastLeakTs).Seconds()

	// Calculate how much water to leak (rate * seconds passed)
	leaked := elapsed * lb.rate

	// Reduce water level by leaked amount, but never below 0
	state.water = max(0, state.water-leaked)

	// Update timestamp
	state.lastLeakTs = now

	// Check if there's space in the bucket
	if state.water < lb.cap {
		state.water++ // Add this request to the bucket
		return true
	}

	// Bucket full – reject the request
	return false
}
