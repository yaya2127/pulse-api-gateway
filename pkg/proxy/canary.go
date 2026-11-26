package proxy

import (
	"hash/fnv"
	"net/http"
	"sync/atomic"
)

type CanaryRule struct {
	HeaderKey    string
	HeaderValue  string
	CookieName   string
	WeightRatio  uint32 // e.g. 10 for 10%
}

type CanaryRouter struct {
	rule         CanaryRule
	stableTarget *Upstream
	canaryTarget *Upstream
	counter      uint64
}

func NewCanaryRouter(rule CanaryRule, stable, canary *Upstream) *CanaryRouter {
	return &CanaryRouter{
		rule:         rule,
		stableTarget: stable,
		canaryTarget: canary,
	}
}

func (c *CanaryRouter) Select(r *http.Request) *Upstream {
	// 1. Explicit Header Override
	if c.rule.HeaderKey != "" && r.Header.Get(c.rule.HeaderKey) == c.rule.HeaderValue {
		return c.canaryTarget
	}

	// 2. Cookie Sticky Routing
	if c.rule.CookieName != "" {
		if cookie, err := r.Cookie(c.rule.CookieName); err == nil && cookie.Value == "canary" {
			return c.canaryTarget
		}
	}

	// 3. Percentage Split
	if c.rule.WeightRatio > 0 {
		val := atomic.AddUint64(&c.counter, 1) % 100
		if uint32(val) < c.rule.WeightRatio {
			return c.canaryTarget
		}
	}

	return c.stableTarget
}
