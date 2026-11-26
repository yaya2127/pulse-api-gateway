package middleware

import (
	"net"
	"testing"
)

func BenchmarkIPFilter_Lookup(b *testing.B) {
	filter, _ := NewIPFilter(ModeAllowlist, []string{
		"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16",
	})
	ip := net.ParseIP("192.168.1.100")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filter.IsAllowed(ip)
	}
}
