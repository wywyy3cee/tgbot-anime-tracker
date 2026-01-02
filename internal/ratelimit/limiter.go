package ratelimit

import (
	"context"
	"sync"
	"time"
)

type RateLimiter struct {
	tokens     int
	maxTokens  int
	refillRate time.Duration
	mu         sync.Mutex
}

func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		maxTokens:  requestsPerMinute,
		tokens:     requestsPerMinute,
		refillRate: time.Minute / time.Duration(requestsPerMinute),
	}

	go rl.refill()
	return rl
}

func (rl *RateLimiter) refill() {
	ticker := time.NewTicker(rl.refillRate)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		if rl.tokens < rl.maxTokens {
			rl.tokens++
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		rl.mu.Lock()
		if rl.tokens > 0 {
			rl.tokens--
			rl.mu.Unlock()
			return nil
		}
		rl.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}
}
