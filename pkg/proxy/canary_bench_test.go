package proxy

import (
	"net/http"
	"net/url"
	"testing"
)

func BenchmarkCanaryRouter_Select(b *testing.B) {
	u1, _ := url.Parse("http://stable")
	u2, _ := url.Parse("http://canary")
	router := NewCanaryRouter(CanaryRule{WeightRatio: 10}, &Upstream{URL: u1}, &Upstream{URL: u2})
	req, _ := http.NewRequest("GET", "/", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = router.Select(req)
	}
}
