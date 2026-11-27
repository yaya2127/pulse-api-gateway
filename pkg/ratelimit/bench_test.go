package ratelimit

import (
	"context"
	"testing"
	"time"
)

func BenchmarkSlidingWindowLimiter_Allow(b *testing.B) {
	limiter := NewSlidingWindowLimiter(1000000, 1*time.Minute)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = limiter.Allow(ctx, "bench-client")
	}
}
