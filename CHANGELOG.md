# Changelog - Pulse Cloud API Gateway

## [v1.1.0] - 2026-11-26
### Added
- W3C TraceContext distributed tracing middleware with child span inheritance
- Mutual TLS (mTLS) certificate verification with client Common Name enforcement
- SNI multi-domain virtual host router on single TLS port
- Canary routing engine with header, cookie, and percentage traffic split
- Nginx-style Smooth Weighted Round Robin (SWRR) load balancer
- Dynamic streaming response compression with sync.Pool gzip encoders
- CIDR IP subnet firewall filter for edge DDoS protection
- Configurable upstream HTTP request retry engine with exponential backoff & jitter
- OpenTelemetry metrics buffer collector
- Admin CLI utility `pulsectl` (status, drain, routes)
