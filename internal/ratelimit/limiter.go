package ratelimit

import (
	"context"
	"sync"
	"time"
)

type RateLimiter struct {
	tokens     int // current available tokens
	maxTokens  int
	refillRate time.Duration // how often to add a token
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
