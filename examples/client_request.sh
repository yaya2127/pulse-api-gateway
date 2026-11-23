#!/usr/bin/env bash
# Client verification script for Pulse API Gateway

GATEWAY_HOST="http://localhost:8080"
METRICS_HOST="http://localhost:9090"

echo "=== 1. Checking Gateway Liveness ==="
curl -i -s "$GATEWAY_HOST/live"
echo -e "
"

echo "=== 2. Checking Gateway Readiness ==="
curl -i -s "$GATEWAY_HOST/ready"
echo -e "
"

echo "=== 3. Querying Prometheus Metrics ==="
curl -s "$METRICS_HOST/metrics" | head -n 20
echo -e "
"
