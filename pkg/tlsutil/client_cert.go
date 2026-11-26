package tlsutil

import (
	"crypto/x509"
	"errors"
	"net/http"
)

var (
	ErrNoClientCert  = errors.New("client certificate required but not provided")
	ErrUntrustedCert = errors.New("client certificate common name not permitted")
)

type MTLSValidator struct {
	allowedCNs map[string]struct{}
}

func NewMTLSValidator(allowedCNs []string) *MTLSValidator {
	m := make(map[string]struct{}, len(allowedCNs))
	for _, cn := range allowedCNs {
		m[cn] = struct{}{}
	}
	return &MTLSValidator{allowedCNs: m}
}

func (v *MTLSValidator) Validate(r *http.Request) error {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		return ErrNoClientCert
	}

	cert := r.TLS.PeerCertificates[0]
	if len(v.allowedCNs) > 0 {
		if _, ok := v.allowedCNs[cert.Subject.CommonName]; !ok {
			return ErrUntrustedCert
		}
	}
	return nil
}
