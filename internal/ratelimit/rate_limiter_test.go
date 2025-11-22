// internal/ratelimit/rate_limiter_test.go
package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestRateLimiter_BasicOperation(t *testing.T) {
	limiter := NewRateLimiter(2, 2*time.Second) // 2 tokens, refill every 2 seconds
	
	ctx := context.Background()
	
	// Should allow first two requests immediately
	for i := range 2 {
		start := time.Now()
		if err := limiter.Wait(ctx); err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
		elapsed := time.Since(start)
		if elapsed > 100*time.Millisecond {
			t.Errorf("request %d took too long: %v", i, elapsed)
		}
	}
	
	// Third request should wait
	start := time.Now()
	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("third request failed: %v", err)
	}
	elapsed := time.Since(start)
	if elapsed < 900*time.Millisecond { // Allow some timing variance
		t.Errorf("third request didn't wait long enough: %v", elapsed)
	}
}

func TestRateLimiter_ContextCancellation(t *testing.T) {
	limiter := NewRateLimiter(1, 10*time.Second) // Very slow refill
	
	// Use up the token
	ctx := context.Background()
	limiter.Wait(ctx)
	
	// Create cancellable context
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	
	// Should return context error
	err := limiter.Wait(ctx)
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

// Add this test to internal/ratelimit/rate_limiter_test.go

func TestRateLimiter_OnResponse(t *testing.T) {
	// Test that OnResponse properly resets the refill timer
	limiter := NewRateLimiter(2, 2*time.Second) // 2 tokens per 2 seconds
	ctx := context.Background()
	
	// Use both tokens
	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("first request failed: %v", err)
	}
	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("second request failed: %v", err)
	}
	
	// Without OnResponse, we'd need to wait 1 second for next token
	// But let's simulate a slow response by waiting 800ms
	time.Sleep(800 * time.Millisecond)
	
	// Call OnResponse to reset the timer
	limiter.OnResponse()
	
	// Now the next token should require a full second wait from NOW,
	// not from the original request time
	start := time.Now()
	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("third request failed: %v", err)
	}
	elapsed := time.Since(start)
	
	// Should wait approximately 1 second (half of 2-second interval for 2 tokens)
	// Allow some variance for timing
	if elapsed < 900*time.Millisecond || elapsed > 1100*time.Millisecond {
		t.Errorf("expected wait of ~1 second after OnResponse, got %v", elapsed)
	}
}

func TestRateLimiter_OnResponseResetsBurst(t *testing.T) {
	// Test that OnResponse affects burst recovery
	limiter := NewRateLimiter(3, 3*time.Second) // 3 tokens per 3 seconds
	ctx := context.Background()
	
	// Use all tokens rapidly
	for i := 0; i < 3; i++ {
		if err := limiter.Wait(ctx); err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
	}
	
	// Wait 500ms (not enough for a token normally)
	time.Sleep(500 * time.Millisecond)
	
	// OnResponse should reset timing
	limiter.OnResponse()
	
	// Wait another 1 second (total would be 1.5s without OnResponse)
	time.Sleep(1 * time.Second)
	
	// With OnResponse, only 1 second has passed since reset
	// So we should have exactly 1 token available
	
	// This should succeed immediately (using the 1 token)
	start := time.Now()
	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("request after OnResponse failed: %v", err)
	}
	if time.Since(start) > 100*time.Millisecond {
		t.Error("first request after OnResponse should be immediate")
	}
	
	// This should wait for next token
	start = time.Now()
	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("second request after OnResponse failed: %v", err)
	}
	elapsed := time.Since(start)
	if elapsed < 900*time.Millisecond {
		t.Errorf("second request should wait ~1 second, got %v", elapsed)
	}
}

func TestRateLimiter_MultipleOnResponseCalls(t *testing.T) {
	// Test that multiple OnResponse calls work correctly
	limiter := NewRateLimiter(2, 2*time.Second)
	ctx := context.Background()
	
	// Use one token
	limiter.Wait(ctx)
	
	// Call OnResponse multiple times quickly
	limiter.OnResponse()
	time.Sleep(100 * time.Millisecond)
	limiter.OnResponse()
	time.Sleep(100 * time.Millisecond)
	limiter.OnResponse()
	
	// The last OnResponse should be what counts
	// Wait should be based on the most recent OnResponse
	start := time.Now()
	limiter.Wait(ctx) // Use second token
	limiter.Wait(ctx) // Should wait for refill
	elapsed := time.Since(start)
	
	// Should wait approximately 1 second from last OnResponse
	if elapsed < 900*time.Millisecond || elapsed > 1100*time.Millisecond {
		t.Errorf("expected wait of ~1 second from last OnResponse, got %v", elapsed)
	}
}