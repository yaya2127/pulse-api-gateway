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

func TestRateLimitMiddleware429(t *testing.T) {
	limiter := NewRateLimiter(2, 1)
	handler := limiter.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	}

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 Too Many Requests, got %d", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Errorf("missing Retry-After header on 429 response")
	}
}
