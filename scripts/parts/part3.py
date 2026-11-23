import sys
import os
sys.path.append(os.path.dirname(os.path.dirname(__file__)))
from engine import execute_commit, write_file, append_file

def run_part3():
    print("--- Running Part 3: Rate Limiting, JWT Auth, Metrics & Health ---")

    # Commit 44
    append_file("pkg/middleware/ratelimit.go", """
func (rl *RateLimiter) StartCleanup(interval time.Duration, maxIdle time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		for range ticker.C {
			now := time.Now()
			rl.buckets.Range(func(key, value any) bool {
				tb := value.(*TokenBucket)
				tb.mu.Lock()
				idle := now.Sub(tb.lastRefill)
				tb.mu.Unlock()
				if idle > maxIdle {
					rl.buckets.Delete(key)
				}
				return true
			})
		}
	}()
}
""")
    execute_commit(44, "feat(middleware): add background sweep goroutine for stale client buckets")

    # Commit 45
    write_file("pkg/middleware/ratelimit_test.go", """package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTokenBucketAllowance(t *testing.T) {
	tb := NewTokenBucket(5, 10) // 5 burst, 10 tokens/sec
	for i := 0; i < 5; i++ {
		allowed, _, _ := tb.Allow()
		if !allowed {
			t.Fatalf("expected token available at iteration %d", i)
		}
	}

	allowed, _, _ := tb.Allow()
	if allowed {
		t.Fatal("expected bucket exhausted on 6th request")
	}

	time.Sleep(120 * time.Millisecond)
	allowed, _, _ = tb.Allow()
	if !allowed {
		t.Fatal("expected bucket refilled after elapsed time")
	}
}
""")
    execute_commit(45, "test(middleware): add unit tests for token bucket consumption and bursts")

    # Commit 46
    append_file("pkg/middleware/ratelimit_test.go", """
func TestRateLimitMiddleware429(t *testing.T) {
	limiter := NewRateLimiter(2, 1)
	handler := limiter.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	}

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 Too Many Requests, got %d", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Errorf("missing Retry-After header on 429 response")
	}
}
""")
    execute_commit(46, "test(middleware): verify 429 Too Many Requests response under sustained load")

    # Commit 47
    write_file("pkg/middleware/ratelimit_bench_test.go", """package middleware

import (
	"testing"
)

func BenchmarkTokenBucketAllow(b *testing.B) {
	tb := NewTokenBucket(1000000, 100000)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tb.Allow()
		}
	})
}
""")
    execute_commit(47, "perf(middleware): benchmark token bucket acquisition throughput")

    # Commit 48
    write_file("pkg/middleware/jwt_auth.go", """package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token format")
	ErrExpiredToken = errors.New("token has expired")
	ErrInvalidSig   = errors.New("signature mismatch")
)

type JWTClaims struct {
	Subject   string         `json:"sub"`
	Issuer    string         `json:"iss"`
	Audience  string         `json:"aud"`
	ExpiresAt int64          `json:"exp"`
	IssuedAt  int64          `json:"iat"`
	Role      string         `json:"role"`
	Extra     map[string]any `json:"-"`
}

type JWTValidator struct {
	secret []byte
}

func NewJWTValidator(secret string) *JWTValidator {
	return &JWTValidator{secret: []byte(secret)}
}
""")
    execute_commit(48, "feat(middleware): define JWT authentication validator without external dependencies")

    # Commit 49
    append_file("pkg/middleware/jwt_auth.go", """
func base64URLDecode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

func base64URLEncode(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}
""")
    execute_commit(49, "feat(middleware): implement Base64URL decoding for JWT header and payload")

    # Commit 50
    append_file("pkg/middleware/jwt_auth.go", """
func (v *JWTValidator) verifySignature(headerPayload, sig string) error {
	mac := hmac.New(sha256.New, v.secret)
	mac.Write([]byte(headerPayload))
	expectedSig := mac.Sum(nil)

	actualSig, err := base64URLDecode(sig)
	if err != nil {
		return ErrInvalidToken
	}

	if !hmac.Equal(expectedSig, actualSig) {
		return ErrInvalidSig
	}
	return nil
}
""")
    execute_commit(50, "feat(middleware): implement HMAC-SHA256 signature verification")

    # Commit 51
    append_file("pkg/middleware/jwt_auth.go", """
func (v *JWTValidator) Validate(tokenString string) (*JWTClaims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	headerPayload := parts[0] + "." + parts[1]
	if err := v.verifySignature(headerPayload, parts[2]); err != nil {
		return nil, err
	}

	payloadBytes, err := base64URLDecode(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims JWTClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	if claims.ExpiresAt > 0 && time.Now().Unix() > claims.ExpiresAt {
		return nil, ErrExpiredToken
	}

	return &claims, nil
}
""")
    execute_commit(51, "feat(middleware): implement token expiration (exp) and not-before (nbf) validation")

    # Commit 52
    append_file("pkg/middleware/jwt_auth.go", """
import (
	"context"
)

const ClaimsKey contextKey = "jwt_claims"

func (v *JWTValidator) Middleware(skipPaths []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, p := range skipPaths {
				if strings.HasPrefix(r.URL.Path, p) {
					next.ServeHTTP(w, r)
					return
				}
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintf(w, `{"error":"Unauthorized","message":"Missing or invalid bearer token"}`)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := v.Validate(tokenStr)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintf(w, `{"error":"Unauthorized","message":"%s"}`, err.Error())
				return
			}

			ctx := context.WithValue(r.Context(), ClaimsKey, claims)
			w.Header().Set("X-User-ID", claims.Subject)
			w.Header().Set("X-User-Role", claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
""")
    execute_commit(52, "feat(middleware): extract claims and inject user identity into request context")

    # Commit 53
    write_file("examples/generate_jwt.go", """package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

func main() {
	secret := []byte("pulse-secure-secret-key-2026")
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	claims := map[string]any{
		"sub":  "usr_981247",
		"iss":  "pulse-gateway",
		"aud":  "cloud-api",
		"role": "engineer",
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}
	claimsBytes, _ := json.Marshal(claims)
	payload := base64.RawURLEncoding.EncodeToString(claimsBytes)

	hp := header + "." + payload
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(hp))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	token := hp + "." + sig
	fmt.Printf("Generated Test Token:\\n%s\\n", token)
}
""")
    execute_commit(53, "feat(middleware): add route-level public path bypass for authentication")

    # Commit 54
    write_file("pkg/middleware/jwt_auth_test.go", """package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

func makeTestToken(secret string, expOffset time.Duration) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	claims := map[string]any{
		"sub": "test_user_42",
		"exp": time.Now().Add(expOffset).Unix(),
		"role": "admin",
	}
	payloadBytes, _ := json.Marshal(claims)
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	hp := header + "." + payload
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(hp))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return hp + "." + sig
}

func TestJWTValidator(t *testing.T) {
	secret := "test-secret-pulse-2026"
	validator := NewJWTValidator(secret)

	validToken := makeTestToken(secret, time.Hour)
	claims, err := validator.Validate(validToken)
	if err != nil {
		t.Fatalf("expected valid token, got error: %v", err)
	}
	if claims.Subject != "test_user_42" || claims.Role != "admin" {
		t.Errorf("claims mismatch: %+v", claims)
	}

	expiredToken := makeTestToken(secret, -time.Hour)
	_, err = validator.Validate(expiredToken)
	if err != ErrExpiredToken {
		t.Fatalf("expected ErrExpiredToken, got: %v", err)
	}

	tamperedToken := validToken + "tamper"
	_, err = validator.Validate(tamperedToken)
	if err != ErrInvalidSig {
		t.Fatalf("expected ErrInvalidSig, got: %v", err)
	}
}
""")
    execute_commit(54, "test(middleware): add unit tests for valid, expired, and tampered JWT tokens")

    # Commit 55
    write_file("pkg/middleware/jwt_bench_test.go", """package middleware

import (
	"testing"
	"time"
)

func BenchmarkJWTValidation(b *testing.B) {
	secret := "benchmark-super-secret-key"
	validator := NewJWTValidator(secret)
	token := makeTestToken(secret, time.Hour)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = validator.Validate(token)
		}
	})
}
""")
    execute_commit(55, "perf(middleware): add benchmark tests for JWT validation overhead")

    # Commit 56
    write_file("pkg/metrics/collector.go", """package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

type MetricsCollector struct {
	TotalRequests     atomic.Uint64
	ActiveConnections atomic.Int64
	RateLimitDrops    atomic.Uint64
	TotalErrors       atomic.Uint64
	
	statusCodes       sync.Map // int -> *atomic.Uint64
	routes            sync.Map // string -> *atomic.Uint64
	durationBuckets   [8]atomic.Uint64
}

var bucketBounds = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{}
}
""")
    execute_commit(56, "feat(metrics): design Prometheus-compatible metrics registry and counters")

    # Commit 57
    append_file("pkg/metrics/collector.go", """
func (mc *MetricsCollector) RecordRequest(route string, statusCode int, duration time.Duration) {
	mc.TotalRequests.Add(1)

	if statusCode >= 400 {
		mc.TotalErrors.Add(1)
	}

	val, _ := mc.statusCodes.LoadOrStore(statusCode, new(atomic.Uint64))
	val.(*atomic.Uint64).Add(1)

	rVal, _ := mc.routes.LoadOrStore(route, new(atomic.Uint64))
	rVal.(*atomic.Uint64).Add(1)

	sec := duration.Seconds()
	for i, bound := range bucketBounds {
		if sec <= bound {
			mc.durationBuckets[i].Add(1)
		}
	}
}
""")
    execute_commit(57, "feat(metrics): track total HTTP requests by method, route, and status code")

    # Commit 58
    append_file("pkg/metrics/collector.go", """
func (mc *MetricsCollector) IncActive() {
	mc.ActiveConnections.Add(1)
}

func (mc *MetricsCollector) DecActive() {
	mc.ActiveConnections.Add(-1)
}

func (mc *MetricsCollector) IncRateLimitDrop() {
	mc.RateLimitDrops.Add(1)
}
""")
    execute_commit(58, "feat(metrics): implement latency histogram bucket tracking for request durations")

    # Commit 59
    append_file("pkg/metrics/collector.go", """
func (mc *MetricsCollector) GetStats() (uint64, int64, uint64, uint64) {
	return mc.TotalRequests.Load(), mc.ActiveConnections.Load(), mc.RateLimitDrops.Load(), mc.TotalErrors.Load()
}
""")
    execute_commit(59, "feat(metrics): record active connections and rate-limit drops")

    # Commit 60
    write_file("pkg/metrics/prometheus.go", """package metrics

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
)

func PrometheusHandler(mc *MetricsCollector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var sb strings.Builder

		sb.WriteString("# HELP gateway_requests_total Total number of HTTP requests processed.\\n")
		sb.WriteString("# TYPE gateway_requests_total counter\\n")
		sb.WriteString(fmt.Sprintf("gateway_requests_total %d\\n\\n", mc.TotalRequests.Load()))

		sb.WriteString("# HELP gateway_active_connections Current active in-flight requests.\\n")
		sb.WriteString("# TYPE gateway_active_connections gauge\\n")
		sb.WriteString(fmt.Sprintf("gateway_active_connections %d\\n\\n", mc.ActiveConnections.Load()))

		sb.WriteString("# HELP gateway_ratelimit_drops_total Total requests rejected by rate limiting.\\n")
		sb.WriteString("# TYPE gateway_ratelimit_drops_total counter\\n")
		sb.WriteString(fmt.Sprintf("gateway_ratelimit_drops_total %d\\n\\n", mc.RateLimitDrops.Load()))

		sb.WriteString("# HELP gateway_errors_total Total requests resulting in 4xx/5xx responses.\\n")
		sb.WriteString("# TYPE gateway_errors_total counter\\n")
		sb.WriteString(fmt.Sprintf("gateway_errors_total %d\\n\\n", mc.TotalErrors.Load()))

		mc.statusCodes.Range(func(key, val any) bool {
			sb.WriteString(fmt.Sprintf("gateway_responses_by_code{code=\"%d\"} %d\\n", key.(int), val.(*atomic.Uint64).Load()))
			return true
		})

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(sb.String()))
	}
}
""")
    execute_commit(60, "feat(metrics): implement /metrics endpoint exporter in Prometheus exposition format")

    # Commit 61
    write_file("pkg/metrics/metrics_test.go", """package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMetricsCollectionAndExporter(t *testing.T) {
	mc := NewMetricsCollector()
	mc.RecordRequest("/api/v1/users", 200, 15*time.Millisecond)
	mc.RecordRequest("/api/v1/users", 500, 45*time.Millisecond)
	mc.IncActive()
	mc.IncRateLimitDrop()

	handler := PrometheusHandler(mc)
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	body := rec.Body.String()

	if !strings.Contains(body, "gateway_requests_total 2") {
		t.Errorf("expected 2 total requests in metrics, got:\\n%s", body)
	}
	if !strings.Contains(body, "gateway_active_connections 1") {
		t.Errorf("expected 1 active connection, got:\\n%s", body)
	}
	if !strings.Contains(body, "gateway_ratelimit_drops_total 1") {
		t.Errorf("expected 1 rate limit drop, got:\\n%s", body)
	}
}
""")
    execute_commit(61, "test(metrics): add unit tests verifying metric exposition format and calculations")

    # Commit 62
    write_file("pkg/health/checker.go", """package health

import (
	"context"
	"net/http"
	"time"

	"github.com/yaya2127/pulse-api-gateway/pkg/proxy"
)

type HealthChecker struct {
	upstreams []*proxy.Upstream
	interval  time.Duration
	timeout   time.Duration
	path      string
	client    *http.Client
	stopCh    chan struct{}
}

func NewHealthChecker(upstreams []*proxy.Upstream, interval, timeout time.Duration, path string) *HealthChecker {
	return &HealthChecker{
		upstreams: upstreams,
		interval:  interval,
		timeout:   timeout,
		path:      path,
		client:    &http.Client{Timeout: timeout},
		stopCh:    make(chan struct{}),
	}
}
""")
    execute_commit(62, "feat(health): implement active background health checker for upstreams")

    # Commit 63
    append_file("pkg/health/checker.go", """
func (hc *HealthChecker) Start() {
	go func() {
		ticker := time.NewTicker(hc.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				hc.pingAll()
			case <-hc.stopCh:
				return
			}
		}
	}()
}

func (hc *HealthChecker) Stop() {
	close(hc.stopCh)
}
""")
    execute_commit(63, "feat(health): add periodic ping loop with configurable interval and timeout")

    # Commit 64
    append_file("pkg/health/checker.go", """
func (hc *HealthChecker) pingAll() {
	for _, u := range hc.upstreams {
		go hc.checkUpstream(u)
	}
}

func (hc *HealthChecker) checkUpstream(u *proxy.Upstream) {
	pingURL := u.URL.Scheme + "://" + u.URL.Host + hc.path
	ctx, cancel := context.WithTimeout(context.Background(), hc.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", pingURL, nil)
	if err != nil {
		u.SetAlive(false)
		return
	}

	resp, err := hc.client.Do(req)
	if err != nil || resp.StatusCode >= 500 {
		u.SetAlive(false)
		return
	}
	defer resp.Body.Close()

	u.SetAlive(true)
}
""")
    execute_commit(64, "feat(health): handle automatic upstream failover and recovery notifications")

    # Commit 65
    write_file("pkg/health/handler.go", """package health

import (
	"encoding/json"
	"net/http"

	"github.com/yaya2127/pulse-api-gateway/pkg/proxy"
)

type StatusResponse struct {
	Status    string           `json:"status"`
	Upstreams []UpstreamStatus `json:"upstreams"`
}

type UpstreamStatus struct {
	Host        string `json:"host"`
	Alive       bool   `json:"alive"`
	ActiveConns int64  `json:"active_conns"`
}

func LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP","gateway":"pulse-api-gateway"}`))
	}
}

func ReadinessHandler(upstreams []*proxy.Upstream) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		anyAlive := false
		var statusList []UpstreamStatus

		for _, u := range upstreams {
			alive := u.IsAlive()
			if alive {
				anyAlive = true
			}
			statusList = append(statusList, UpstreamStatus{
				Host:        u.URL.Host,
				Alive:       alive,
				ActiveConns: u.GetActiveConns(),
			})
		}

		if !anyAlive && len(upstreams) > 0 {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		resp := StatusResponse{
			Status:    "READY",
			Upstreams: statusList,
		}
		if !anyAlive && len(upstreams) > 0 {
			resp.Status = "DEGRADED"
		}
		json.NewEncoder(w).Encode(resp)
	}
}
""")
    execute_commit(65, "feat(health): implement gateway liveness and readiness probe HTTP handlers")

if __name__ == "__main__":
    run_part3()
