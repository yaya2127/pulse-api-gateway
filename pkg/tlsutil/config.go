package tlsutil

import (
	"crypto/tls"
	"errors"
)

type TLSProfile string

const (
	ProfileModern       TLSProfile = "modern"
	ProfileIntermediate TLSProfile = "intermediate"
)

var (
	ErrNilTLSConfig = errors.New("tls config cannot be nil")
)

func NewTLSConfig(profile TLSProfile) *tls.Config {
	cfg := &tls.Config{
		MinVersion: tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
		PreferServerCipherSuites: true,
	}

	if profile == ProfileIntermediate {
		cfg.MinVersion = tls.VersionTLS12
		cfg.CipherSuites = append(cfg.CipherSuites,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		)
	}

	return cfg
}
