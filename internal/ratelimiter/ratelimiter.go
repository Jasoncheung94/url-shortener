package ratelimiter

import (
	"sync"
	"time"
)

// FixedWindowRateLimiter2 represents fixed window rate limiting.
type FixedWindowRateLimiter2 struct {
	sync.RWMutex
	clients map[string]int
	limit   int
	window  time.Duration
}

// NewFixedWindowLimiter2 returns instance of fixed window rate limiting.
func NewFixedWindowLimiter2(limit int, window time.Duration) *FixedWindowRateLimiter2 {
	return &FixedWindowRateLimiter2{
		clients: make(map[string]int),
		limit:   limit,
		window:  window,
	}
}

// Allow determines if IP is rate limited.
func (rl *FixedWindowRateLimiter2) Allow(ip string) (bool, time.Duration) {
	rl.RLock()
	count, exists := rl.clients[ip]
	rl.RUnlock()

	if !exists || count < rl.limit {
		rl.Lock()
		if !exists { // go routine that resets limit over and over when first added.
			go rl.resetCount(ip) // means spawning go routine every request with new IP!
		}

		rl.clients[ip]++
		rl.Unlock()
		return true, 0
	}

	return false, rl.window
}

func (rl *FixedWindowRateLimiter2) resetCount(ip string) {
	time.Sleep(rl.window)
	rl.Lock()
	delete(rl.clients, ip)
	rl.Unlock()
}
