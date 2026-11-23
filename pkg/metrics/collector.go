package metrics

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
