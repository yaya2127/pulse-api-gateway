package resilience

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestRetryPolicy_SuccessAfterFailures(t *testing.T) {
	policy := DefaultRetryPolicy()
	policy.InitialDelay = 1 * time.Millisecond

	attempts := 0
	ctx := context.Background()

	resp, err := policy.Execute(ctx, func() (*http.Response, error) {
		attempts++
		if attempts < 3 {
			return &http.Response{StatusCode: http.StatusBadGateway}, nil
		}
		return &http.Response{StatusCode: http.StatusOK}, nil
	})

	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected successful response on attempt 3")
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetryPolicy_ContextCancellation(t *testing.T) {
	policy := DefaultRetryPolicy()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := policy.Execute(ctx, func() (*http.Response, error) {
		return nil, errors.New("connection failed")
	})

	if err == nil {
		t.Fatal("expected error due to canceled context")
	}
}
