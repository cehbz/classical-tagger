// internal/uploader/rate_limiter.go
package uploader

import (
	"context"
	"sync"
	"time"
)

// RateLimiter implements a leaky bucket rate limiter
// Based on the existing discogs rate limiter pattern
type RateLimiter struct {
	capacity   int           // max tokens in bucket
	refillRate time.Duration // time between token refills
	tokens     int           // current tokens
	lastRefill time.Time     // last refill timestamp
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(capacity int, refillRate time.Duration) *RateLimiter {
	return &RateLimiter{
		capacity:   capacity,
		refillRate: refillRate / time.Duration(capacity), // Per-token refill time
		tokens:     capacity,
		lastRefill: time.Now(),
	}
}

// Wait blocks until a token is available
func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		rl.mu.Lock()
		
		// Refill tokens based on elapsed time
		now := time.Now()
		elapsed := now.Sub(rl.lastRefill)
		tokensToAdd := int(elapsed / rl.refillRate)
		
		if tokensToAdd > 0 {
			rl.tokens += tokensToAdd
			if rl.tokens > rl.capacity {
				rl.tokens = rl.capacity
			}
			rl.lastRefill = rl.lastRefill.Add(time.Duration(tokensToAdd) * rl.refillRate)
		}
		
		// Check if we have a token available
		if rl.tokens > 0 {
			rl.tokens--
			rl.mu.Unlock()
			return nil
		}
		
		// Calculate wait time until next token
		waitTime := rl.refillRate - now.Sub(rl.lastRefill)
		rl.mu.Unlock()
		
		// Wait with context cancellation support
		select {
		case <-time.After(waitTime):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// OnResponse updates the timestamp for rate calculations
func (rl *RateLimiter) OnResponse() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	// Update lastRefill based on when we receive the response
	// This ensures rate limiting is based on actual response times
	rl.lastRefill = time.Now()
}