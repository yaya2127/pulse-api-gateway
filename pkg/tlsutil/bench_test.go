package tlsutil

import (
	"crypto/tls"
	"testing"
)

func BenchmarkSNIRouter_Lookup(b *testing.B) {
	router := NewSNIRouter()
	cert := &tls.Certificate{}
	router.AddCertificate("gateway.corp.com", cert)
	hello := &tls.ClientHelloInfo{ServerName: "gateway.corp.com"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = router.GetCertificate(hello)
	}
}
