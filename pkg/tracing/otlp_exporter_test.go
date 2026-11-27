package tracing

import (
	"context"
	"testing"
)

func TestOTLPExporter(t *testing.T) {
	exp := NewOTLPExporter(10)
	exp.Export(OTLPSpan{
		TraceID:    "4bf92f3577b34da6a3ce929d0e0e4736",
		SpanID:     "00f067aa0ba902b7",
		Name:       "HTTP GET /api/v1/orders",
		DurationUs: 820,
	})

	data, err := exp.FlushToJSON(context.Background())
	if err != nil || len(data) == 0 {
		t.Fatalf("failed to flush OTLP spans: %v", err)
	}
}
