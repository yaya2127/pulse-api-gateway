package proxy

import (
	"net/http"
	"net/url"
	"testing"
)

func TestCanaryRouter_HeaderMatch(t *testing.T) {
	u1, _ := url.Parse("http://stable.internal")
	u2, _ := url.Parse("http://canary.internal")

	stable := &Upstream{URL: u1, Healthy: true}
	canary := &Upstream{URL: u2, Healthy: true}

	router := NewCanaryRouter(CanaryRule{
		HeaderKey:   "X-Canary-Release",
		HeaderValue: "v2",
		WeightRatio: 0,
	}, stable, canary)

	reqMatch, _ := http.NewRequest("GET", "/", nil)
	reqMatch.Header.Set("X-Canary-Release", "v2")
	if router.Select(reqMatch) != canary {
		t.Fatal("expected canary target on header match")
	}

	reqNoMatch, _ := http.NewRequest("GET", "/", nil)
	if router.Select(reqNoMatch) != stable {
		t.Fatal("expected stable target without header match")
	}
}

func TestCanaryRouter_Percentage(t *testing.T) {
	u1, _ := url.Parse("http://stable")
	u2, _ := url.Parse("http://canary")
	stable := &Upstream{URL: u1, Healthy: true}
	canary := &Upstream{URL: u2, Healthy: true}

	router := NewCanaryRouter(CanaryRule{
		WeightRatio: 25, // 25%
	}, stable, canary)

	canaryCount := 0
	total := 1000
	req, _ := http.NewRequest("GET", "/", nil)
	for i := 0; i < total; i++ {
		if router.Select(req) == canary {
			canaryCount++
		}
	}

	expected := 250
	if canaryCount != expected {
		t.Errorf("got %d canary routes, expected %d", canaryCount, expected)
	}
}
