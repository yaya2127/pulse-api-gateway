package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTokenBucketAllowance(t *testing.T) {
	tb := NewTokenBucket(5, 10) // 5 burst, 10 tokens/sec
	for i := 0; i < 5; i++ {
		allowed, _, _ := tb.Allow()
		if !allowed {
			t.Fatalf("expected token available at iteration %d", i)
		}
	}

	allowed, _, _ := tb.Allow()
	if allowed {
		t.Fatal("expected bucket exhausted on 6th request")
	}

	time.Sleep(120 * time.Millisecond)
	allowed, _, _ = tb.Allow()
	if !allowed {
		t.Fatal("expected bucket refilled after elapsed time")
	}
}
