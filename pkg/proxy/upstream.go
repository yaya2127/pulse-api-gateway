package proxy

import (
	"net/url"
	"sync"
	"sync/atomic"
)

type Upstream struct {
	URL           *url.URL
	Weight        int
	alive         atomic.Bool
	activeConns   atomic.Int64
	totalRequests atomic.Uint64
	totalErrors   atomic.Uint64
	mu            sync.RWMutex
}

func NewUpstream(targetURL string, weight int) (*Upstream, error) {
	parsed, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}
	if weight <= 0 {
		weight = 1
	}
	u := &Upstream{
		URL:    parsed,
		Weight: weight,
	}
	u.alive.Store(true)
	return u, nil
}

func (u *Upstream) IsAlive() bool {
	return u.alive.Load()
}

func (u *Upstream) SetAlive(alive bool) {
	u.alive.Store(alive)
}

func (u *Upstream) IncrementConns() {
	u.activeConns.Add(1)
	u.totalRequests.Add(1)
}

func (u *Upstream) DecrementConns() {
	u.activeConns.Add(-1)
}

func (u *Upstream) RecordError() {
	u.totalErrors.Add(1)
}

func (u *Upstream) GetActiveConns() int64 {
	return u.activeConns.Load()
}

func (u *Upstream) GetTotalRequests() uint64 {
	return u.totalRequests.Load()
}

func (u *Upstream) GetTotalErrors() uint64 {
	return u.totalErrors.Load()
}
