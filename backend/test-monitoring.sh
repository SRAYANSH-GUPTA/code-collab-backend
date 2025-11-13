#!/bin/bash

echo "Testing CodeCollab Monitoring Setup"
echo "===================================="
echo ""

echo "1. Testing application /metrics endpoint..."
if curl -s http://localhost:8080/metrics | grep -q "http_requests_total"; then
    echo "   ✓ Metrics endpoint is working"
else
    echo "   ✗ Metrics endpoint failed"
    exit 1
fi

echo ""
echo "2. Testing Prometheus..."
if curl -s http://localhost:9090/-/ready | grep -q "Prometheus Server is Ready"; then
    echo "   ✓ Prometheus is ready"
else
    echo "   ✗ Prometheus is not ready"
fi

echo ""
echo "3. Testing Loki..."
if curl -s http://localhost:3100/ready; then
    echo "   ✓ Loki is ready"
else
    echo "   ✗ Loki is not ready"
fi

echo ""
echo "4. Testing Grafana..."
if curl -s http://localhost:3000/api/health | grep -q "ok"; then
    echo "   ✓ Grafana is healthy"
else
    echo "   ✗ Grafana is not healthy"
fi

echo ""
echo "5. Generating test traffic..."
for i in {1..20}; do
    curl -s http://localhost:8080/ > /dev/null
    curl -s http://localhost:8080/health > /dev/null
done
echo "   ✓ Generated 40 test requests"

echo ""
echo "6. Checking Prometheus targets..."
if curl -s http://localhost:9090/api/v1/targets | grep -q "codecollab-backend"; then
    echo "   ✓ Application is being scraped by Prometheus"
else
    echo "   ✗ Application not found in Prometheus targets"
fi

echo ""
echo "===================================="
echo "Monitoring Setup Test Complete!"
echo ""
echo "Access URLs:"
echo "  - Grafana Dashboard: http://localhost:3000 (admin/admin123)"
echo "  - Prometheus: http://localhost:9090"
echo "  - Application: http://localhost:8080"
echo "  - Metrics: http://localhost:8080/metrics"
echo ""
