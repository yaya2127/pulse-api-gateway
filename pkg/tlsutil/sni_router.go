package tlsutil

import (
	"crypto/tls"
	"errors"
	"sync"
)

type SNIRouter struct {
	mu    sync.RWMutex
	certs map[string]*tls.Certificate
}

func NewSNIRouter() *SNIRouter {
	return &SNIRouter{
		certs: make(map[string]*tls.Certificate),
	}
}

func (s *SNIRouter) AddCertificate(domain string, cert *tls.Certificate) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.certs[domain] = cert
}

func (s *SNIRouter) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if cert, ok := s.certs[hello.ServerName]; ok {
		return cert, nil
	}

	if cert, ok := s.certs["*"]; ok {
		return cert, nil
	}

	return nil, errors.New("no matching certificate for SNI: " + hello.ServerName)
}
