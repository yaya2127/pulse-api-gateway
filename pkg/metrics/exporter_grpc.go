package metrics

import (
	"context"
	"sync"
	"time"
)

type MetricDataPoint struct {
	Name      string
	Value     float64
	Labels    map[string]string
	Timestamp time.Time
}

type GRPCExporter struct {
	mu     sync.Mutex
	buffer []MetricDataPoint
	limit  int
}

func NewGRPCExporter(bufferLimit int) *GRPCExporter {
	return &GRPCExporter{
		buffer: make([]MetricDataPoint, 0, bufferLimit),
		limit:  bufferLimit,
	}
}

func (e *GRPCExporter) Record(dp MetricDataPoint) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.buffer) < e.limit {
		e.buffer = append(e.buffer, dp)
	}
}

func (e *GRPCExporter) Flush(ctx context.Context) []MetricDataPoint {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := e.buffer
	e.buffer = make([]MetricDataPoint, 0, e.limit)
	return out
}
