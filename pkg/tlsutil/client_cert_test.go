package tlsutil

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"net/http"
	"testing"
)

func TestMTLSValidator(t *testing.T) {
	val := NewMTLSValidator([]string{"trusted-client.internal"})

	reqNoTLS, _ := http.NewRequest("GET", "/", nil)
	if err := val.Validate(reqNoTLS); err != ErrNoClientCert {
		t.Errorf("expected ErrNoClientCert, got %v", err)
	}

	reqTLS, _ := http.NewRequest("GET", "/", nil)
	reqTLS.TLS = &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{
			{Subject: pkix.Name{CommonName: "untrusted-client"}},
		},
	}
	if err := val.Validate(reqTLS); err != ErrUntrustedCert {
		t.Errorf("expected ErrUntrustedCert, got %v", err)
	}

	reqValid, _ := http.NewRequest("GET", "/", nil)
	reqValid.TLS = &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{
			{Subject: pkix.Name{CommonName: "trusted-client.internal"}},
		},
	}
	if err := val.Validate(reqValid); err != nil {
		t.Errorf("unexpected error for valid cert: %v", err)
	}
}
