package proxy

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestReverseProxyForwarding(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Backend-Received", "true")
		fmt.Fprintf(w, "upstream response from %s", r.URL.Path)
	}))
	defer backend.Close()

	u, err := NewUpstream(backend.URL, 1)
	if err != nil {
		t.Fatalf("failed to init upstream: %v", err)
	}

	balancer := NewRoundRobinBalancer([]*Upstream{u})
	cb := NewCircuitBreaker(3, time.Second)
	proxy := NewReverseProxy(balancer, cb, 2*time.Second)

	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	rec := httptest.NewRecorder()

	proxy.ServeHTTP(rec, req)
	resp := rec.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "upstream response from /api/v1/test" {
		t.Errorf("unexpected body: %s", string(body))
	}
}
