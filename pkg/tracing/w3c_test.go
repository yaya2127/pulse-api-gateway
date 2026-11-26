package tracing

import (
	"testing"
)

func TestParseTraceParent_Valid(t *testing.T) {
	header := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	tc, err := ParseTraceParent(header)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tc.TraceID.String() != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Errorf("got trace ID %s, want 4bf92f3577b34da6a3ce929d0e0e4736", tc.TraceID.String())
	}
	if tc.SpanID.String() != "00f067aa0ba902b7" {
		t.Errorf("got span ID %s, want 00f067aa0ba902b7", tc.SpanID.String())
	}
	if !tc.Sampled {
		t.Errorf("expected sampled=true")
	}

	formatted := FormatTraceParent(tc)
	if formatted != header {
		t.Errorf("got formatted %s, want %s", formatted, header)
	}
}

func TestParseTraceParent_Invalid(t *testing.T) {
	badHeaders := []string{
		"01-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		"00-short-00f067aa0ba902b7-01",
		"invalid-format",
	}

	for _, h := range badHeaders {
		if _, err := ParseTraceParent(h); err == nil {
			t.Errorf("expected error for header: %s", h)
		}
	}
}
