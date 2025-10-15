# Operations Guide

This guide provides information for operating the Bazaruto Insurance Platform in production environments.

## Monitoring

### Health Endpoints

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

### Health Check Response

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
    }
  }
}
```

### Prometheus Metrics

```bash
# Metrics endpoint
curl http://localhost:8080/metrics

# Key metrics to monitor:
# - http_requests_total
# - http_request_duration_seconds
# - database_connections_active
# - redis_connections_active
# - job_queue_size
# - job_processing_duration_seconds
```

## Logging

### Log Levels

- **DEBUG**: Detailed information for debugging
- **INFO**: General information about application flow
- **WARN**: Warning messages for potential issues
- **ERROR**: Error messages for failed operations
- **FATAL**: Critical errors that cause application shutdown

### Log Format

```json
{
  "level": "info",
  "timestamp": "2024-01-15T10:30:00Z",
  "message": "User created successfully",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "request_id": "req_123456789"
}
```

### Log Aggregation

```bash
# Using Fluentd
<source>
  @type tail
  path /var/log/bazaruto/*.log
  pos_file /var/log/fluentd/bazaruto.log.pos
  tag bazaruto
  format json
</source>

<match bazaruto>
  @type elasticsearch
  host elasticsearch.example.com
  port 9200
  index_name bazaruto-logs
</match>
```

## Alerting

### Key Alerts

1. **Application Down**
   - Alert when health check fails
   - Threshold: 2 consecutive failures
   - Severity: Critical

2. **High Error Rate**
   - Alert when error rate > 5%
   - Threshold: 5% over 5 minutes
   - Severity: High

3. **High Response Time**
   - Alert when p95 response time > 2s
   - Threshold: 2s over 5 minutes
   - Severity: Medium

4. **Database Connection Issues**
   - Alert when database connections > 80% of max
   - Threshold: 80% over 5 minutes
   - Severity: High

5. **Job Queue Backlog**
   - Alert when job queue size > 1000
   - Threshold: 1000 jobs
   - Severity: Medium

### Alert Configuration

```yaml
# Prometheus alert rules
groups:
- name: bazaruto
  rules:
  - alert: ApplicationDown
    expr: up{job="bazaruto"} == 0
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "Bazaruto application is down"
      
  - alert: HighErrorRate
    expr: rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) > 0.05
    for: 5m
    labels:
      severity: high
    annotations:
      summary: "High error rate detected"
```

## Performance Tuning

### Database Optimization

```sql
-- Connection pooling
max_connections = 100
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 4MB
maintenance_work_mem = 64MB

-- Query optimization
-- Add indexes for frequently queried columns
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_policies_user_id ON policies(user_id);
CREATE INDEX idx_claims_status ON claims(status);
```

### Redis Optimization

```bash
# Redis configuration
maxmemory 2gb
maxmemory-policy allkeys-lru
tcp-keepalive 60
timeout 300
```

### Application Tuning

```yaml
# Application configuration
server:
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s

db:
  max_connections: 25
  min_connections: 5
  max_lifetime: 1h
  idle_timeout: 30m

redis:
  max_connections: 10
  min_connections: 1
  max_lifetime: 1h
```

## Troubleshooting

### Common Issues

#### 1. Database Connection Issues

```bash
# Check database connectivity
kubectl exec -it bazarutod-pod -- nc -zv postgres-service 5432

# Check connection pool status
curl http://localhost:8080/metrics | grep database_connections

# Check database logs
kubectl logs postgres-pod
```

#### 2. Redis Connection Issues

```bash
# Check Redis connectivity
kubectl exec -it bazarutod-pod -- redis-cli -h redis-service ping

# Check Redis memory usage
kubectl exec -it bazarutod-pod -- redis-cli -h redis-service info memory
```

#### 3. High Memory Usage

```bash
# Check memory usage
kubectl top pods -n bazaruto

# Check for memory leaks
kubectl exec -it bazarutod-pod -- curl http://localhost:8080/debug/pprof/heap
```

#### 4. Job Processing Issues

```bash
# Check job queue status
kubectl exec -it bazarutod-pod -- curl http://localhost:8080/healthz/jobs

# Check failed jobs
kubectl exec -it bazarutod-pod -- curl http://localhost:8080/jobs/failed
```

### Debug Commands

```bash
# Application logs
kubectl logs -f deployment/bazarutod -n bazaruto

# Job worker logs
kubectl logs -f deployment/bazarutod-workers -n bazaruto

# Database logs
kubectl logs -f deployment/postgres -n bazaruto

# Redis logs
kubectl logs -f deployment/redis -n bazaruto
```

## Backup and Recovery

### Database Backup

```bash
#!/bin/bash
# Automated backup script

BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/bazaruto_$DATE.sql"

# Create backup
pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME > $BACKUP_FILE

# Compress backup
gzip $BACKUP_FILE

# Remove old backups (keep 30 days)
find $BACKUP_DIR -name "bazaruto_*.sql.gz" -mtime +30 -delete
```

### Configuration Backup

```bash
#!/bin/bash
# Backup business rules configuration

CONFIG_DIR="/config"
BACKUP_DIR="/backups/config"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup
cp $CONFIG_DIR/business_rules.json $BACKUP_DIR/business_rules_$DATE.json

# Remove old backups (keep 30 days)
find $BACKUP_DIR -name "business_rules_*.json" -mtime +30 -delete
```

### Recovery Procedures

#### Database Recovery

```bash
# Restore from backup
gunzip -c backup_20240115_103000.sql.gz | psql -h $DB_HOST -U $DB_USER -d $DB_NAME

# Verify restoration
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT COUNT(*) FROM users;"
```

#### Application Recovery

```bash
# Scale down application
kubectl scale deployment bazarutod --replicas=0 -n bazaruto

# Restore configuration
kubectl cp business_rules.json bazarutod-pod:/app/config/ -n bazaruto

# Scale up application
kubectl scale deployment bazarutod --replicas=3 -n bazaruto
```

## Security Operations

### SSL/TLS Management

```bash
# Renew SSL certificates
certbot renew --nginx

# Check certificate expiration
openssl x509 -in /etc/ssl/certs/bazaruto.crt -text -noout | grep "Not After"
```

### Security Monitoring

```bash
# Check for failed login attempts
grep "authentication failed" /var/log/bazaruto/access.log | tail -100

# Check for suspicious activity
grep "status=403" /var/log/bazaruto/access.log | tail -50
```

### Access Control

```bash
# Review user permissions
kubectl get rolebindings -n bazaruto

# Check service account permissions
kubectl describe serviceaccount bazarutod -n bazaruto
```

## Incident Response

### Incident Classification

- **P1 (Critical)**: Application down, data loss, security breach
- **P2 (High)**: Major functionality affected, performance degradation
- **P3 (Medium)**: Minor functionality affected, non-critical issues
- **P4 (Low)**: Cosmetic issues, documentation updates

### Response Procedures

#### P1 Incident Response

1. **Immediate Response** (0-15 minutes)
   - Acknowledge incident
   - Assess impact
   - Notify stakeholders

2. **Investigation** (15-60 minutes)
   - Gather information
   - Identify root cause
   - Implement temporary fix

3. **Resolution** (1-4 hours)
   - Implement permanent fix
   - Verify resolution
   - Monitor system stability

4. **Post-Incident** (24-48 hours)
   - Conduct post-mortem
   - Document lessons learned
   - Implement preventive measures

### Communication

```bash
# Incident notification template
Subject: [P1] Bazaruto Application Down - Incident #INC-2024-001

Incident Details:
- Severity: P1 (Critical)
- Impact: Application unavailable
- Start Time: 2024-01-15 10:30:00 UTC
- Affected Services: API, Webhooks, Job Processing

Current Status: Investigating
Next Update: 15 minutes

Incident Commander: [Name]
Technical Lead: [Name]
```

## Maintenance Windows

### Scheduled Maintenance

```bash
# Maintenance window checklist
- [ ] Notify users 24 hours in advance
- [ ] Create maintenance ticket
- [ ] Backup current configuration
- [ ] Perform maintenance tasks
- [ ] Verify system functionality
- [ ] Close maintenance ticket
- [ ] Notify users of completion
```

### Zero-Downtime Deployments

```bash
# Rolling update deployment
kubectl set image deployment/bazarutod bazarutod=bazaruto:v1.2.0 -n bazaruto

# Check rollout status
kubectl rollout status deployment/bazarutod -n bazaruto

# Rollback if needed
kubectl rollout undo deployment/bazarutod -n bazaruto
```