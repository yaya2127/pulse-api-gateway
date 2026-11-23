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

type LeastConnectionsBalancer struct {
	upstreams []*Upstream
	mu        sync.RWMutex
}

func NewLeastConnectionsBalancer(upstreams []*Upstream) *LeastConnectionsBalancer {
	return &LeastConnectionsBalancer{upstreams: upstreams}
}

func (b *LeastConnectionsBalancer) Next(r *http.Request) (*Upstream, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var candidate *Upstream
	minConns := int64(^uint64(0) >> 1) // max int64

	for _, u := range b.upstreams {
		if !u.IsAlive() {
			continue
		}
		conns := u.GetActiveConns()
		if conns < minConns {
			minConns = conns
			candidate = u
		}
	}

	if candidate == nil {
		return nil, ErrNoAvailableUpstream
	}
	return candidate, nil
}

func (b *LeastConnectionsBalancer) GetUpstreams() []*Upstream {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.upstreams
}

func (b *LeastConnectionsBalancer) AddUpstream(u *Upstream) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.upstreams = append(b.upstreams, u)
}

import (
	"hash/fnv"
)

type IPHashBalancer struct {
	upstreams []*Upstream
	mu        sync.RWMutex
}

func NewIPHashBalancer(upstreams []*Upstream) *IPHashBalancer {
	return &IPHashBalancer{upstreams: upstreams}
}

func (b *IPHashBalancer) Next(r *http.Request) (*Upstream, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var alive []*Upstream
	for _, u := range b.upstreams {
		if u.IsAlive() {
			alive = append(alive, u)
		}
	}
	if len(alive) == 0 {
		return nil, ErrNoAvailableUpstream
	}

	ip := r.RemoteAddr
	hasher := fnv.New32a()
	hasher.Write([]byte(ip))
	idx := int(hasher.Sum32() % uint32(len(alive)))

	return alive[idx], nil
}

func (b *IPHashBalancer) GetUpstreams() []*Upstream {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.upstreams
}

func (b *IPHashBalancer) AddUpstream(u *Upstream) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.upstreams = append(b.upstreams, u)
}

type WeightedRoundRobinBalancer struct {
	upstreams []*Upstream
	current   int
	weight    int
	mu        sync.Mutex
}

func NewWeightedRoundRobinBalancer(upstreams []*Upstream) *WeightedRoundRobinBalancer {
	return &WeightedRoundRobinBalancer{upstreams: upstreams}
}

func (b *WeightedRoundRobinBalancer) Next(r *http.Request) (*Upstream, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	n := len(b.upstreams)
	if n == 0 {
		return nil, ErrNoAvailableUpstream
	}

	maxWeight := 0
	for _, u := range b.upstreams {
		if u.IsAlive() && u.Weight > maxWeight {
			maxWeight = u.Weight
		}
	}
	if maxWeight == 0 {
		return nil, ErrNoAvailableUpstream
	}

	for {
		b.current = (b.current + 1) % n
		if b.current == 0 {
			b.weight--
			if b.weight <= 0 {
				b.weight = maxWeight
			}
		}
		u := b.upstreams[b.current]
		if u.IsAlive() && u.Weight >= b.weight {
			return u, nil
		}
	}
}

func (b *WeightedRoundRobinBalancer) GetUpstreams() []*Upstream {
	return b.upstreams
}

func (b *WeightedRoundRobinBalancer) AddUpstream(u *Upstream) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.upstreams = append(b.upstreams, u)
}
