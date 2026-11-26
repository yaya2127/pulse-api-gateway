package metrics

import (
	"context"
	"testing"
	"time"
)

func TestGRPCExporter(t *testing.T) {
	exp := NewGRPCExporter(100)
	exp.Record(MetricDataPoint{
		Name:      "gateway_requests_total",
		Value:     1.0,
		Timestamp: time.Now(),
	})

	pts := exp.Flush(context.Background())
	if len(pts) != 1 {
		t.Fatalf("expected 1 flushed metric point, got %d", len(pts))
	}
	if pts[0].Name != "gateway_requests_total" {
		t.Errorf("got metric %s, want gateway_requests_total", pts[0].Name)
	}
}
