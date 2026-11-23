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

func main() {
	flag.Parse()
	printBanner()

	logger := log.New(os.Stdout, "[PULSE] ", log.LstdFlags|log.Lmicroseconds)
	logger.Printf("Loading configuration from %s...", *configPath)

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Printf("Warning: Using default fallback config (%v)", err)
		cfg = &config.Config{
			Server: config.ServerConfig{Port: *portFlag, MetricsPort: *metricsFlag},
		}
	}

	metricsCol := metrics.NewMetricsCollector()
	mux := http.NewServeMux()
	var allUpstreams []*proxy.Upstream

	for _, rCfg := range cfg.Routes {
		rp, upstreams, err := buildProxyForRoute(rCfg)
		if err != nil {
			logger.Fatalf("Failed to initialize route %s: %v", rCfg.PathPrefix, err)
		}
		allUpstreams = append(allUpstreams, upstreams...)

		chain := middleware.NewChain(
			middleware.RequestIDMiddleware,
			middleware.RecoveryMiddleware(logger),
			middleware.CORSMiddleware(middleware.DefaultCORSConfig()),
			middleware.LoggerMiddleware(logger),
		)

		if rCfg.RateLimitRPM > 0 {
			rateLimiter := middleware.NewRateLimiter(float64(rCfg.RateLimitBurst), float64(rCfg.RateLimitRPM)/60.0)
			rateLimiter.StartCleanup(10*time.Minute, 30*time.Minute)
			chain = chain.Append(rateLimiter.Middleware())
		}

		if rCfg.RequireAuth {
			jwtVal := middleware.NewJWTValidator("pulse-secure-secret-key-2026")
			chain = chain.Append(jwtVal.Middleware([]string{"/health", "/metrics"}))
		}

		mux.Handle(rCfg.PathPrefix, chain.Then(rp))
		logger.Printf("Configured route %s -> %d upstreams (strategy: %s)",
			rCfg.PathPrefix, len(upstreams), rCfg.BalancerStrategy)
	}

mux.HandleFunc("/live", health.LivenessHandler())
	mux.HandleFunc("/ready", health.ReadinessHandler(allUpstreams))

	metricsMux := http.NewServeMux()
	metricsMux.HandleFunc("/metrics", metrics.PrometheusHandler(metricsCol))

	metricsServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.MetricsPort),
		Handler: metricsMux,
	}

	go func() {
		logger.Printf("Prometheus metrics server listening on :%d", cfg.Server.MetricsPort)
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Printf("Metrics server error: %v", err)
		}
	}()

	mainServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Printf("Pulse API Gateway listening on http://0.0.0.0:%d", cfg.Server.Port)
		if err := mainServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Gateway server error: %v", err)
		}
	}()

stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Println("Initiating graceful shutdown...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = metricsServer.Shutdown(ctx)
	if err := mainServer.Shutdown(ctx); err != nil {
		logger.Fatalf("Server shutdown failed: %v", err)
	}
	logger.Println("Pulse API Gateway exited cleanly.")
}
