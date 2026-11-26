package tracing

import (
	"context"
	"fmt"
	"time"
)

type contextKey string

const (
	TraceContextKey contextKey = "pulse_trace_context"
)

type TraceID [16]byte
type SpanID [8]byte

func (t TraceID) String() string {
	return fmt.Sprintf("%032x", [16]byte(t))
}

func (s SpanID) String() string {
	return fmt.Sprintf("%016x", [8]byte(s))
}

type TraceContext struct {
	TraceID    TraceID
	SpanID     SpanID
	ParentID   SpanID
	Sampled    bool
	TraceState string
	StartTime  time.Time
}

type Span struct {
	Context   TraceContext
	Name      string
	StartTime time.Time
	EndTime   time.Time
	Tags      map[string]string
}
