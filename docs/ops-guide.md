# Operations Guide

This guide provides comprehensive information for operating the Bazaruto Insurance Platform in production environments.

## Table of Contents

- [Monitoring](#monitoring)
- [Logging](#logging)
- [Metrics](#metrics)
- [Alerting](#alerting)
- [Health Checks](#health-checks)
- [Performance Tuning](#performance-tuning)
- [Troubleshooting](#troubleshooting)
- [Backup and Recovery](#backup-and-recovery)
- [Security Operations](#security-operations)
- [Incident Response](#incident-response)

## Monitoring

### Application Monitoring

#### Health Endpoints

```bash
# Basic health check
curl http://localhost:8080/healthz

# Detailed health check
curl http://localhost:8080/healthz/detailed

# Readiness check
curl http://localhost:8080/readyz

# Liveness check
curl http://localhost:8080/livez
```

#### Health Check Response

```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.1.0",
  "checks": {
    "database": {
      "status": "healthy",
      "response_time": "2ms"
    },
    "redis": {
      "status": "healthy",
      "response_time": "1ms"
    },
    "jobs": {
      "status": "healthy",
      "pending_jobs": 5,
      "failed_jobs": 0
    }
  }
}
```

### Infrastructure Monitoring

#### System Metrics

Monitor these system metrics:

- **CPU Usage**: < 70% average
- **Memory Usage**: < 80% average
- **Disk Usage**: < 85% average
- **Network I/O**: Monitor for anomalies
- **Load Average**: < number of CPU cores

#### Database Metrics

- **Connection Pool**: Monitor active/idle connections
- **Query Performance**: Slow query detection
- **Lock Contention**: Monitor for deadlocks
- **Replication Lag**: For read replicas

#### Redis Metrics

- **Memory Usage**: Monitor Redis memory consumption
- **Hit Rate**: Cache hit ratio > 90%
- **Connection Count**: Monitor active connections
- **Key Expiration**: Monitor key TTL patterns

## Logging

### Log Levels

Configure appropriate log levels:

```yaml
# Production
log_level: info

# Development
log_level: debug

# Troubleshooting
log_level: debug
```

### Log Format

```json
{
  "timestamp": "2024-01-15T10:30:00.123Z",
  "level": "info",
  "message": "User created successfully",
  "service": "bazaruto",
  "version": "1.1.0",
  "trace_id": "abc123def456",
  "span_id": "def456ghi789",
  "user_id": "user-123",
  "email": "user@example.com"
}
```

### Log Aggregation

#### ELK Stack Setup

```yaml
# docker-compose.logging.yml
version: '3.8'

services:
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.11.0
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
    ports:
      - "9200:9200"

  logstash:
    image: docker.elastic.co/logstash/logstash:8.11.0
    volumes:
      - ./logstash.conf:/usr/share/logstash/pipeline/logstash.conf
    ports:
      - "5044:5044"

  kibana:
    image: docker.elastic.co/kibana/kibana:8.11.0
    ports:
      - "5601:5601"
    environment:
      - ELASTICSEARCH_HOSTS=http://elasticsearch:9200
```

#### Logstash Configuration

```ruby
# logstash.conf
input {
  beats {
    port => 5044
  }
}

filter {
  if [fields][service] == "bazaruto" {
    json {
      source => "message"
    }
    
    date {
      match => [ "timestamp", "ISO8601" ]
    }
  }
}

output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "bazaruto-%{+YYYY.MM.dd}"
  }
}
```

### Log Analysis

#### Common Log Queries

```bash
# Find errors in the last hour
grep "level.*error" /var/log/bazaruto/app.log | tail -100

# Find slow database queries
grep "slow query" /var/log/bazaruto/app.log

# Find authentication failures
grep "authentication failed" /var/log/bazaruto/app.log

# Find job failures
grep "job.*failed" /var/log/bazaruto/app.log
```

#### Kibana Queries

```json
// Find errors by service
{
  "query": {
    "bool": {
      "must": [
        { "term": { "level": "error" } },
        { "term": { "service": "bazaruto" } }
      ]
    }
  }
}

// Find slow requests
{
  "query": {
    "range": {
      "response_time": {
        "gte": 1000
      }
    }
  }
}
```

## Metrics

### Prometheus Metrics

#### Application Metrics

```go
// Custom business metrics
var (
    userRegistrations = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "bazaruto_user_registrations_total",
            Help: "Total number of user registrations",
        },
        []string{"source"},
    )
    
    quoteCalculations = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "bazaruto_quote_calculation_duration_seconds",
            Help: "Time spent calculating quotes",
            Buckets: prometheus.DefBuckets,
        },
        []string{"product_type"},
    )
)
```

#### Key Metrics to Monitor

- **Request Rate**: `rate(http_requests_total[5m])`
- **Response Time**: `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))`
- **Error Rate**: `rate(http_requests_total{status=~"5.."}[5m])`
- **Database Connections**: `bazaruto_db_connections_active`
- **Job Queue Size**: `bazaruto_jobs_pending`
- **Cache Hit Rate**: `rate(redis_keyspace_hits_total[5m]) / rate(redis_keyspace_hits_total[5m] + redis_keyspace_misses_total[5m])`

### Grafana Dashboards

#### Application Dashboard

```json
{
  "dashboard": {
    "title": "Bazaruto Application",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{method}} {{path}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total{status=~\"5..\"}[5m])",
            "legendFormat": "5xx errors"
          }
        ]
      }
    ]
  }
}
```

## Alerting

### Alert Rules

#### Critical Alerts

```yaml
# alerts/critical.yml
groups:
- name: critical
  rules:
  - alert: HighErrorRate
    expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "High error rate detected"
      description: "Error rate is {{ $value }} errors per second"

  - alert: DatabaseDown
    expr: up{job="postgres"} == 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "Database is down"
      description: "PostgreSQL database is not responding"

  - alert: HighMemoryUsage
    expr: (node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / node_memory_MemTotal_bytes > 0.9
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "High memory usage"
      description: "Memory usage is {{ $value | humanizePercentage }}"
```

#### Warning Alerts

```yaml
# alerts/warning.yml
groups:
- name: warning
  rules:
  - alert: HighCPUUsage
    expr: 100 - (avg by(instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80
    for: 10m
    labels:
      severity: warning
    annotations:
      summary: "High CPU usage"
      description: "CPU usage is {{ $value }}%"

  - alert: SlowResponseTime
    expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 2
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Slow response time"
      description: "95th percentile response time is {{ $value }}s"
```

### Alert Channels

#### Slack Integration

```yaml
# alertmanager.yml
global:
  slack_api_url: 'https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'slack-notifications'

receivers:
- name: 'slack-notifications'
  slack_configs:
  - channel: '#alerts'
    title: 'Bazaruto Alert'
    text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
```

#### Email Integration

```yaml
receivers:
- name: 'email-notifications'
  email_configs:
  - to: 'ops@bazaruto.com'
    from: 'alerts@bazaruto.com'
    subject: 'Bazaruto Alert: {{ .GroupLabels.alertname }}'
    body: |
      {{ range .Alerts }}
      Alert: {{ .Annotations.summary }}
      Description: {{ .Annotations.description }}
      {{ end }}
```

## Health Checks

### Application Health Checks

```go
// Health check implementation
func (h *HealthHandler) CheckHealth(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    checks := map[string]interface{}{
        "database": h.checkDatabase(ctx),
        "redis":    h.checkRedis(ctx),
        "jobs":     h.checkJobs(ctx),
    }
    
    status := "healthy"
    for _, check := range checks {
        if check.Status != "healthy" {
            status = "unhealthy"
            break
        }
    }
    
    response := HealthResponse{
        Status:    status,
        Timestamp: time.Now(),
        Version:   version.Version,
        Checks:    checks,
    }
    
    w.Header().Set("Content-Type", "application/json")
    if status == "unhealthy" {
        w.WriteHeader(http.StatusServiceUnavailable)
    }
    json.NewEncoder(w).Encode(response)
}
```

### Kubernetes Health Checks

```yaml
# k8s/health-checks.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: health-check-config
data:
  health-check.sh: |
    #!/bin/bash
    curl -f http://localhost:8080/healthz || exit 1
```

## Performance Tuning

### Database Optimization

#### Connection Pool Tuning

```yaml
# config.yaml
db:
  max_connections: 50
  min_connections: 10
  max_lifetime: 1h
  idle_timeout: 30m
  acquire_timeout: 30s
```

#### Query Optimization

```sql
-- Add indexes for frequently queried columns
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_policies_user_id ON policies(user_id);
CREATE INDEX idx_payments_status ON payments(status);

-- Monitor slow queries
SELECT query, mean_time, calls 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;
```

### Application Optimization

#### Memory Optimization

```go
// Use object pools for frequently allocated objects
var userPool = sync.Pool{
    New: func() interface{} {
        return &models.User{}
    },
}

func (s *UserService) CreateUser(ctx context.Context, userData *CreateUserRequest) (*models.User, error) {
    user := userPool.Get().(*models.User)
    defer userPool.Put(user)
    
    // Use user object
    return user, nil
}
```

#### Caching Strategy

```go
// Implement caching for frequently accessed data
func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
    // Try cache first
    if user, found := s.cache.Get(id.String()); found {
        return user.(*models.User), nil
    }
    
    // Fetch from database
    user, err := s.store.GetUser(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Cache for 5 minutes
    s.cache.Set(id.String(), user, 5*time.Minute)
    return user, nil
}
```

## Troubleshooting

### Common Issues

#### 1. High Memory Usage

**Symptoms:**
- Memory usage > 80%
- OOM kills
- Slow response times

**Diagnosis:**
```bash
# Check memory usage
free -h
ps aux --sort=-%mem | head -10

# Check for memory leaks
go tool pprof http://localhost:8080/debug/pprof/heap
```

**Solutions:**
- Increase memory limits
- Optimize code for memory usage
- Implement object pooling
- Check for goroutine leaks

#### 2. Database Connection Issues

**Symptoms:**
- "too many connections" errors
- Slow database queries
- Connection timeouts

**Diagnosis:**
```sql
-- Check active connections
SELECT count(*) FROM pg_stat_activity;

-- Check connection pool status
SELECT * FROM pg_stat_bgwriter;
```

**Solutions:**
- Tune connection pool settings
- Add connection pooling (PgBouncer)
- Optimize queries
- Scale database horizontally

#### 3. Job Queue Backlog

**Symptoms:**
- High number of pending jobs
- Slow job processing
- Job failures

**Diagnosis:**
```bash
# Check job queue status
bazarutod jobs stats --queue mailers

# Check worker status
bazarutod queues list
```

**Solutions:**
- Scale workers horizontally
- Optimize job processing
- Implement job prioritization
- Add more queue backends

### Debugging Tools

#### Application Debugging

```bash
# Enable debug logging
export BAZARUTO_LOG_LEVEL=debug

# Use pprof for profiling
go tool pprof http://localhost:8080/debug/pprof/profile
go tool pprof http://localhost:8080/debug/pprof/heap
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

#### Database Debugging

```sql
-- Enable query logging
ALTER SYSTEM SET log_statement = 'all';
ALTER SYSTEM SET log_min_duration_statement = 1000;
SELECT pg_reload_conf();

-- Check for locks
SELECT * FROM pg_locks WHERE NOT granted;
```

## Backup and Recovery

### Database Backup

#### Automated Backup Script

```bash
#!/bin/bash
# backup.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backups"
DB_NAME="bazaruto"

# Create backup
pg_dump -h $DB_HOST -U $DB_USER $DB_NAME | gzip > $BACKUP_DIR/backup_$DATE.sql.gz

# Upload to S3
aws s3 cp $BACKUP_DIR/backup_$DATE.sql.gz s3://bazaruto-backups/

# Cleanup old backups (keep 30 days)
find $BACKUP_DIR -name "backup_*.sql.gz" -mtime +30 -delete
```

#### Point-in-Time Recovery

```bash
# Restore from backup
gunzip -c backup_20240115_120000.sql.gz | psql -h $DB_HOST -U $DB_USER $DB_NAME

# Point-in-time recovery
pg_basebackup -h $DB_HOST -U $DB_USER -D /restore -Ft -z -P
```

### Application Backup

#### Configuration Backup

```bash
# Backup configuration
tar -czf config_backup_$(date +%Y%m%d).tar.gz config.yaml secrets/

# Backup application data
tar -czf app_data_backup_$(date +%Y%m%d).tar.gz /var/lib/bazaruto/
```

## Security Operations

### Security Monitoring

#### Audit Logging

```go
// Audit log implementation
func (s *UserService) CreateUser(ctx context.Context, user *models.User) error {
    // Create user
    err := s.store.CreateUser(ctx, user)
    if err != nil {
        // Log failed attempt
        s.auditLogger.Log(ctx, "user_creation_failed", map[string]interface{}{
            "user_id": user.ID,
            "email":   user.Email,
            "error":   err.Error(),
        })
        return err
    }
    
    // Log successful creation
    s.auditLogger.Log(ctx, "user_created", map[string]interface{}{
        "user_id": user.ID,
        "email":   user.Email,
    })
    
    return nil
}
```

#### Security Alerts

```yaml
# security-alerts.yml
groups:
- name: security
  rules:
  - alert: MultipleFailedLogins
    expr: increase(bazaruto_auth_failed_logins_total[5m]) > 10
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "Multiple failed login attempts"
      description: "{{ $value }} failed login attempts in the last 5 minutes"

  - alert: SuspiciousActivity
    expr: increase(bazaruto_suspicious_requests_total[1h]) > 100
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Suspicious activity detected"
      description: "{{ $value }} suspicious requests in the last hour"
```

### Vulnerability Management

#### Dependency Scanning

```bash
# Scan for vulnerabilities
go list -json -deps ./... | nancy sleuth

# Update dependencies
go get -u ./...
go mod tidy
```

#### Container Scanning

```bash
# Scan Docker image
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
  aquasec/trivy image bazaruto:latest
```

## Incident Response

### Incident Classification

#### Severity Levels

- **P1 (Critical)**: Service completely down
- **P2 (High)**: Major functionality affected
- **P3 (Medium)**: Minor functionality affected
- **P4 (Low)**: Cosmetic issues

### Incident Response Process

#### 1. Detection and Alerting

```bash
# Check service status
curl -f http://localhost:8080/healthz

# Check logs
tail -f /var/log/bazaruto/app.log

# Check metrics
curl http://localhost:8080/metrics
```

#### 2. Initial Response

```bash
# Scale up if needed
kubectl scale deployment bazaruto --replicas=5

# Restart service if needed
kubectl rollout restart deployment/bazaruto

# Check resource usage
kubectl top pods -n bazaruto
```

#### 3. Investigation

```bash
# Check application logs
kubectl logs -f deployment/bazaruto -n bazaruto

# Check system logs
journalctl -u bazaruto -f

# Check database status
kubectl exec -it postgres-0 -- psql -U postgres -c "SELECT * FROM pg_stat_activity;"
```

#### 4. Resolution

```bash
# Apply hotfix if needed
kubectl apply -f hotfix.yaml

# Verify fix
curl -f http://localhost:8080/healthz

# Monitor for stability
watch -n 5 'curl -s http://localhost:8080/healthz | jq .status'
```

### Post-Incident Review

#### Incident Report Template

```markdown
# Incident Report: [Incident ID]

## Summary
Brief description of the incident

## Timeline
- [Time] - Incident detected
- [Time] - Investigation started
- [Time] - Root cause identified
- [Time] - Resolution implemented
- [Time] - Service restored

## Root Cause
Detailed analysis of the root cause

## Impact
- Users affected: [number]
- Downtime: [duration]
- Business impact: [description]

## Resolution
Steps taken to resolve the incident

## Prevention
Actions to prevent similar incidents

## Lessons Learned
Key takeaways and improvements
```

This operations guide provides comprehensive information for maintaining and troubleshooting the Bazaruto Insurance Platform in production environments.


