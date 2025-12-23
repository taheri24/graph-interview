# Prometheus Monitoring Setup Guide

## Overview

This document describes the complete Prometheus monitoring setup for the Task Management API. Prometheus is configured to automatically scrape metrics from your API and is pre-configured in the Docker Compose stack.

## Quick Start

### Start All Services with Prometheus

```bash
docker-compose up -d
```

This will start:
- **PostgreSQL** on port 5432
- **API** on port 8080 with metrics endpoint at `/metrics`
- **Prometheus** on port 9090

### Access Prometheus UI

Open your browser and navigate to:
```
http://localhost:9090
```

## Configuration Files

### docker-compose.yml
The main Prometheus service configuration:

```yaml
prometheus:
  image: prom/prometheus:latest
  container_name: task-api-prometheus
  ports:
    - "9090:9090"
  volumes:
    - ./prometheus.yml:/etc/prometheus/prometheus.yml
    - prometheus_data:/prometheus
  command:
    - "--config.file=/etc/prometheus/prometheus.yml"
    - "--storage.tsdb.path=/prometheus"
  networks:
    - task-network
  restart: unless-stopped
  depends_on:
    - api
```

### prometheus.yml
Scrape configuration for metrics collection:

```yaml
global:
  scrape_interval: 15s        # Default scrape interval
  evaluation_interval: 15s    # Default rule evaluation interval
  external_labels:
    monitor: 'task-api-monitor'

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'task-api'
    static_configs:
      - targets: ['api:8080']
    metrics_path: '/metrics'
    scrape_interval: 10s      # Task API metrics every 10 seconds
    scrape_timeout: 5s
```

## API Metrics Endpoint

The API exposes metrics at:
```
http://localhost:8080/metrics
```

### Metrics Details

| Metric | Type | Description |
|--------|------|-------------|
| `requests_total` | Counter | Total HTTP requests (labeled by method, path, status) |
| `request_latency_histogram_seconds` | Histogram | Request latency distribution |
| `tasks_count` | Gauge | Current number of tasks in database |

## Useful PromQL Queries

### Request Monitoring

```promql
# Total requests per second
rate(requests_total[1m])

# Requests by status code
requests_total{job="task-api"} by (status)

# HTTP error rate
rate(requests_total{status=~"5.."}[5m])

# Request count by endpoint
requests_total{job="task-api"} by (path)
```

### Performance Metrics

```promql
# 95th percentile latency (p95)
histogram_quantile(0.95, rate(request_latency_histogram_seconds_bucket[5m]))

# 99th percentile latency (p99)
histogram_quantile(0.99, rate(request_latency_histogram_seconds_bucket[5m]))

# Average request latency
avg(rate(request_latency_histogram_seconds_sum[5m])) / avg(rate(request_latency_histogram_seconds_count[5m]))
```

### Application State

```promql
# Current task count
tasks_count

# Task creation rate
rate(requests_total{path="/tasks",method="POST"}[5m])
```

## Middleware Integration

The API includes comprehensive middleware for metrics collection:

### setup.go
- `SetupGlobalMiddleware()` - Configures recovery, request ID, and metrics middleware
- `SetupMetricsEndpoint()` - Registers the `/metrics` endpoint

### metrics.go
Implements the metrics middleware:
- Tracks request latency with histogram buckets
- Counts total requests by method, path, and status
- Manages task count gauge
- Automatically excludes health and swagger endpoints

## Troubleshooting

### Prometheus not scraping metrics

1. Check if API is running:
```bash
curl http://localhost:8080/health
```

2. Check if metrics endpoint is accessible:
```bash
curl http://localhost:8080/metrics
```

3. View Prometheus targets:
Open http://localhost:9090/targets

### Metrics not appearing

1. Wait at least 10 seconds for the first scrape
2. Check Prometheus logs:
```bash
docker logs task-api-prometheus
```

3. Verify API logs:
```bash
docker logs task-api
```

### Data persistence

Metrics are stored in the `prometheus_data` volume:
```bash
# View volume details
docker volume ls | grep prometheus

# Clean up (WARNING: deletes metrics history)
docker volume rm graph-interview_prometheus_data
```

## Performance Tuning

### Adjust Scrape Interval

Edit `prometheus.yml` and change:
```yaml
global:
  scrape_interval: 30s  # Increase for lower resource usage
```

Then reload:
```bash
docker-compose restart prometheus
```

### Storage Retention

Add command flag in docker-compose.yml:
```yaml
command:
  - "--storage.tsdb.retention.time=30d"  # Keep 30 days of data
```

## Monitoring the Monitoring

Prometheus self-monitors itself. Query:
```promql
# Prometheus scrape duration
prometheus_tsdb_symbol_table_size_bytes

# Metrics processed per second
rate(prometheus_tsdb_compaction_chunk_samples[5m])
```

## Integration with Other Tools

### Grafana Integration
1. Add Prometheus as a data source
2. Configure: `URL: http://prometheus:9090`
3. Create dashboards using the metrics above

### Alerting
Create alert rules in Prometheus:
```yaml
alert_rules:
  - alert: HighErrorRate
    expr: rate(requests_total{status=~"5.."}[5m]) > 0.05
    for: 5m
```

## Next Steps

1. **Monitor in Production**: Use Prometheus with a reverse proxy (nginx)
2. **Add Grafana**: Visualize metrics with Grafana dashboards
3. **Set Alerts**: Configure alert rules for critical metrics
4. **Scale**: Use Prometheus federation for multiple instances
5. **Export**: Export metrics to long-term storage (S3, GCS, etc.)

## Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Go Prometheus Client](https://github.com/prometheus/client_golang)
- [PromQL Guide](https://prometheus.io/docs/prometheus/latest/querying/basics/)
