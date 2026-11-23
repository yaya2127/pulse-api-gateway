package metrics

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
		t.Errorf("expected 2 total requests in metrics, got:\n%s", body)
	}
	if !strings.Contains(body, "gateway_active_connections 1") {
		t.Errorf("expected 1 active connection, got:\n%s", body)
	}
	if !strings.Contains(body, "gateway_ratelimit_drops_total 1") {
		t.Errorf("expected 1 rate limit drop, got:\n%s", body)
	}
}
