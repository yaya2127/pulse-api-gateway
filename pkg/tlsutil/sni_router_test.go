package tlsutil

import (
	"crypto/tls"
	"testing"
)

func TestSNIRouter(t *testing.T) {
	router := NewSNIRouter()
	dummyCert := &tls.Certificate{}
	router.AddCertificate("api.pulse.dev", dummyCert)

	hello := &tls.ClientHelloInfo{ServerName: "api.pulse.dev"}
	cert, err := router.GetCertificate(hello)
	if err != nil || cert != dummyCert {
		t.Fatalf("expected matching cert, got error %v", err)
	}

	badHello := &tls.ClientHelloInfo{ServerName: "unknown.dev"}
	if _, err := router.GetCertificate(badHello); err == nil {
		t.Fatal("expected error for unmapped domain")
	}
}
