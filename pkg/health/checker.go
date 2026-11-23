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

func (hc *HealthChecker) pingAll() {
	for _, u := range hc.upstreams {
		go hc.checkUpstream(u)
	}
}

func (hc *HealthChecker) checkUpstream(u *proxy.Upstream) {
	pingURL := u.URL.Scheme + "://" + u.URL.Host + hc.path
	ctx, cancel := context.WithTimeout(context.Background(), hc.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", pingURL, nil)
	if err != nil {
		u.SetAlive(false)
		return
	}

	resp, err := hc.client.Do(req)
	if err != nil || resp.StatusCode >= 500 {
		u.SetAlive(false)
		return
	}
	defer resp.Body.Close()

	u.SetAlive(true)
}
