package health

import (
	"context"
	"net/http"
	"time"

	"github.com/yaya2127/pulse-api-gateway/pkg/proxy"
)

type HealthChecker struct {
	upstreams []*proxy.Upstream
	interval  time.Duration
	timeout   time.Duration
	path      string
	client    *http.Client
	stopCh    chan struct{}
}

func NewHealthChecker(upstreams []*proxy.Upstream, interval, timeout time.Duration, path string) *HealthChecker {
	return &HealthChecker{
		upstreams: upstreams,
		interval:  interval,
		timeout:   timeout,
		path:      path,
		client:    &http.Client{Timeout: timeout},
		stopCh:    make(chan struct{}),
	}
}

func (hc *HealthChecker) Start() {
	go func() {
		ticker := time.NewTicker(hc.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				hc.pingAll()
			case <-hc.stopCh:
				return
			}
		}
	}()
}

func (hc *HealthChecker) Stop() {
	close(hc.stopCh)
}
