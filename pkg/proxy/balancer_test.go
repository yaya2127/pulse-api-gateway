package proxy

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
