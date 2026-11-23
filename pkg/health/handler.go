package health

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
