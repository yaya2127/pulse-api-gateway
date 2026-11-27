package resilience

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestHedgingPolicy_FastPrimary(t *testing.T) {
	hp := NewHedgingPolicy(50 * time.Millisecond)
	ctx := context.Background()

	resp, err := hp.ExecuteHedging(ctx, func(c context.Context) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK}, nil
	})

	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK from fast primary call")
	}
}
