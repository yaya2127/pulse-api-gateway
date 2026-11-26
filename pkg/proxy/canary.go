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
