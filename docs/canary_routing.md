# Canary Traffic Routing & Blue/Green Deployments

Pulse API Gateway v1.1.0 introduces progressive canary deployments with zero downtime.

## Capabilities
* **Percentage-based splitting**: Route a configurable ratio of production traffic to canary pools (e.g. 90/10).
* **Header override**: Force requests to canary with `X-Canary-Release: true`.
* **Sticky sessions**: Maintain user routing stability through signed cookies.
