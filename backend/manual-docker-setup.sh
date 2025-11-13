#!/bin/bash

echo "Setting up monitoring stack manually..."

docker network create monitoring-net 2>/dev/null || echo "Network already exists"

echo "Building application..."
docker build -t codecollab-backend:latest .

echo "Starting Loki..."
docker run -d \
  --name loki \
  --network monitoring-net \
  -p 3100:3100 \
  -v "$(pwd)/loki-config.yml:/etc/loki/local-config.yaml" \
  -v loki-data:/loki \
  --user 0 \
  grafana/loki:latest \
  -config.file=/etc/loki/local-config.yaml

echo "Starting Prometheus..."
docker run -d \
  --name prometheus \
  --network monitoring-net \
  -p 9090:9090 \
  -v "$(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml" \
  -v prometheus-data:/prometheus \
  prom/prometheus:latest \
  --config.file=/etc/prometheus/prometheus.yml \
  --storage.tsdb.path=/prometheus \
  --storage.tsdb.retention.time=30d \
  --web.enable-lifecycle

echo "Starting Grafana..."
docker run -d \
  --name grafana \
  --network monitoring-net \
  -p 3000:3000 \
  -e GF_SECURITY_ADMIN_USER=admin \
  -e GF_SECURITY_ADMIN_PASSWORD=admin123 \
  -e GF_USERS_ALLOW_SIGN_UP=false \
  -v "$(pwd)/grafana/provisioning:/etc/grafana/provisioning" \
  -v "$(pwd)/grafana/dashboards:/var/lib/grafana/dashboards" \
  -v grafana-data:/var/lib/grafana \
  grafana/grafana:latest

echo "Starting Application..."
docker run -d \
  --name codecollab-backend \
  --network monitoring-net \
  -p 8080:8080 \
  --env-file .env \
  -e PORT=8080 \
  -e ENV=production \
  -e LOKI_URL=http://loki:3100 \
  codecollab-backend:latest

echo ""
echo "All services started!"
echo ""
echo "Check status with: docker ps"
echo ""
echo "Access URLs:"
echo "  Grafana:    http://localhost:3000 (admin/admin123)"
echo "  Prometheus: http://localhost:9090"
echo "  Loki:       http://localhost:3100"
echo "  App:        http://localhost:8080"
