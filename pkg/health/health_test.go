package health

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yaya2127/pulse-api-gateway/pkg/proxy"
)

func TestHealthCheckerAndLiveness(t *testing.T) {
	liveServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer liveServer.Close()

	u, _ := proxy.NewUpstream(liveServer.URL, 1)
	checker := NewHealthChecker([]*proxy.Upstream{u}, 50*time.Millisecond, 50*time.Millisecond, "/health")
	checker.Start()
	defer checker.Stop()

	time.Sleep(120 * time.Millisecond)
	if !u.IsAlive() {
		t.Fatal("expected upstream to be marked healthy")
	}

	liveReq := httptest.NewRequest("GET", "/live", nil)
	liveRec := httptest.NewRecorder()
	LivenessHandler().ServeHTTP(liveRec, liveReq)
	if liveRec.Code != http.StatusOK {
		t.Errorf("expected 200 from liveness probe, got %d", liveRec.Code)
	}
}
