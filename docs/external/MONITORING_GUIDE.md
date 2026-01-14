# Monitoring Guide for Synapse Sentinel

This guide provides comprehensive information about monitoring, logging, and observability for the Synapse Sentinel system, including both the Sentinel Agent and Hub API.

## Table of Contents

1. [Health Check Endpoints](#health-check-endpoints)
2. [Metrics Collection](#metrics-collection)
3. [Structured Logging](#structured-logging)
4. [Alerting Thresholds](#alerting-thresholds)
5. [Dashboard Examples](#dashboard-examples)
6. [Troubleshooting](#troubleshooting)

## Health Check Endpoints

The Hub API provides three health check endpoints for monitoring service status:

### GET /health

Basic health check endpoint that returns service status.

**Response (200 OK):**
```json
{
  "status": "ok",
  "service": "sentinel-hub",
  "version": "1.0.0",
  "timestamp": "2024-12-10T14:30:00Z"
}
```

**Use Cases:**
- Load balancer health checks
- Basic service availability monitoring
- Uptime monitoring

### GET /health/db

Database connectivity check with connection pool statistics.

**Response (200 OK - Healthy):**
```json
{
  "status": "healthy",
  "service": "database",
  "stats": {
    "open_connections": 5,
    "in_use": 2,
    "idle": 3,
    "wait_count": 0
  },
  "timestamp": "2024-12-10T14:30:00Z"
}
```

**Response (503 Service Unavailable - Unhealthy):**
```json
{
  "status": "unhealthy",
  "service": "database",
  "error": "connection refused",
  "timestamp": "2024-12-10T14:30:00Z"
}
```

**Use Cases:**
- Database connectivity monitoring
- Connection pool health monitoring
- Database performance monitoring

**Alerting Thresholds:**
- **Warning**: `wait_count > 10` - Connection pool may be exhausted
- **Critical**: `status == "unhealthy"` - Database is unavailable

### GET /health/ready

Readiness check that verifies the service is ready to accept traffic. Checks both database connectivity and storage availability.

**Response (200 OK - Ready):**
```json
{
  "status": "ready",
  "timestamp": "2024-12-10T14:30:00Z"
}
```

**Response (503 Service Unavailable - Not Ready):**
```json
{
  "status": "not_ready",
  "reason": "database_unavailable",
  "error": "connection refused",
  "timestamp": "2024-12-10T14:30:00Z"
}
```

**Use Cases:**
- Kubernetes readiness probes
- Service mesh health checks
- Pre-deployment verification

**Alerting Thresholds:**
- **Critical**: `status == "not_ready"` - Service cannot accept traffic

## Metrics Collection

### Prometheus Metrics Endpoint

The Hub API exposes Prometheus-formatted metrics at `/api/v1/metrics/prometheus`.

**Access:** Requires API key authentication (protected endpoint)

**Example Response:**
```
# HTTP Request Metrics
sentinel_http_requests_total{endpoint="/api/v1/documents/ingest"} 1250
sentinel_http_errors_total{endpoint="/api/v1/documents/ingest"} 5
sentinel_http_request_duration_ms{endpoint="/api/v1/documents/ingest"} 245.50

# Database Connection Pool Metrics
sentinel_db_open_connections 5
sentinel_db_in_use 2
sentinel_db_idle 3
sentinel_db_wait_count 0
sentinel_db_wait_duration_ms 0

# Service Metrics
sentinel_uptime_seconds 86400.00
```

### Available Metrics

#### HTTP Request Metrics

- **`sentinel_http_requests_total{endpoint="..."}`** (Counter)
  - Total number of HTTP requests per endpoint
  - Labels: `endpoint` - The API endpoint path

- **`sentinel_http_errors_total{endpoint="..."}`** (Counter)
  - Total number of HTTP errors (status >= 400) per endpoint
  - Labels: `endpoint` - The API endpoint path

- **`sentinel_http_request_duration_ms{endpoint="..."}`** (Gauge)
  - Average request duration in milliseconds per endpoint
  - Labels: `endpoint` - The API endpoint path

#### Database Metrics

- **`sentinel_db_open_connections`** (Gauge)
  - Current number of open database connections

- **`sentinel_db_in_use`** (Gauge)
  - Number of connections currently in use

- **`sentinel_db_idle`** (Gauge)
  - Number of idle connections in the pool

- **`sentinel_db_wait_count`** (Counter)
  - Total number of connections waited for

- **`sentinel_db_wait_duration_ms`** (Gauge)
  - Total wait duration for connections in milliseconds

#### Service Metrics

- **`sentinel_uptime_seconds`** (Gauge)
  - Service uptime in seconds since last restart

### Prometheus Configuration

Add the following scrape configuration to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'sentinel-hub'
    scrape_interval: 15s
    metrics_path: '/api/v1/metrics/prometheus'
    static_configs:
      - targets: ['hub.example.com:8080']
    basic_auth:
      username: 'your-api-key'
      password: ''  # API key goes in username field
```

**Note:** The Prometheus endpoint requires API key authentication. Configure Prometheus to authenticate using basic auth with the API key as the username.

## Structured Logging

### Sentinel Agent Logging

The Sentinel Agent supports structured logging with configurable levels and formats.

#### Environment Variables

- **`SENTINEL_LOG_LEVEL`** - Log level (default: `info`)
  - Values: `debug`, `info`, `warn`, `error`
  
- **`SENTINEL_LOG_FORMAT`** - Log format (default: `text`)
  - Values: `text`, `json`
  
- **`SENTINEL_LOG_FILE`** - Log file path (default: stderr)
  - If set, logs will be written to the specified file

#### Text Format Example

```
[2024-12-10T14:30:00Z] INFO: Sentinel Agent starting version=v24 log_level=INFO log_format=text
[2024-12-10T14:30:15Z] DEBUG: Loading configuration from .sentinelsrc
[2024-12-10T14:30:20Z] INFO: Audit completed findings=5 severity=warning
[2024-12-10T14:30:25Z] ERROR: Hub API request failed endpoint=/api/v1/analyze/comprehensive error=timeout
```

#### JSON Format Example

```json
{"timestamp":"2024-12-10T14:30:00Z","level":"INFO","message":"Sentinel Agent starting","version":"v24","log_level":"INFO","log_format":"json"}
{"timestamp":"2024-12-10T14:30:15Z","level":"DEBUG","message":"Loading configuration from .sentinelsrc"}
{"timestamp":"2024-12-10T14:30:20Z","level":"INFO","message":"Audit completed","findings":5,"severity":"warning"}
{"timestamp":"2024-12-10T14:30:25Z","level":"ERROR","message":"Hub API request failed","endpoint":"/api/v1/analyze/comprehensive","error":"timeout"}
```

#### Log Levels

- **DEBUG**: Detailed diagnostic information for troubleshooting
- **INFO**: General informational messages about normal operation
- **WARN**: Warning messages for potentially problematic situations
- **ERROR**: Error messages for failures that don't stop the service

### Hub API Logging

The Hub API uses Go's standard `log` package with structured output. Logs include:
- Request/response logging via chi middleware
- Database connection pool monitoring
- Error logging with context

## Alerting Thresholds

### Recommended Alert Rules

#### Critical Alerts

1. **Service Down**
   - Condition: `/health` returns non-200 status
   - Severity: Critical
   - Action: Page on-call engineer

2. **Database Unavailable**
   - Condition: `/health/db` returns `status: "unhealthy"`
   - Severity: Critical
   - Action: Page on-call engineer

3. **Service Not Ready**
   - Condition: `/health/ready` returns non-200 status
   - Severity: Critical
   - Action: Page on-call engineer

4. **High Error Rate**
   - Condition: `rate(sentinel_http_errors_total[5m]) > 10`
   - Severity: Critical
   - Action: Page on-call engineer

#### Warning Alerts

1. **Database Connection Pool Exhaustion**
   - Condition: `sentinel_db_wait_count > 10` over 5 minutes
   - Severity: Warning
   - Action: Notify team, investigate connection pool configuration

2. **High Request Latency**
   - Condition: `sentinel_http_request_duration_ms > 1000` for any endpoint
   - Severity: Warning
   - Action: Investigate performance issues

3. **Elevated Error Rate**
   - Condition: `rate(sentinel_http_errors_total[5m]) > 5`
   - Severity: Warning
   - Action: Monitor and investigate

### Prometheus Alert Rules

```yaml
groups:
  - name: sentinel_hub
    interval: 30s
    rules:
      - alert: SentinelHubDown
        expr: up{job="sentinel-hub"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Sentinel Hub is down"
          description: "Sentinel Hub has been down for more than 1 minute"

      - alert: DatabaseUnavailable
        expr: sentinel_db_wait_count > 10
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Database connection pool exhausted"
          description: "Database wait count is {{ $value }}, indicating connection pool issues"

      - alert: HighErrorRate
        expr: rate(sentinel_http_errors_total[5m]) > 10
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }} errors/second"

      - alert: HighLatency
        expr: sentinel_http_request_duration_ms > 1000
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High request latency"
          description: "Request duration is {{ $value }}ms for endpoint {{ $labels.endpoint }}"
```

## Dashboard Examples

### Grafana Dashboard Configuration

#### Key Panels

1. **Service Health**
   - Health check status (up/down)
   - Database connectivity status
   - Service readiness status

2. **Request Metrics**
   - Request rate per endpoint
   - Error rate per endpoint
   - Request duration (p50, p95, p99)

3. **Database Metrics**
   - Connection pool utilization
   - Wait count and duration
   - Connection pool size

4. **Service Metrics**
   - Uptime
   - Request volume over time
   - Error percentage

#### Example Grafana Queries

**Request Rate:**
```
rate(sentinel_http_requests_total[5m])
```

**Error Rate:**
```
rate(sentinel_http_errors_total[5m])
```

**Average Request Duration:**
```
sentinel_http_request_duration_ms
```

**Database Connection Pool Utilization:**
```
sentinel_db_in_use / sentinel_db_open_connections * 100
```

**Error Percentage:**
```
rate(sentinel_http_errors_total[5m]) / rate(sentinel_http_requests_total[5m]) * 100
```

## Troubleshooting

### Common Issues

#### Health Check Failing

1. **Check Database Connectivity**
   ```bash
   curl http://localhost:8080/health/db
   ```
   - Verify database is running
   - Check database connection string in environment variables
   - Review database connection pool configuration

2. **Check Storage Directory**
   ```bash
   ls -la $DOCUMENT_STORAGE
   ```
   - Verify `DOCUMENT_STORAGE` environment variable is set
   - Ensure storage directory exists and is writable

#### High Error Rate

1. **Review Logs**
   ```bash
   # For Sentinel Agent
   export SENTINEL_LOG_LEVEL=debug
   ./sentinel audit
   
   # For Hub API
   # Check application logs for error details
   ```

2. **Check Metrics**
   ```bash
   curl -H "Authorization: Bearer $API_KEY" \
     http://localhost:8080/api/v1/metrics/prometheus | grep errors
   ```

3. **Database Connection Pool Issues**
   - Check `sentinel_db_wait_count` metric
   - Review connection pool configuration
   - Consider increasing `MaxOpenConns` if needed

#### Performance Issues

1. **Monitor Request Duration**
   - Check `sentinel_http_request_duration_ms` metrics
   - Identify slow endpoints
   - Review database query performance

2. **Database Performance**
   - Check connection pool utilization
   - Review slow query logs
   - Optimize database indexes

### Log Analysis

#### Filtering Logs

**Text Format:**
```bash
grep "ERROR" sentinel.log
grep "endpoint=/api/v1" sentinel.log
```

**JSON Format:**
```bash
jq 'select(.level == "ERROR")' sentinel.log
jq 'select(.endpoint == "/api/v1/analyze/comprehensive")' sentinel.log
```

#### Common Log Patterns

- **Timeout Errors**: Look for `error=timeout` in logs
- **Database Errors**: Look for `database` or `connection` in error messages
- **Authentication Errors**: Look for `401` or `403` status codes

## Best Practices

1. **Monitor Health Checks Regularly**
   - Set up automated health check monitoring
   - Configure alerts for health check failures
   - Use health checks for load balancer configuration

2. **Collect Metrics Continuously**
   - Set up Prometheus scraping
   - Retain metrics for at least 30 days
   - Set up dashboards for key metrics

3. **Use Structured Logging**
   - Use JSON format for log aggregation systems
   - Include relevant context in log fields
   - Set appropriate log levels

4. **Set Up Alerting**
   - Configure alerts for critical issues
   - Test alerting rules regularly
   - Document alert response procedures

5. **Regular Review**
   - Review metrics trends weekly
   - Analyze error patterns
   - Optimize based on metrics data

## Additional Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Security Guide](./SECURITY_GUIDE.md)
- [Deployment Guide](./DEPLOYMENT_GUIDE.md)










