package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

type ReverseProxy struct {
	balancer       Balancer
	circuitBreaker *CircuitBreaker
	transport      *http.Transport
	timeout        time.Duration
}

func NewReverseProxy(balancer Balancer, cb *CircuitBreaker, timeout time.Duration) *ReverseProxy {
	return &ReverseProxy{
		balancer:       balancer,
		circuitBreaker: cb,
		timeout:        timeout,
		transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		},
	}
}

func (rp *ReverseProxy) prepareRequest(r *http.Request, target *Upstream) *http.Request {
	outReq := r.Clone(r.Context())
	outReq.URL.Scheme = target.URL.Scheme
	outReq.URL.Host = target.URL.Host
	outReq.Host = target.URL.Host

	clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		prior := r.Header.Get("X-Forwarded-For")
		if prior != "" {
			clientIP = prior + ", " + clientIP
		}
		outReq.Header.Set("X-Forwarded-For", clientIP)
	}

	if r.TLS != nil {
		outReq.Header.Set("X-Forwarded-Proto", "https")
	} else {
		outReq.Header.Set("X-Forwarded-Proto", "http")
	}
	outReq.Header.Set("X-Forwarded-Host", r.Host)

	return outReq
}
