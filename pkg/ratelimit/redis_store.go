package ratelimit

import (
	"context"
	"errors"
	"sync"
	"time"
)

type SlidingWindowLimiter struct {
	mu           sync.Mutex
	rateLimit    int
	windowSize   time.Duration
	clientTimestamps map[string][]time.Time
}

func NewSlidingWindowLimiter(rateLimit int, windowSize time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		rateLimit:        rateLimit,
		windowSize:       windowSize,
		clientTimestamps: make(map[string][]time.Time),
	}
}

func (l *SlidingWindowLimiter) Allow(ctx context.Context, clientID string) (bool, error) {
	if clientID == "" {
		return false, errors.New("client ID cannot be empty")
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.windowSize)

	timestamps := l.clientTimestamps[clientID]
	var active []time.Time
	for _, ts := range timestamps {
		if ts.After(cutoff) {
			active = append(active, ts)
		}
	}

	if len(active) >= l.rateLimit {
		l.clientTimestamps[clientID] = active
		return false, nil
	}

	active = append(active, now)
	l.clientTimestamps[clientID] = active
	return true, nil
}
