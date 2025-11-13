# Quick Setup Guide

## Prerequisites

- Docker and Docker Compose installed
- Go 1.23+ (for local development)

## Option 1: Docker Compose (Recommended)

### Step 1: Start All Services

```bash
docker-compose up -d
```

This starts:
- CodeCollab Backend (port 8080)
- Prometheus (port 9090)
- Grafana (port 3000)
- Loki (port 3100)

### Step 2: Access Grafana

1. Open http://localhost:3000
2. Login with:
   - Username: `admin`
   - Password: `admin123`
3. Navigate to "Dashboards" → "CodeCollab Monitoring Dashboard"

### Step 3: Test the Setup

Run the test script:

```bash
./test-monitoring.sh
```

### Step 4: View Logs and Metrics

- **Grafana Dashboard**: http://localhost:3000
- **Prometheus Metrics**: http://localhost:9090
- **Application Metrics**: http://localhost:8080/metrics
- **Application API**: http://localhost:8080

## Option 2: Local Development

### Step 1: Install Dependencies

```bash
go mod download
```

### Step 2: Set Environment Variables

```bash
cp .env.example .env
```

Edit `.env` and configure as needed.

### Step 3: Run the Application

```bash
go run main.go
```

### Step 4: Run Monitoring Stack Separately

```bash
docker-compose up prometheus grafana loki -d
```

## What's Included

### Metrics
- HTTP request count (by method, endpoint, status)
- HTTP request duration (p50, p95, p99)
- HTTP request/response sizes
- Active/in-flight requests
- Goroutine count
- Go runtime stats (memory, GC, CPU)

### Logs
- Full request/response logging
- Method, path, headers
- Request and response bodies
- Status codes and duration
- Searchable and filterable in Grafana

### Dashboard Panels
1. Request Rate (req/s)
2. Response Time Percentiles (p50, p95, p99)
3. Error Rate (%)
4. Active Requests
5. Go Runtime Goroutines
6. Status Code Distribution
7. Top 5 Slowest Endpoints
8. Request/Response Logs
9. Error Logs (5xx)
10. Client Error Logs (4xx)

## Stopping Services

```bash
docker-compose down
```

To remove all data:

```bash
docker-compose down -v
```

## Troubleshooting

See [MONITORING.md](MONITORING.md) for detailed troubleshooting steps.

## Next Steps

- Customize the Grafana dashboard
- Set up alerting rules
- Adjust retention periods
- Configure authentication for production

For complete documentation, see [MONITORING.md](MONITORING.md).
