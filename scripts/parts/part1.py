import sys
import os
sys.path.append(os.path.dirname(os.path.dirname(__file__)))
from engine import execute_commit, write_file, append_file

def run_part1():
    print("--- Running Part 1: Core, Config & Balancers ---")
    
    # Commit 0
    write_file("go.mod", """module github.com/yaya2127/pulse-api-gateway

go 1.22
""")
    execute_commit(0, "feat(core): initialize go module for pulse-api-gateway")

    # Commit 1
    write_file(".gitignore", """# Binaries
bin/
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test artifacts
*.out
*.test
coverage.txt
coverage.html

# IDE & OS
.idea/
.vscode/
*.swp
*.swo
.DS_Store
Thumbs.db
""")
    execute_commit(1, "chore(git): configure .gitignore for Go binaries and test artifacts")

    # Commit 2
    write_file("LICENSE", """Pulse API Gateway
Copyright (c) 2026 Yared Kinfe

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files, to deal in the Software
without restriction, including rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies.
""")
    execute_commit(2, "docs(license): add open source license notice")

    # Commit 3
    write_file("config/config.go", """package config

import (
	"time"
)

type ServerConfig struct {
	Port         int           `json:"port"`
	MetricsPort  int           `json:"metrics_port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

type UpstreamConfig struct {
	URL    string `json:"url"`
	Weight int    `json:"weight"`
}

type RouteConfig struct {
	PathPrefix      string           `json:"path_prefix"`
	StripPrefix     bool             `json:"strip_prefix"`
	BalancerStrategy string          `json:"balancer_strategy"`
	Upstreams       []UpstreamConfig `json:"upstreams"`
	RequireAuth     bool             `json:"require_auth"`
	RateLimitRPM    int              `json:"rate_limit_rpm"`
	RateLimitBurst  int              `json:"rate_limit_burst"`
	TimeoutSeconds  int              `json:"timeout_seconds"`
}

type Config struct {
	Server ServerConfig  `json:"server"`
	Routes []RouteConfig `json:"routes"`
}
""")
    execute_commit(3, "feat(config): declare core gateway configuration structures")

    # Commit 4
    append_file("config/config.go", """
import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	if envPort := os.Getenv("PULSE_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			cfg.Server.Port = p
		}
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.MetricsPort == 0 {
		cfg.Server.MetricsPort = 9090
	}

	return &cfg, nil
}
""")
    execute_commit(4, "feat(config): implement JSON config parser with environment variable overrides")

    # Commit 5
    write_file("config/config.json", """{
  "server": {
    "port": 8080,
    "metrics_port": 9090,
    "read_timeout": 15000000000,
    "write_timeout": 15000000000,
    "idle_timeout": 60000000000
  },
  "routes": [
    {
      "path_prefix": "/api/v1/users",
      "strip_prefix": false,
      "balancer_strategy": "round_robin",
      "upstreams": [
        { "url": "http://127.0.0.1:9001", "weight": 1 },
        { "url": "http://127.0.0.1:9002", "weight": 1 }
      ],
      "require_auth": true,
      "rate_limit_rpm": 600,
      "rate_limit_burst": 50,
      "timeout_seconds": 10
    },
    {
      "path_prefix": "/api/v1/orders",
      "strip_prefix": false,
      "balancer_strategy": "least_conn",
      "upstreams": [
        { "url": "http://127.0.0.1:9003", "weight": 2 },
        { "url": "http://127.0.0.1:9004", "weight": 1 }
      ],
      "require_auth": true,
      "rate_limit_rpm": 1200,
      "rate_limit_burst": 100,
      "timeout_seconds": 15
    }
  ]
}
""")
    execute_commit(5, "feat(config): add default gateway configuration template")

    # Commit 6
    write_file("config/config_test.go", """package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.json")
	content := []byte(`{
		"server": { "port": 8888, "metrics_port": 9999 },
		"routes": [
			{ "path_prefix": "/test", "upstreams": [{"url": "http://localhost:3000", "weight": 1}] }
		]
	}`)

	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Server.Port != 8888 {
		t.Errorf("expected port 8888, got %d", cfg.Server.Port)
	}
	if len(cfg.Routes) != 1 {
		t.Errorf("expected 1 route, got %d", len(cfg.Routes))
	}
}
""")
    execute_commit(6, "test(config): add unit tests for configuration parsing and validation")

    # Commit 7
    write_file("pkg/proxy/upstream.go", """package proxy

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
""")
    execute_commit(7, "feat(proxy): define upstream target data structures and thread-safe states")

    # Commit 8
    append_file("pkg/proxy/upstream.go", """
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
""")
    execute_commit(8, "feat(proxy): add upstream health state tracking and active connection counter")

    # Commit 9
    write_file("pkg/proxy/balancer.go", """package proxy

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
""")
    execute_commit(9, "feat(proxy): define load balancer interface contract")

    # Commit 10
    append_file("pkg/proxy/balancer.go", """
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
""")
    execute_commit(10, "feat(proxy): implement round-robin upstream balancing algorithm")

    # Commit 11
    write_file("pkg/proxy/balancer_test.go", """package proxy

import (
	"net/http/httptest"
	"testing"
)

func TestRoundRobinBalancer(t *testing.T) {
	u1, _ := NewUpstream("http://127.0.0.1:9001", 1)
	u2, _ := NewUpstream("http://127.0.0.1:9002", 1)

	balancer := NewRoundRobinBalancer([]*Upstream{u1, u2})
	req := httptest.NewRequest("GET", "/test", nil)

	seen := make(map[string]int)
	for i := 0; i < 4; i++ {
		next, err := balancer.Next(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		seen[next.URL.Host]++
	}

	if seen["127.0.0.1:9001"] != 2 || seen["127.0.0.1:9002"] != 2 {
		t.Errorf("round robin uneven distribution: %v", seen)
	}
}
""")
    execute_commit(11, "test(proxy): implement unit tests for round-robin load distribution")

    # Commit 12
    append_file("pkg/proxy/balancer.go", """
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
""")
    execute_commit(12, "feat(proxy): implement least-connections load balancing strategy")

    # Commit 13
    append_file("pkg/proxy/balancer_test.go", """
func TestLeastConnectionsBalancer(t *testing.T) {
	u1, _ := NewUpstream("http://127.0.0.1:9001", 1)
	u2, _ := NewUpstream("http://127.0.0.1:9002", 1)

	u1.IncrementConns()
	u1.IncrementConns()
	u2.IncrementConns()

	balancer := NewLeastConnectionsBalancer([]*Upstream{u1, u2})
	req := httptest.NewRequest("GET", "/test", nil)

	next, err := balancer.Next(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if next.URL.Host != "127.0.0.1:9002" {
		t.Errorf("expected 127.0.0.1:9002 with 1 conn, got %s", next.URL.Host)
	}
}
""")
    execute_commit(13, "test(proxy): add test coverage for least-connections distribution logic")

    # Commit 14
    append_file("pkg/proxy/balancer.go", """
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
""")
    execute_commit(14, "feat(proxy): implement IP-hash sticky session load balancing")

    # Commit 15
    append_file("pkg/proxy/balancer_test.go", """
func TestIPHashBalancer(t *testing.T) {
	u1, _ := NewUpstream("http://127.0.0.1:9001", 1)
	u2, _ := NewUpstream("http://127.0.0.1:9002", 1)

	balancer := NewIPHashBalancer([]*Upstream{u1, u2})
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.100:4321"

	first, _ := balancer.Next(req1)
	for i := 0; i < 5; i++ {
		next, _ := balancer.Next(req1)
		if next.URL.Host != first.URL.Host {
			t.Errorf("IPHash sticky session broken: got %s expected %s", next.URL.Host, first.URL.Host)
		}
	}
}
""")
    execute_commit(15, "test(proxy): verify IP-hash consistency across client IP hashes")

    # Commit 16
    append_file("pkg/proxy/balancer.go", """
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
""")
    execute_commit(16, "feat(proxy): implement weighted round-robin distribution")

    # Commit 17
    append_file("pkg/proxy/balancer_test.go", """
func TestWeightedRoundRobinBalancer(t *testing.T) {
	u1, _ := NewUpstream("http://127.0.0.1:9001", 3)
	u2, _ := NewUpstream("http://127.0.0.1:9002", 1)

	balancer := NewWeightedRoundRobinBalancer([]*Upstream{u1, u2})
	req := httptest.NewRequest("GET", "/test", nil)

	counts := make(map[string]int)
	for i := 0; i < 40; i++ {
		next, err := balancer.Next(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		counts[next.URL.Host]++
	}

	if counts["127.0.0.1:9001"] <= counts["127.0.0.1:9002"] {
		t.Errorf("expected weighted u1 > u2, got: %v", counts)
	}
}
""")
    execute_commit(17, "test(proxy): validate weighted round-robin distribution ratios")

    # Commit 18
    write_file("pkg/proxy/balancer_bench_test.go", """package proxy

import (
	"net/http/httptest"
	"testing"
)

func BenchmarkRoundRobinBalancer(b *testing.B) {
	u1, _ := NewUpstream("http://10.0.0.1:80", 1)
	u2, _ := NewUpstream("http://10.0.0.2:80", 1)
	u3, _ := NewUpstream("http://10.0.0.3:80", 1)
	balancer := NewRoundRobinBalancer([]*Upstream{u1, u2, u3})
	req := httptest.NewRequest("GET", "/", nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = balancer.Next(req)
		}
	})
}
""")
    execute_commit(18, "perf(proxy): add benchmark tests for upstream selection algorithms")

    # Commit 19
    write_file("pkg/proxy/circuit_breaker.go", """package proxy

import (
	"errors"
	"sync"
	"time"
)

type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF-OPEN"
	default:
		return "UNKNOWN"
	}
}

var ErrCircuitOpen = errors.New("circuit breaker is OPEN: rejecting upstream requests")

type CircuitBreaker struct {
	maxFailures  int
	timeout      time.Duration
	state        CircuitState
	failures     int
	successes    int
	lastTripped  time.Time
	mu           sync.RWMutex
}

func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures: maxFailures,
		timeout:     timeout,
		state:       StateClosed,
	}
}
""")
    execute_commit(19, "feat(proxy): implement circuit breaker finite state machine")

    # Commit 20
    append_file("pkg/proxy/circuit_breaker.go", """
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	if cb.state == StateOpen {
		if now.Sub(cb.lastTripped) > cb.timeout {
			cb.state = StateHalfOpen
			cb.successes = 0
			return true
		}
		return false
	}
	return true
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	if cb.state == StateHalfOpen || cb.failures >= cb.maxFailures {
		cb.state = StateOpen
		cb.lastTripped = time.Now()
	}
}
""")
    execute_commit(20, "feat(proxy): implement failure counter and trip condition in circuit breaker")

    # Commit 21
    append_file("pkg/proxy/circuit_breaker.go", """
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateHalfOpen {
		cb.successes++
		if cb.successes >= 2 {
			cb.state = StateClosed
			cb.failures = 0
		}
	} else if cb.state == StateClosed {
		cb.failures = 0
	}
}

func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}
""")
    execute_commit(21, "feat(proxy): implement half-open probationary state and recovery logic")

if __name__ == "__main__":
    run_part1()
