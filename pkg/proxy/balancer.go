package proxy

import (
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
)

var ErrNoAvailableUpstream = errors.New("no alive upstream available in pool")

type Balancer interface {
	Next(r *http.Request) (*Upstream, error)
	GetUpstreams() []*Upstream
	AddUpstream(u *Upstream)
}

type RoundRobinBalancer struct {
	upstreams []*Upstream
	counter   atomic.Uint64
	mu        sync.RWMutex
}

func NewRoundRobinBalancer(upstreams []*Upstream) *RoundRobinBalancer {
	return &RoundRobinBalancer{
		upstreams: upstreams,
	}
}

func (b *RoundRobinBalancer) Next(r *http.Request) (*Upstream, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	total := len(b.upstreams)
	if total == 0 {
		return nil, ErrNoAvailableUpstream
	}

	start := b.counter.Add(1)
	for i := 0; i < total; i++ {
		idx := int((start + uint64(i)) % uint64(total))
		u := b.upstreams[idx]
		if u.IsAlive() {
			return u, nil
		}
	}
	return nil, ErrNoAvailableUpstream
}

func (b *RoundRobinBalancer) GetUpstreams() []*Upstream {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.upstreams
}

func (b *RoundRobinBalancer) AddUpstream(u *Upstream) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.upstreams = append(b.upstreams, u)
}
