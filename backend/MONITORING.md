# CodeCollab Monitoring & Analytics

This document describes the complete monitoring and analytics setup for the CodeCollab backend using Prometheus, Grafana, and Loki.

## Overview

The monitoring stack consists of:

- **Prometheus**: Metrics collection and storage
- **Grafana**: Visualization and dashboarding
- **Loki**: Log aggregation and querying
- **Go Application**: Instrumented with Prometheus client and Loki logger

## Quick Start

### Prerequisites

- Docker and Docker Compose installed
- Ports 3000, 8080, 9090, and 3100 available

### Running the Stack

Start all services with a single command:

```bash
docker-compose up -d
```

This will start:
- CodeCollab Backend (port 8080)
- Prometheus (port 9090)
- Grafana (port 3000)
- Loki (port 3100)

### Accessing Services

#### Grafana Dashboard
- **URL**: http://localhost:3000
- **Username**: admin
- **Password**: admin123
- **Dashboard**: Navigate to "CodeCollab Monitoring Dashboard" (auto-imported)

#### Prometheus
- **URL**: http://localhost:9090
- **Metrics endpoint**: http://localhost:8080/metrics

#### Application
- **API**: http://localhost:8080
- **Health Check**: http://localhost:8080/health
- **API Docs**: http://localhost:8080/docs
- **Metrics**: http://localhost:8080/metrics

### Stopping Services

```bash
docker-compose down
```

To remove volumes (data will be lost):

```bash
docker-compose down -v
```

## Metrics Tracked

### HTTP Metrics

1. **Request Count** (`http_requests_total`)
   - Counter of total HTTP requests
   - Labels: method, endpoint, status
   - Use case: Track request volume and patterns

2. **Request Duration** (`http_request_duration_seconds`)
   - Histogram of request latency
   - Labels: method, endpoint, status
   - Buckets: Default Prometheus buckets
   - Use case: Monitor response times (p50, p95, p99)

3. **Request Size** (`http_request_size_bytes`)
   - Histogram of request body sizes
   - Labels: method, endpoint
   - Buckets: Exponential (100, 1000, 10000, ...)
   - Use case: Track payload sizes

4. **Response Size** (`http_response_size_bytes`)
   - Histogram of response body sizes
   - Labels: method, endpoint
   - Buckets: Exponential (100, 1000, 10000, ...)
   - Use case: Monitor bandwidth usage

5. **In-Flight Requests** (`http_requests_in_flight`)
   - Gauge of currently processing requests
   - Use case: Monitor concurrent load

### System Metrics

1. **Goroutines** (`go_goroutines_count`)
   - Gauge of active goroutines
   - Collected every 15 seconds
   - Use case: Monitor concurrency and potential goroutine leaks

2. **Go Runtime Metrics** (built-in)
   - Memory usage
   - GC statistics
   - CPU usage
   - Automatically collected by Prometheus client

## Logging with Loki

### What Gets Logged

Every HTTP request/response is logged with:

- Timestamp
- HTTP method
- Request path
- Query parameters
- Request headers
- Request body (up to 10KB)
- Response body (up to 10KB)
- Status code
- Request duration (ms)
- Client IP address
- User agent
- Content length
- Response size

### Log Labels

All logs are tagged with:
- `job`: codecollab-backend
- `service`: codecollab
- `environment`: production/development
- `method`: HTTP method
- `path`: Request path
- `status`: HTTP status code

### Querying Logs in Grafana

Example LogQL queries:

```logql
{job="codecollab-backend"}

{job="codecollab-backend",status="500"}

{job="codecollab-backend",method="POST"}

{job="codecollab-backend",path="/ws"}

{job="codecollab-backend"} |= "error"

{job="codecollab-backend"} | json | duration_ms > 1000
```

## Dashboard Panels

The Grafana dashboard includes:

### 1. Request Rate (req/s)
- Time series of requests per second
- Grouped by method and endpoint
- Query: `rate(http_requests_total[1m])`

### 2. Response Time Percentiles
- p50, p95, p99 latency over time
- Helps identify performance degradation
- Query: `histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))`

### 3. Error Rate (%)
- Gauge showing percentage of 5xx errors
- Color-coded thresholds (green < 1%, yellow < 5%, red >= 5%)
- Query: `100 * (sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])))`

### 4. Active Requests
- Current number of in-flight requests
- Stat panel showing real-time value
- Query: `http_requests_in_flight`

### 5. Go Runtime - Goroutines
- Time series of goroutine count
- Helps detect goroutine leaks
- Query: `go_goroutines_count`

### 6. Status Code Distribution
- Pie chart of HTTP status codes
- Visual breakdown of 2xx, 4xx, 5xx responses
- Query: `sum by (status) (rate(http_requests_total[5m]))`

### 7. Top 5 Slowest Endpoints (p99)
- Table showing endpoints with highest p99 latency
- Sorted by latency descending
- Query: `topk(5, histogram_quantile(0.99, sum by (endpoint, le) (rate(http_request_duration_seconds_bucket[5m]))))`

### 8. Request/Response Logs
- Full log viewer with all HTTP requests/responses
- Searchable and filterable
- Shows complete request/response details

### 9. Error Logs (5xx)
- Filtered view showing only server errors
- Quick access to error investigation

