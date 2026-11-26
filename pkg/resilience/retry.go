package resilience

import (
	"context"
	"math/rand"
	"net/http"
	"time"
)

type RetryPolicy struct {
	MaxAttempts   int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
	RetryStatuses map[int]bool
}

func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxAttempts:   3,
		InitialDelay:  50 * time.Millisecond,
		MaxDelay:      1 * time.Second,
		BackoffFactor: 2.0,
		RetryStatuses: map[int]bool{
			http.StatusBadGateway:         true,
			http.StatusServiceUnavailable: true,
			http.StatusGatewayTimeout:     true,
		},
	}
}

func (p *RetryPolicy) Execute(ctx context.Context, fn func() (*http.Response, error)) (*http.Response, error) {
	var resp *http.Response
	var err error
	delay := p.InitialDelay

	for attempt := 0; attempt < p.MaxAttempts; attempt++ {
		resp, err = fn()
		if err == nil && !p.RetryStatuses[resp.StatusCode] {
			return resp, nil
		}

		if attempt == p.MaxAttempts-1 {
			break
		}

		jitter := time.Duration(rand.Float64() * float64(delay) * 0.2)
		sleepDur := delay + jitter
		if sleepDur > p.MaxDelay {
			sleepDur = p.MaxDelay
		}

		select {
		case <-time.After(sleepDur):
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		delay = time.Duration(float64(delay) * p.BackoffFactor)
	}

	return resp, err
}
