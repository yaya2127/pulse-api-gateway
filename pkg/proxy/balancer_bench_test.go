package proxy

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
