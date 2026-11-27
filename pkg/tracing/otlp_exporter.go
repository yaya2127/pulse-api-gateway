package tracing

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

type OTLPSpan struct {
	TraceID   string            `json:"trace_id"`
	SpanID    string            `json:"span_id"`
	ParentID  string            `json:"parent_id,omitempty"`
	Name      string            `json:"name"`
	DurationUs int64            `json:"duration_us"`
	Attributes map[string]string `json:"attributes"`
}

type OTLPExporter struct {
	mu     sync.Mutex
	spans  []OTLPSpan
	limit  int
}

func NewOTLPExporter(limit int) *OTLPExporter {
	return &OTLPExporter{
		spans: make([]OTLPSpan, 0, limit),
		limit: limit,
	}
}

func (e *OTLPExporter) Export(span OTLPSpan) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.spans) < e.limit {
		e.spans = append(e.spans, span)
	}
}

func (e *OTLPExporter) FlushToJSON(ctx context.Context) ([]byte, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	data, err := json.Marshal(e.spans)
	if err == nil {
		e.spans = make([]OTLPSpan, 0, e.limit)
	}
	return data, err
}
