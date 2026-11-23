package middleware

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type TokenBucket struct {
	capacity   float64
	tokens     float64
	refillRate float64
	lastRefill time.Time
	mu         sync.Mutex
}

func NewTokenBucket(capacity float64, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) Allow() (bool, int, time.Duration) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.lastRefill = now

	tb.tokens = math.Min(tb.capacity, tb.tokens+elapsed*tb.refillRate)

	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		remaining := int(tb.tokens)
		return true, remaining, 0
	}

	missing := 1.0 - tb.tokens
	retryAfter := time.Duration((missing / tb.refillRate) * float64(time.Second))
	return false, 0, retryAfter
}

type RateLimiter struct {
	buckets    sync.Map
	capacity   float64
	refillRate float64
}

func NewRateLimiter(capacity float64, refillRate float64) *RateLimiter {
	return &RateLimiter{
		capacity:   capacity,
		refillRate: refillRate,
	}
}

func (rl *RateLimiter) getBucket(key string) *TokenBucket {
	val, loaded := rl.buckets.Load(key)
	if loaded {
		return val.(*TokenBucket)
	}
	newBucket := NewTokenBucket(rl.capacity, rl.refillRate)
	actual, _ := rl.buckets.LoadOrStore(key, newBucket)
	return actual.(*TokenBucket)
}
