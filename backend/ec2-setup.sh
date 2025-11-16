#!/bin/bash

echo "=========================================="
echo "EC2 Deployment Setup Script"
echo "=========================================="
echo ""

EC2_PUBLIC_IP=$(curl -s http://checkip.amazonaws.com)
if [ -z "$EC2_PUBLIC_IP" ]; then
    echo "Enter your EC2 public IP address:"
    read EC2_PUBLIC_IP
fi

echo "Your EC2 Public IP: $EC2_PUBLIC_IP"
echo ""

echo "Step 1: Creating .env file..."
if [ ! -f .env ]; then
    cp .env.example .env
    echo "✓ Created .env file from .env.example"
    echo "  Please edit .env and add your credentials:"
    echo "  nano .env"
else
    echo "✓ .env file already exists"
fi

echo ""
echo "Step 2: Updating Grafana configuration for public access..."
sed -i.bak "s|GF_SERVER_ROOT_URL=.*|GF_SERVER_ROOT_URL=http://${EC2_PUBLIC_IP}:3000|g" docker-compose.yml
echo "✓ Updated Grafana root URL to http://${EC2_PUBLIC_IP}:3000"

echo ""
echo "Step 3: Starting Docker services..."
docker-compose down
docker-compose up -d --build

echo ""
echo "Step 4: Waiting for services to start..."
sleep 10

echo ""
echo "=========================================="
echo "Deployment Complete!"
echo "=========================================="
echo ""
echo "Access URLs (from anywhere):"
echo "  📊 Grafana:    http://${EC2_PUBLIC_IP}:3000"
echo "     Username:   admin"
echo "     Password:   admin123"
echo ""
echo "  📈 Prometheus: http://${EC2_PUBLIC_IP}:9090"
echo "  🔍 Loki:       http://${EC2_PUBLIC_IP}:3100"
echo "  🚀 API:        http://${EC2_PUBLIC_IP}:8080"
echo "  📋 API Docs:   http://${EC2_PUBLIC_IP}:8080/docs"
echo ""
echo "Security Group Ports Required:"
echo "  - Port 8080 (Application)"
echo "  - Port 3000 (Grafana)"
echo "  - Port 9090 (Prometheus - optional)"
echo "  - Port 3100 (Loki - optional)"
echo ""
echo "Next Steps:"
echo "  1. Open AWS Console → EC2 → Security Groups"
echo "  2. Add inbound rules for ports: 8080, 3000, 9090, 3100"
echo "  3. Access Grafana at: http://${EC2_PUBLIC_IP}:3000"
echo "=========================================="