### 10. Client Error Logs (4xx)
- Filtered view showing client errors
- Helps identify API misuse patterns

## Configuration Files

### prometheus.yml
Prometheus scrape configuration:
- Scrapes CodeCollab app every 15 seconds
- Collects metrics from `/metrics` endpoint
- Includes job labels for service identification

### loki-config.yml
Loki log storage configuration:
- 7-day retention period
- Filesystem storage
- Rate limiting: 10MB/s ingestion
- TSDB schema for efficient querying

### docker-compose.yml
Container orchestration:
- All services on `monitoring` network
- Persistent volumes for data retention
- Health checks and restart policies
- Environment variable configuration

### Grafana Provisioning

**datasources.yml**: Auto-configures:
- Prometheus datasource (default)
- Loki datasource
- Both accessible via service discovery

**dashboard.yml**: Auto-imports:
- CodeCollab Monitoring Dashboard
- Updates every 10 seconds
- Allows UI updates

## Development Setup

### Running Locally (without Docker)

1. Install dependencies:
```bash
go mod download
```

2. Set environment variables:
```bash
export LOKI_URL=http://localhost:3100
export PORT=8080
export ENV=development
```

3. Run the application:
```bash
go run main.go
```

4. Access metrics:
```bash
curl http://localhost:8080/metrics
```

### Testing Metrics

Generate test traffic:

```bash
for i in {1..100}; do
  curl http://localhost:8080/
  curl http://localhost:8080/health
done
```

View metrics in Prometheus:
- http://localhost:9090/graph

Query examples:
- `http_requests_total`
- `rate(http_requests_total[1m])`
- `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))`

## Production Considerations

### Security

1. **Change Grafana password**:
   - Update `GF_SECURITY_ADMIN_PASSWORD` in docker-compose.yml
   - Use strong passwords in production

2. **Secure Prometheus**:
   - Add authentication if exposing externally
   - Use network policies to restrict access

3. **Rate Limiting**:
   - Loki has built-in rate limiting (10MB/s)
   - Adjust `ingestion_rate_mb` if needed

### Performance

1. **Metrics Cardinality**:
   - Current setup uses endpoint paths as labels
   - For high-cardinality endpoints (e.g., /user/:id), consider aggregating

2. **Log Volume**:
   - Request/response bodies limited to 10KB
   - Logs older than 7 days are automatically deleted
   - Adjust retention in loki-config.yml if needed

3. **Storage**:
   - Prometheus retention: 30 days
   - Loki retention: 7 days (168h)
   - Monitor volume usage: `docker system df -v`

### Scaling

1. **Horizontal Scaling**:
   - Multiple app instances will all send to same Prometheus/Loki
   - Add instance labels to distinguish sources

2. **Storage**:
   - Use remote storage for Prometheus in production
   - Consider S3 for Loki in production

3. **High Availability**:
   - Run multiple Prometheus replicas
   - Use Loki distributed mode for large deployments

## Troubleshooting

### Metrics Not Appearing

1. Check application is exposing metrics:
```bash
curl http://localhost:8080/metrics
```

2. Check Prometheus targets:
- Navigate to http://localhost:9090/targets
- Ensure `codecollab-backend` target is UP

3. Check Prometheus config:
```bash
docker exec prometheus cat /etc/prometheus/prometheus.yml
```

### Logs Not Appearing

1. Check Loki is running:
```bash
docker logs loki
```

2. Test Loki directly:
```bash
curl http://localhost:3100/ready
```

3. Check application Loki connection:
- Look for "Loki logger initialized" in app logs
- Verify LOKI_URL environment variable

### Dashboard Not Loading

1. Check Grafana datasources:
- Navigate to Configuration > Data Sources
- Verify Prometheus and Loki are connected

2. Check dashboard provisioning:
```bash
docker exec grafana ls /var/lib/grafana/dashboards
```

3. Re-import dashboard:
- Copy content from grafana/dashboards/codecollab-dashboard.json
- Import via Grafana UI

### High Memory Usage

1. Reduce metric retention:
```yaml
--storage.tsdb.retention.time=15d
```

2. Reduce log retention:
```yaml
retention_period: 72h
```

3. Limit log size:
- Adjust body size limits in middleware/logging.go

## Alerting (Future Enhancement)

To add alerting:

1. Configure Alertmanager in docker-compose.yml
2. Add alert rules to Prometheus
3. Configure notification channels (email, Slack, etc.)

Example alert rule:

```yaml
groups:
  - name: codecollab
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        annotations:
          summary: "High error rate detected"
```

## Metrics Retention & Data Management

### Prometheus
- Default retention: 30 days
- Storage path: prometheus-data volume
- Estimated size: ~1GB per million samples

### Loki
- Default retention: 7 days
- Storage path: loki-data volume
- Estimated size: Varies by request volume

### Grafana
- Dashboards: grafana-data volume
- No automatic cleanup needed

## Additional Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Loki Documentation](https://grafana.com/docs/loki/)
- [Prometheus Go Client](https://github.com/prometheus/client_golang)

## Support

For issues or questions:
1. Check application logs: `docker logs codecollab-backend`
2. Check Prometheus logs: `docker logs prometheus`
3. Check Loki logs: `docker logs loki`
4. Check Grafana logs: `docker logs grafana`
