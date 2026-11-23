package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yaya2127/pulse-api-gateway/config"
	"github.com/yaya2127/pulse-api-gateway/pkg/health"
	"github.com/yaya2127/pulse-api-gateway/pkg/metrics"
	"github.com/yaya2127/pulse-api-gateway/pkg/middleware"
	"github.com/yaya2127/pulse-api-gateway/pkg/proxy"
)

var (
	configPath  = flag.String("config", "config/config.json", "Path to gateway config JSON")
	portFlag    = flag.Int("port", 8080, "Gateway HTTP listen port")
	metricsFlag = flag.Int("metrics-port", 9090, "Prometheus metrics listen port")
)

const banner = `
===================================================================
   ___       __             ___   ___ ___   ___       __                         
  / _ \__ __/ /__ ___  ___ / _ | / _ \_ _| / _ \___ _/ /____ _    _____ ___ __  
 / ___/ // / (_-</ -_)(_-</ __ |/ ___/| | / // / _ '/ __/ -_) |/|/ / _ '/ // /  
/_/   \_,_/_/___/\__//___/_/ |_/_/   |___/_/\_,_/\_,_/\__/\__/|__,__/\_,_/\_, /   
                                                                          /___/    
                  High-Performance Cloud API Gateway v1.0.0
===================================================================
`

func printBanner() {
	fmt.Println(banner)
}

func buildProxyForRoute(rCfg config.RouteConfig) (*proxy.ReverseProxy, []*proxy.Upstream, error) {
	var upstreams []*proxy.Upstream
	for _, upCfg := range rCfg.Upstreams {
		u, err := proxy.NewUpstream(upCfg.URL, upCfg.Weight)
		if err != nil {
			return nil, nil, err
		}
		upstreams = append(upstreams, u)
	}

	var balancer proxy.Balancer
	switch rCfg.BalancerStrategy {
	case "least_conn":
		balancer = proxy.NewLeastConnectionsBalancer(upstreams)
	case "ip_hash":
		balancer = proxy.NewIPHashBalancer(upstreams)
	default:
		balancer = proxy.NewRoundRobinBalancer(upstreams)
	}

	cb := proxy.NewCircuitBreaker(5, 10*time.Second)
	timeout := time.Duration(rCfg.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 15 * time.Second
	}

	return proxy.NewReverseProxy(balancer, cb, timeout), upstreams, nil
}
