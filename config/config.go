package config

import (
	"time"
)

type ServerConfig struct {
	Port         int           `json:"port"`
	MetricsPort  int           `json:"metrics_port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

type UpstreamConfig struct {
	URL    string `json:"url"`
	Weight int    `json:"weight"`
}

type RouteConfig struct {
	PathPrefix      string           `json:"path_prefix"`
	StripPrefix     bool             `json:"strip_prefix"`
	BalancerStrategy string          `json:"balancer_strategy"`
	Upstreams       []UpstreamConfig `json:"upstreams"`
	RequireAuth     bool             `json:"require_auth"`
	RateLimitRPM    int              `json:"rate_limit_rpm"`
	RateLimitBurst  int              `json:"rate_limit_burst"`
	TimeoutSeconds  int              `json:"timeout_seconds"`
}

type Config struct {
	Server ServerConfig  `json:"server"`
	Routes []RouteConfig `json:"routes"`
}

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	if envPort := os.Getenv("PULSE_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			cfg.Server.Port = p
		}
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.MetricsPort == 0 {
		cfg.Server.MetricsPort = 9090
	}

	return &cfg, nil
}
