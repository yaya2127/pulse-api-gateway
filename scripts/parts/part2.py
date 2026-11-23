import sys
import os
sys.path.append(os.path.dirname(os.path.dirname(__file__)))
from engine import execute_commit, write_file, append_file

def run_part2():
    print("--- Running Part 2: Reverse Proxy & Middleware Pipeline ---")

    # Commit 22
    write_file("pkg/proxy/circuit_breaker_test.go", """package proxy

import (
	"testing"
	"time"
)

func TestCircuitBreakerTransitions(t *testing.T) {
	cb := NewCircuitBreaker(3, 50*time.Millisecond)

	if !cb.Allow() {
		t.Fatal("expected circuit breaker to be closed initially")
	}

	cb.RecordFailure()
	cb.RecordFailure()
	if cb.GetState() != StateClosed {
		t.Fatalf("expected state Closed after 2 failures, got %v", cb.GetState())
	}

	cb.RecordFailure()
	if cb.GetState() != StateOpen {
		t.Fatalf("expected state Open after 3 failures, got %v", cb.GetState())
	}

	if cb.Allow() {
		t.Fatal("expected request rejected when circuit is Open")
	}

	time.Sleep(60 * time.Millisecond)
	if !cb.Allow() {
		t.Fatal("expected request permitted to probe Half-Open")
	}

	cb.RecordSuccess()
	cb.RecordSuccess()
	if cb.GetState() != StateClosed {
		t.Fatalf("expected state Closed after probationary successes, got %v", cb.GetState())
	}
}
""")
    execute_commit(22, "test(proxy): add unit tests for circuit breaker state transitions")

    # Commit 23
    append_file("pkg/proxy/circuit_breaker_test.go", """
func TestCircuitBreakerConcurrency(t *testing.T) {
	cb := NewCircuitBreaker(10, 100*time.Millisecond)
	done := make(chan bool)

	for i := 0; i < 20; i++ {
		go func() {
			for j := 0; j < 50; j++ {
				cb.Allow()
				cb.RecordSuccess()
			}
			done <- true
		}()
	}

	for i := 0; i < 20; i++ {
		<-done
	}

	if cb.GetState() != StateClosed {
		t.Errorf("expected circuit to remain Closed, got %v", cb.GetState())
	}
}
""")
    execute_commit(23, "test(proxy): add concurrency tests for multi-goroutine circuit breaker access")

    # Commit 24
    write_file("pkg/proxy/reverse_proxy.go", """package proxy

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
""")
    execute_commit(24, "feat(proxy): construct reverse proxy transport and HTTP director")

    # Commit 25
    append_file("pkg/proxy/reverse_proxy.go", """
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
""")
    execute_commit(25, "feat(proxy): inject X-Forwarded headers and request identification")

    # Commit 26
    append_file("pkg/proxy/reverse_proxy.go", """
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
""")
    execute_commit(26, "feat(proxy): integrate circuit breaker into reverse proxy execution flow")

    # Commit 27
    append_file("pkg/proxy/reverse_proxy.go", """
func (rp *ReverseProxy) handleError(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")
	if ctxErr := r.Context().Err(); ctxErr == context.DeadlineExceeded {
		w.WriteHeader(http.StatusGatewayTimeout)
		fmt.Fprintf(w, `{"error":"Gateway timeout","message":"Upstream server took too long to respond"}`)
		return
	}
	w.WriteHeader(http.StatusBadGateway)
	fmt.Fprintf(w, `{"error":"Bad gateway","message":"Failed to reach upstream service: %s"}`, err.Error())
}
""")
    execute_commit(27, "feat(proxy): implement custom error handler and graceful fallback responses")

    # Commit 28
    write_file("pkg/proxy/reverse_proxy_test.go", """package proxy

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
""")
    execute_commit(28, "test(proxy): add integration tests for reverse proxy request forwarding")

    # Commit 29
    append_file("pkg/proxy/reverse_proxy_test.go", """
func TestReverseProxyCircuitBreakerTrip(t *testing.T) {
	u, _ := NewUpstream("http://127.0.0.1:54321", 1) // unreachable port
	balancer := NewRoundRobinBalancer([]*Upstream{u})
	cb := NewCircuitBreaker(2, 500*time.Millisecond)
	proxy := NewReverseProxy(balancer, cb, 100*time.Millisecond)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		proxy.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadGateway {
			t.Errorf("expected 502, got %d", rec.Code)
		}
	}

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 circuit open, got %d", rec.Code)
	}
}
""")
    execute_commit(29, "test(proxy): test reverse proxy timeout handling and backend failure responses")

    # Commit 30
    write_file("pkg/middleware/chain.go", """package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

type Chain struct {
	middlewares []Middleware
}

func NewChain(middlewares ...Middleware) Chain {
	return Chain{middlewares: append(([]Middleware)(nil), middlewares...)}
}

func (c Chain) Then(h http.Handler) http.Handler {
	if h == nil {
		h = http.DefaultServeMux
	}
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		h = c.middlewares[i](h)
	}
	return h
}

func (c Chain) Append(m ...Middleware) Chain {
	newMiddlewares := make([]Middleware, 0, len(c.middlewares)+len(m))
	newMiddlewares = append(newMiddlewares, c.middlewares...)
	newMiddlewares = append(newMiddlewares, m...)
	return Chain{middlewares: newMiddlewares}
}
""")
    execute_commit(30, "feat(middleware): create composable HTTP middleware pipeline and chain helper")

    # Commit 31
    write_file("pkg/middleware/request_id.go", """package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type contextKey string

const RequestIDKey contextKey = "X-Request-ID"

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			bytes := make([]byte, 16)
			if _, err := rand.Read(bytes); err == nil {
				reqID = hex.EncodeToString(bytes)
			} else {
				reqID = "req-default-fallback"
			}
		}

		ctx := context.WithValue(r.Context(), RequestIDKey, reqID)
		w.Header().Set("X-Request-ID", reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
""")
    execute_commit(31, "feat(middleware): implement unique request ID generator middleware")

    # Commit 32
    write_file("pkg/middleware/recovery.go", """package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

func RecoveryMiddleware(logger *log.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					stack := debug.Stack()
					if logger != nil {
						logger.Printf("[PANIC RECOVERED] %v\\n%s", rec, string(stack))
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, `{"error":"Internal server error","message":"An unexpected error occurred"}`)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
""")
    execute_commit(32, "feat(middleware): implement panic recovery middleware with stack tracing")

    # Commit 33
    write_file("pkg/middleware/recovery_test.go", """package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecoveryMiddleware(t *testing.T) {
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("critical null pointer simulation")
	})

	recovery := RecoveryMiddleware(nil)(panicHandler)
	req := httptest.NewRequest("GET", "/panic", nil)
	rec := httptest.NewRecorder()

	recovery.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 status code after panic, got %d", rec.Code)
	}
}
""")
    execute_commit(33, "test(middleware): add unit tests for panic recovery and error isolation")

    # Commit 34
    write_file("pkg/middleware/logger.go", """package middleware

import (
	"log"
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	bytesWritten int64
}

func (lw *loggingResponseWriter) WriteHeader(code int) {
	lw.statusCode = code
	lw.ResponseWriter.WriteHeader(code)
}

func (lw *loggingResponseWriter) Write(b []byte) (int, error) {
	n, err := lw.ResponseWriter.Write(b)
	lw.bytesWritten += int64(n)
	return n, err
}
""")
    execute_commit(34, "feat(middleware): implement structured access logger middleware")

    # Commit 35
    append_file("pkg/middleware/logger.go", """
func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}
""")
    execute_commit(35, "feat(middleware): add response writer wrapper to capture status code and bytes")

    # Commit 36
    append_file("pkg/middleware/logger.go", """
func LoggerMiddleware(logger *log.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lw := newLoggingResponseWriter(w)

			next.ServeHTTP(lw, r)

			duration := time.Since(start)
			reqID := r.Header.Get("X-Request-ID")
			if reqID == "" {
				reqID = "-"
			}

			if logger != nil {
				logger.Printf("[%s] %s %s %d %s %d bytes (req_id: %s)",
					r.Method, r.URL.Path, r.RemoteAddr, lw.statusCode, duration, lw.bytesWritten, reqID)
			}
		})
	}
}
""")
    execute_commit(36, "feat(middleware): format log entries with duration, IP, route, and status")

    # Commit 37
    write_file("pkg/middleware/cors.go", """package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAgeSeconds    int
}

func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		AllowCredentials: true,
		MaxAgeSeconds: 86400,
	}
}
""")
    execute_commit(37, "feat(middleware): implement cross-origin resource sharing (CORS) handler")

    # Commit 38
    append_file("pkg/middleware/cors.go", """
func CORSMiddleware(cfg CORSConfig) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", strings.Join(cfg.AllowedOrigins, ", "))
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
				if cfg.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
				if cfg.MaxAgeSeconds > 0 {
					w.Header().Set("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAgeSeconds))
				}
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
""")
    execute_commit(38, "feat(middleware): add CORS preflight OPTIONS response handling")

    # Commit 39
    write_file("pkg/middleware/cors_test.go", """package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddleware(t *testing.T) {
	cfg := DefaultCORSConfig()
	handler := CORSMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content for preflight, got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Errorf("missing Access-Control-Allow-Origin header")
	}
}
""")
    execute_commit(39, "test(middleware): write unit tests for CORS headers and origin filtering")

    # Commit 40
    write_file("pkg/middleware/ratelimit.go", """package middleware

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type TokenBucket struct {
	capacity   float64
	tokens     float64
	refillRate float64
	lastRefill time.Time
	mu         sync.Mutex
}

func NewTokenBucket(capacity float64, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}
""")
    execute_commit(40, "feat(middleware): implement token-bucket rate limiter structure")

    # Commit 41
    append_file("pkg/middleware/ratelimit.go", """
func (tb *TokenBucket) Allow() (bool, int, time.Duration) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.lastRefill = now

	tb.tokens = math.Min(tb.capacity, tb.tokens+elapsed*tb.refillRate)

	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		remaining := int(tb.tokens)
		return true, remaining, 0
	}

	missing := 1.0 - tb.tokens
	retryAfter := time.Duration((missing / tb.refillRate) * float64(time.Second))
	return false, 0, retryAfter
}
""")
    execute_commit(41, "feat(middleware): implement fractional token refill algorithm based on elapsed time")

    # Commit 42
    append_file("pkg/middleware/ratelimit.go", """
type RateLimiter struct {
	buckets    sync.Map
	capacity   float64
	refillRate float64
}

func NewRateLimiter(capacity float64, refillRate float64) *RateLimiter {
	return &RateLimiter{
		capacity:   capacity,
		refillRate: refillRate,
	}
}

func (rl *RateLimiter) getBucket(key string) *TokenBucket {
	val, loaded := rl.buckets.Load(key)
	if loaded {
		return val.(*TokenBucket)
	}
	newBucket := NewTokenBucket(rl.capacity, rl.refillRate)
	actual, _ := rl.buckets.LoadOrStore(key, newBucket)
	return actual.(*TokenBucket)
}
""")
    execute_commit(42, "feat(middleware): implement thread-safe client rate limit registry")

    # Commit 43
    append_file("pkg/middleware/ratelimit.go", """
func (rl *RateLimiter) Middleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				clientIP = r.RemoteAddr
			}

			bucket := rl.getBucket(clientIP)
			allowed, remaining, retryAfter := bucket.Allow()

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(int(rl.capacity)))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

			if !allowed {
				retrySec := int(math.Ceil(retryAfter.Seconds()))
				if retrySec < 1 {
					retrySec = 1
				}
				w.Header().Set("Retry-After", strconv.Itoa(retrySec))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				fmt.Fprintf(w, `{"error":"Rate limit exceeded","retry_after_seconds":%d}`, retrySec)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
""")
    execute_commit(43, "feat(middleware): add rate limit HTTP headers (Limit, Remaining, Reset)")

if __name__ == "__main__":
    run_part2()
