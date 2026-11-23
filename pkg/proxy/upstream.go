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
