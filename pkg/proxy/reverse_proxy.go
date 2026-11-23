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
