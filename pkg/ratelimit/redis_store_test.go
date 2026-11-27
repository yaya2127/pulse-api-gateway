package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestSlidingWindowLimiter(t *testing.T) {
	limiter := NewSlidingWindowLimiter(5, 50*time.Millisecond)
	ctx := context.Background()
	clientID := "client-api-test"

	for i := 0; i < 5; i++ {
		allowed, err := limiter.Allow(ctx, clientID)
		if err != nil || !allowed {
			t.Fatalf("expected request %d to be allowed", i)
		}
	}

	// 6th request within window must be rejected
	allowed, err := limiter.Allow(ctx, clientID)
	if err != nil || allowed {
		t.Fatal("expected 6th request to be rejected")
	}

	// After window expiry, request must be allowed again
	time.Sleep(60 * time.Millisecond)
	allowed, err = limiter.Allow(ctx, clientID)
	if err != nil || !allowed {
		t.Fatal("expected request to be allowed after sliding window expiration")
	}
}
