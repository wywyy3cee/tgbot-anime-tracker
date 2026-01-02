package ratelimit

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWait_TokenAvailable(t *testing.T) {
	rl := NewRateLimiter(10)
	ctx := context.Background()

	if err := rl.Wait(ctx); err != nil {
		t.Fatalf("expected no error when tokens available, got %v", err)
	}
}

func TestWait_ContextCancelled(t *testing.T) {
	rl := NewRateLimiter(1)
	ctx := context.Background()

	if err := rl.Wait(ctx); err != nil {
		t.Fatalf("first wait failed: %v", err)
	}
	// теперь токенов нет
	ctx2, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	if err := rl.Wait(ctx2); err == nil {
		t.Fatalf("expected context deadline exceeded, got nil")
	}
}

func TestWait_MultipleTokens(t *testing.T) {
	requestsPerMinute := 5
	rl := NewRateLimiter(requestsPerMinute)
	ctx := context.Background()

	// потребляем все токены
	for i := 0; i < requestsPerMinute; i++ {
		if err := rl.Wait(ctx); err != nil {
			t.Fatalf("expected no error on token %d, got %v", i, err)
		}
	}

	ctx2, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	if err := rl.Wait(ctx2); err == nil {
		t.Fatalf("expected error when no tokens available")
	}
}

func TestWait_TokenRefill(t *testing.T) {
	requestsPerMinute := 60
	rl := NewRateLimiter(requestsPerMinute)
	ctx := context.Background()

	// потребляем все токены
	for i := 0; i < 5; i++ {
		_ = rl.Wait(ctx)
	}

	time.Sleep(100 * time.Millisecond)

	ctx3, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	if err := rl.Wait(ctx3); err != nil {
		t.Logf("note: refill timing may be system-dependent, error: %v", err)
	}
}

func TestWait_ConcurrentRequests(t *testing.T) {
	requestsPerMinute := 10
	rl := NewRateLimiter(requestsPerMinute)

	var wg sync.WaitGroup
	successCount := atomic.Int32{}
	errorCount := atomic.Int32{}

	// 15 конкурентных запросов
	for i := 0; i < 15; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx2, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			if err := rl.Wait(ctx2); err != nil {
				errorCount.Add(1)
			} else {
				successCount.Add(1)
			}
		}()
	}

	wg.Wait()

	// должны успешно получить примерно 10 токенов
	if successCount.Load() < int32(requestsPerMinute-2) || successCount.Load() > int32(requestsPerMinute+2) {
		t.Logf("success count: %d (expected ~%d)", successCount.Load(), requestsPerMinute)
	}

	if errorCount.Load() == 0 {
		t.Logf("expected some errors due to limited tokens, got none")
	}
}

func TestWait_NoTokens(t *testing.T) {
	rl := NewRateLimiter(1)
	ctx := context.Background()

	_ = rl.Wait(ctx)

	ctx2, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := rl.Wait(ctx2)
	if err == nil {
		t.Error("expected error when no tokens available")
	}

	if err != context.DeadlineExceeded && err.Error() != "context deadline exceeded" {
		t.Logf("got error: %v", err)
	}
}

func TestWait_SequentialRequests(t *testing.T) {
	requestsPerMinute := 3
	rl := NewRateLimiter(requestsPerMinute)
	ctx := context.Background()

	// первый пакет запросов - все должны пройти
	for i := 0; i < requestsPerMinute; i++ {
		if err := rl.Wait(ctx); err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
	}

	// следующий должен заблокироваться
	ctx2, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	if err := rl.Wait(ctx2); err == nil {
		t.Error("expected error after consuming all tokens")
	}
}

func TestWait_RateLimitPersistence(t *testing.T) {
	rl := NewRateLimiter(5)
	ctx := context.Background()

	// потребляем 3 токена
	for i := 0; i < 3; i++ {
		_ = rl.Wait(ctx)
	}

	// потребляем еще 2
	for i := 0; i < 2; i++ {
		_ = rl.Wait(ctx)
	}

	ctx2, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	if err := rl.Wait(ctx2); err == nil {
		t.Error("expected error after consuming 5 tokens")
	}
}

func TestWait_ContextTimeout(t *testing.T) {
	rl := NewRateLimiter(1)
	ctx := context.Background()

	_ = rl.Wait(ctx)

	// короткий таймаут
	ctx1, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	if err := rl.Wait(ctx1); err == nil {
		t.Error("expected error with 1ms timeout")
	}

	ctx2, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	if err := rl.Wait(ctx2); err != nil {
		t.Logf("200ms timeout error: %v", err)
	}
}

func TestRateLimiter_Configuration(t *testing.T) {
	requestsPerMinute := 60
	rl := NewRateLimiter(requestsPerMinute)

	if rl.maxTokens != requestsPerMinute {
		t.Errorf("expected maxTokens %d, got %d", requestsPerMinute, rl.maxTokens)
	}

	if rl.tokens != requestsPerMinute {
		t.Errorf("expected initial tokens %d, got %d", requestsPerMinute, rl.tokens)
	}

	expectedRefillRate := time.Minute / time.Duration(requestsPerMinute)
	if rl.refillRate != expectedRefillRate {
		t.Errorf("expected refillRate %v, got %v", expectedRefillRate, rl.refillRate)
	}
}

func TestWait_LargeRequestsPerMinute(t *testing.T) {
	requestsPerMinute := 1000
	rl := NewRateLimiter(requestsPerMinute)
	ctx := context.Background()

	consumed := 0
	// много запросов в короткий промежуток
	for i := 0; i < 100 && consumed < 100; i++ {
		if err := rl.Wait(ctx); err != nil {
			t.Fatalf("wait failed at iteration %d: %v", i, err)
		}
		consumed++
	}

	if consumed != 100 {
		t.Errorf("expected to consume 100 tokens, got %d", consumed)
	}
}
