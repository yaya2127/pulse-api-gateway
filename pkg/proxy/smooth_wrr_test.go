package proxy

import (
	"net/url"
	"testing"
)

func TestSmoothWeightedRR(t *testing.T) {
	uA, _ := url.Parse("http://nodeA")
	uB, _ := url.Parse("http://nodeB")
	uC, _ := url.Parse("http://nodeC")

	nodes := []*WeightedNode{
		{Upstream: &Upstream{URL: uA, Healthy: true}, Weight: 5},
		{Upstream: &Upstream{URL: uB, Healthy: true}, Weight: 1},
		{Upstream: &Upstream{URL: uC, Healthy: true}, Weight: 1},
	}

	balancer := NewSmoothWeightedRR(nodes)
	counts := make(map[string]int)

	for i := 0; i < 7; i++ {
		u := balancer.Next()
		counts[u.URL.Host]++
	}

	if counts["nodeA"] != 5 || counts["nodeB"] != 1 || counts["nodeC"] != 1 {
		t.Fatalf("unexpected SWRR distribution: %v", counts)
	}
}
