package proxy

import (
	"net/url"
	"testing"
)

func BenchmarkSmoothWeightedRR_Next(b *testing.B) {
	uA, _ := url.Parse("http://nodeA")
	uB, _ := url.Parse("http://nodeB")
	nodes := []*WeightedNode{
		{Upstream: &Upstream{URL: uA, Healthy: true}, Weight: 7},
		{Upstream: &Upstream{URL: uB, Healthy: true}, Weight: 3},
	}
	balancer := NewSmoothWeightedRR(nodes)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = balancer.Next()
	}
}
