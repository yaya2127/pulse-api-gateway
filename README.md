# 🛡️ Pulse API Gateway

[![Go Version](https://img.shields.io/badge/go-1.22-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](#)

A high-performance cloud API gateway and reverse proxy built in Go with zero external dependencies. Features dynamic load balancing, token-bucket rate limiting, HMAC-SHA256 JWT authentication, circuit breaking, and Prometheus metrics.

---

## ✨ Features

- **Dynamic Load Balancing**: Round-Robin, Least Connections, IP Hash, and Weighted Round-Robin.
- **Distributed Rate Limiting**: Token-bucket algorithm with burst handling and automated stale bucket GC.
- **HMAC JWT Authentication**: Zero-alloc base64url and cryptographic HMAC-SHA256 signature validation.
- **Circuit Breaker**: Finite state machine with automated cooldowns and half-open health probes.
- **Prometheus Metrics**: Built-in `/metrics` exporter exposing request rates, status codes, and latency histograms.
- **Observability**: Structured access logging with unique `X-Request-ID` tracing across requests.
- **Container Ready**: Lightweight multi-stage Docker build under 15MB.

---

## 🚀 Quickstart

### Build and Run Locally
```bash
# Clone the repository
git clone https://github.com/yaya2127/pulse-api-gateway.git
cd pulse-api-gateway

# Run the test suite
go test -v ./...

# Build the gateway binary
go build -o bin/pulse-gateway ./cmd/gateway

# Launch gateway with default config
./bin/pulse-gateway -config config/config.json
```

### Docker
```bash
docker-compose up -d --build
```

---

## 📊 Endpoints

| Endpoint | Port | Description |
|---|---|---|
| `/live` | 8080 | Gateway liveness probe |
| `/ready` | 8080 | Gateway upstream readiness probe |
| `/metrics` | 9090 | Prometheus metrics text format |
| `/api/v1/*` | 8080 | Proxied microservice routes |

---

## 👤 Author
- **Yared Kinfe** ([@yaya2127](https://github.com/yaya2127))
- Portfolio: [yareds.tech](https://yareds.tech)

## 🏷️ Release
- **v1.0.0**: Initial production release with full middleware pipeline, circuit breaker, and metrics exporter.

## v1.1.0 Release Highlights
* **W3C Distributed Tracing**: Full `traceparent` propagation across microservices.
* **Canary Blue/Green Engine**: Safe progressive rollouts with sticky sessions.
* **Mutual TLS (mTLS)**: Enterprise zero-trust certificate validation.
* **Smooth Weighted Round Robin**: High-concurrency balanced traffic distribution.
