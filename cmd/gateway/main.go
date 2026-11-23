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
