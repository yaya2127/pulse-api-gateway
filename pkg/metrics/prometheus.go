package metrics

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
)

func PrometheusHandler(mc *MetricsCollector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var sb strings.Builder

		sb.WriteString("# HELP gateway_requests_total Total number of HTTP requests processed.\n")
		sb.WriteString("# TYPE gateway_requests_total counter\n")
		sb.WriteString(fmt.Sprintf("gateway_requests_total %d\n\n", mc.TotalRequests.Load()))

		sb.WriteString("# HELP gateway_active_connections Current active in-flight requests.\n")
		sb.WriteString("# TYPE gateway_active_connections gauge\n")
		sb.WriteString(fmt.Sprintf("gateway_active_connections %d\n\n", mc.ActiveConnections.Load()))

		sb.WriteString("# HELP gateway_ratelimit_drops_total Total requests rejected by rate limiting.\n")
		sb.WriteString("# TYPE gateway_ratelimit_drops_total counter\n")
		sb.WriteString(fmt.Sprintf("gateway_ratelimit_drops_total %d\n\n", mc.RateLimitDrops.Load()))

		sb.WriteString("# HELP gateway_errors_total Total requests resulting in 4xx/5xx responses.\n")
		sb.WriteString("# TYPE gateway_errors_total counter\n")
		sb.WriteString(fmt.Sprintf("gateway_errors_total %d\n\n", mc.TotalErrors.Load()))

		mc.statusCodes.Range(func(key, val any) bool {
			sb.WriteString(fmt.Sprintf("gateway_responses_by_code{code="%d"} %d\n", key.(int), val.(*atomic.Uint64).Load()))
			return true
		})

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(sb.String()))
	}
}
