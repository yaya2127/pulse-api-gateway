package resilience

import (
	"context"
	"net/http"
	"time"
)

type HedgingPolicy struct {
	DelayThreshold time.Duration
}

func NewHedgingPolicy(delay time.Duration) *HedgingPolicy {
	return &HedgingPolicy{DelayThreshold: delay}
}

func (h *HedgingPolicy) ExecuteHedging(ctx context.Context, fn func(context.Context) (*http.Response, error)) (*http.Response, error) {
	ctx1, cancel1 := context.WithCancel(ctx)
	defer cancel1()

	type result struct {
		resp *http.Response
		err  error
	}

	resCh := make(chan result, 2)
	go func() {
		resp, err := fn(ctx1)
		resCh <- result{resp: resp, err: err}
	}()

	timer := time.NewTimer(h.DelayThreshold)
	defer timer.Stop()

	select {
	case res := <-resCh:
		return res.resp, res.err
	case <-timer.C:
		// Dispatch speculative hedge
		ctx2, cancel2 := context.WithCancel(ctx)
		defer cancel2()

		go func() {
			resp, err := fn(ctx2)
			resCh <- result{resp: resp, err: err}
		}()

		res := <-resCh
		return res.resp, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
