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

func (rp *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if rp.circuitBreaker != nil && !rp.circuitBreaker.Allow() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"error":"Upstream circuit open","message":"Service temporarily unavailable"}`)
		return
	}

	target, err := rp.balancer.Next(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, `{"error":"Bad gateway","message":"No available upstreams"}`)
		return
	}

	target.IncrementConns()
	defer target.DecrementConns()

	ctx, cancel := context.WithTimeout(r.Context(), rp.timeout)
	defer cancel()

	outReq := rp.prepareRequest(r, target).WithContext(ctx)
	resp, err := rp.transport.RoundTrip(outReq)
	if err != nil {
		if rp.circuitBreaker != nil {
			rp.circuitBreaker.RecordFailure()
		}
		target.RecordError()
		rp.handleError(w, r, err)
		return
	}
	defer resp.Body.Close()

	if rp.circuitBreaker != nil {
		if resp.StatusCode >= 500 {
			rp.circuitBreaker.RecordFailure()
		} else {
			rp.circuitBreaker.RecordSuccess()
		}
	}

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
