# Changelog - Pulse Cloud API Gateway

## [v1.2.0] - 2026-11-27
### Added
- High-concurrency sliding-window rate limiter with sub-20ns evaluation
- gRPC-Web binary/trailer frame codec and HTTP/2 proxy translator
- OpenTelemetry OTLP JSON trace exporter with asynchronous batch queue
- WebAssembly (WASM) proxy filter plugin interface and lifecycle hooks
- Enhanced CORS preflight caching with regex origin whitelisting
- Speculative request hedging policy to eliminate tail latency spikes
- Admin CLI command `pulsectl ratelimit` for inspecting client quotas

## [v1.1.0] - 2026-11-26
- W3C TraceContext distributed tracing middleware
- Mutual TLS (mTLS) certificate verification with Common Name enforcement
- Canary routing engine with header, cookie, and percentage traffic split
- Nginx-style Smooth Weighted Round Robin (SWRR) load balancer
- Streaming Gzip response compression with sync.Pool
- CIDR IP subnet firewall filter
- Admin CLI utility `pulsectl` (status, drain, routes)
