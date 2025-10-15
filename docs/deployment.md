# Deployment Guide

This guide covers deploying the Bazaruto Insurance Platform in various environments.

## Prerequisites

- Go 1.22+ (for building from source)
- Docker and Docker Compose (for containerized deployment)
- Kubernetes cluster (for production deployment)
- PostgreSQL 14+
- Redis 6+

## Local Development

### Using Docker Compose

1. **Clone the repository**
   ```bash
   git clone https://github.com/edsonmichaque/bazaruto.git
   cd bazaruto
   ```

2. **Start dependencies**
   ```bash
   docker-compose up -d postgres redis
   ```

3. **Run migrations**
   ```bash
   make migrate
   # or
   go run cmd/bazarutod/main.go migrate
   ```

4. **Start the application**
   ```bash
   make run
   # or
   go run cmd/bazarutod/main.go serve
   ```

### Using Local Services

1. **Install PostgreSQL and Redis locally**

2. **Create database**
   ```sql
   CREATE DATABASE bazaruto;
   ```

3. **Configure environment**
   ```bash
   export BAZARUTO_DB_HOST=localhost
   export BAZARUTO_DB_NAME=bazaruto
   export BAZARUTO_DB_USER=postgres
   export BAZARUTO_DB_PASSWORD=password
   export BAZARUTO_REDIS_ADDRESS=localhost:6379
   ```

4. **Run the application**
   ```bash
   go run cmd/bazarutod/main.go serve
   ```

## Docker Deployment

### Building Docker Image

```bash
# Build the image
docker build -t bazaruto:latest .

# Build with specific tag
docker build -t bazaruto:v1.0.0 .
```

### Docker Compose Deployment

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - BAZARUTO_DB_HOST=postgres
      - BAZARUTO_DB_NAME=bazaruto
      - BAZARUTO_DB_USER=postgres
      - BAZARUTO_DB_PASSWORD=password
      - BAZARUTO_REDIS_ADDRESS=redis:6379
    depends_on:
      - postgres
      - redis
    restart: unless-stopped

  postgres:
    image: postgres:14
    environment:
      - POSTGRES_DB=bazaruto
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    restart: unless-stopped

  redis:
    image: redis:6-alpine
    ports:
      - "6379:6379"
    restart: unless-stopped

volumes:
  postgres_data:
```

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop services
docker-compose down
```

### Production Docker Compose

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  app:
    image: bazaruto:latest
    ports:
      - "8080:8080"
    environment:
      - BAZARUTO_DB_HOST=${DB_HOST}
      - BAZARUTO_DB_NAME=${DB_NAME}
      - BAZARUTO_DB_USER=${DB_USER}
      - BAZARUTO_DB_PASSWORD=${DB_PASSWORD}
      - BAZARUTO_REDIS_ADDRESS=${REDIS_ADDRESS}
      - BAZARUTO_REDIS_PASSWORD=${REDIS_PASSWORD}
      - BAZARUTO_AUTH_JWT_SECRET=${JWT_SECRET}
      - BAZARUTO_LOG_LEVEL=info
      - BAZARUTO_METRICS_ENABLED=true
      - BAZARUTO_TRACING_ENABLED=true
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/healthz"]
      interval: 30s
      timeout: 10s
      retries: 3
    deploy:
      replicas: 3
      resources:
        limits:
          memory: 512M
          cpus: '0.5'
        reservations:
          memory: 256M
          cpus: '0.25'
```

## Kubernetes Deployment

### Namespace

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: bazaruto
```

### ConfigMap

```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: bazaruto-config
  namespace: bazaruto
data:
  config.yaml: |
    server:
      addr: ":8080"
      read_timeout: 30s
      write_timeout: 30s
    
    db:
      host: postgres-service
      port: 5432
      name: bazaruto
      user: postgres
      ssl_mode: require
    
    redis:
      address: redis-service:6379
    
    log_level: info
    log_format: json
    metrics_enabled: true
    tracing:
      enabled: true
      service_name: bazaruto
      endpoint: http://jaeger:14268/api/traces
```

### Secrets

```yaml
# k8s/secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: bazaruto-secrets
  namespace: bazaruto
type: Opaque
data:
  db-password: <base64-encoded-password>
  redis-password: <base64-encoded-password>
  jwt-secret: <base64-encoded-secret>
  email-password: <base64-encoded-password>
  payment-secret: <base64-encoded-secret>
```

### Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bazaruto
  namespace: bazaruto
spec:
  replicas: 3
  selector:
    matchLabels:
      app: bazaruto
  template:
    metadata:
      labels:
        app: bazaruto
    spec:
      containers:
      - name: bazaruto
        image: bazaruto:latest
        ports:
        - containerPort: 8080
        env:
        - name: BAZARUTO_DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: bazaruto-secrets
              key: db-password
        - name: BAZARUTO_REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: bazaruto-secrets
              key: redis-password
        - name: BAZARUTO_AUTH_JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: bazaruto-secrets
              key: jwt-secret
        volumeMounts:
        - name: config
          mountPath: /app/config.yaml
          subPath: config.yaml
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      volumes:
      - name: config
        configMap:
          name: bazaruto-config
```

### Service

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: bazaruto-service
  namespace: bazaruto
spec:
  selector:
    app: bazaruto
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
```

### Ingress

```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: bazaruto-ingress
  namespace: bazaruto
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - api.bazaruto.com
    secretName: bazaruto-tls
  rules:
  - host: api.bazaruto.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: bazaruto-service
            port:
              number: 80
```

### Horizontal Pod Autoscaler

```yaml
# k8s/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: bazaruto-hpa
  namespace: bazaruto
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: bazaruto
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### Deploy to Kubernetes

```bash
# Apply all manifests
kubectl apply -f k8s/

# Check deployment status
kubectl get pods -n bazaruto

# View logs
kubectl logs -f deployment/bazaruto -n bazaruto

# Scale deployment
kubectl scale deployment bazaruto --replicas=5 -n bazaruto
```

## Cloud Deployment

### AWS EKS

1. **Create EKS cluster**
   ```bash
   eksctl create cluster --name bazaruto-cluster --region us-west-2
   ```

2. **Deploy application**
   ```bash
   kubectl apply -f k8s/
   ```

3. **Configure load balancer**
   ```bash
   kubectl apply -f k8s/aws-load-balancer.yaml
   ```

### Google GKE

1. **Create GKE cluster**
   ```bash
   gcloud container clusters create bazaruto-cluster --zone us-central1-a
   ```

2. **Deploy application**
   ```bash
   kubectl apply -f k8s/
   ```

3. **Configure ingress**
   ```bash
   kubectl apply -f k8s/gke-ingress.yaml
   ```

### Azure AKS

1. **Create AKS cluster**
   ```bash
   az aks create --resource-group bazaruto-rg --name bazaruto-cluster
   ```

2. **Deploy application**
   ```bash
   kubectl apply -f k8s/
   ```

## Database Setup

### PostgreSQL

1. **Create database**
   ```sql
   CREATE DATABASE bazaruto;
   CREATE USER bazaruto_user WITH PASSWORD 'secure_password';
   GRANT ALL PRIVILEGES ON DATABASE bazaruto TO bazaruto_user;
   ```

2. **Run migrations**
   ```bash
   go run cmd/bazarutod/main.go migrate
   ```

3. **Set up replication (production)**
   ```sql
   -- Master configuration
   ALTER SYSTEM SET wal_level = replica;
   ALTER SYSTEM SET max_wal_senders = 3;
   ALTER SYSTEM SET max_replication_slots = 3;
   
   -- Create replication user
   CREATE USER replicator WITH REPLICATION ENCRYPTED PASSWORD 'replication_password';
   ```

### Redis

1. **Basic setup**
   ```bash
   # Start Redis
   redis-server
   
   # Test connection
   redis-cli ping
   ```

2. **Production setup with clustering**
   ```bash
   # Redis cluster configuration
   redis-server --port 7000 --cluster-enabled yes --cluster-config-file nodes-7000.conf
   redis-server --port 7001 --cluster-enabled yes --cluster-config-file nodes-7001.conf
   redis-server --port 7002 --cluster-enabled yes --cluster-config-file nodes-7002.conf
   ```

## Monitoring Setup

### Prometheus

```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
- job_name: 'bazaruto'
  static_configs:
  - targets: ['bazaruto-service:80']
  metrics_path: /metrics
  scrape_interval: 5s
```

### Grafana

```yaml
# monitoring/grafana-dashboard.json
{
  "dashboard": {
    "title": "Bazaruto Platform",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])"
          }
        ]
      }
    ]
  }
}
```

### Jaeger

```yaml
# monitoring/jaeger.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: jaeger
spec:
  replicas: 1
  selector:
    matchLabels:
      app: jaeger
  template:
    metadata:
      labels:
        app: jaeger
    spec:
      containers:
      - name: jaeger
        image: jaegertracing/all-in-one:latest
        ports:
        - containerPort: 16686
        - containerPort: 14268
        env:
        - name: COLLECTOR_OTLP_ENABLED
          value: "true"
```

## Security Considerations

### Network Security

1. **Firewall rules**
   ```bash
   # Allow only necessary ports
   ufw allow 22/tcp    # SSH
   ufw allow 80/tcp    # HTTP
   ufw allow 443/tcp   # HTTPS
   ufw deny 5432/tcp   # PostgreSQL (internal only)
   ufw deny 6379/tcp   # Redis (internal only)
   ```

2. **TLS/SSL certificates**
   ```bash
   # Using Let's Encrypt
   certbot --nginx -d api.bazaruto.com
   ```

### Application Security

1. **Environment variables**
   ```bash
   # Use secrets management
   export BAZARUTO_AUTH_JWT_SECRET=$(aws secretsmanager get-secret-value --secret-id bazaruto/jwt-secret --query SecretString --output text)
   ```

2. **Database security**
   ```sql
   -- Create read-only user for monitoring
   CREATE USER bazaruto_monitor WITH PASSWORD 'monitor_password';
   GRANT SELECT ON ALL TABLES IN SCHEMA public TO bazaruto_monitor;
   ```

## Backup and Recovery

### Database Backup

```bash
# Create backup
pg_dump -h localhost -U postgres bazaruto > backup_$(date +%Y%m%d_%H%M%S).sql

# Restore backup
psql -h localhost -U postgres bazaruto < backup_20240101_120000.sql
```

### Automated Backups

```bash
#!/bin/bash
# backup.sh
DATE=$(date +%Y%m%d_%H%M%S)
pg_dump -h $DB_HOST -U $DB_USER $DB_NAME | gzip > /backups/backup_$DATE.sql.gz
aws s3 cp /backups/backup_$DATE.sql.gz s3://bazaruto-backups/
```

## Troubleshooting

### Common Issues

1. **Database connection issues**
   ```bash
   # Check database connectivity
   telnet $DB_HOST $DB_PORT
   
   # Check database logs
   docker logs postgres
   ```

2. **Redis connection issues**
   ```bash
   # Check Redis connectivity
   redis-cli -h $REDIS_HOST -p $REDIS_PORT ping
   
   # Check Redis logs
   docker logs redis
   ```

3. **Application startup issues**
   ```bash
   # Check application logs
   kubectl logs -f deployment/bazaruto -n bazaruto
   
   # Check configuration
   kubectl describe configmap bazaruto-config -n bazaruto
   ```

### Performance Issues

1. **High memory usage**
   ```bash
   # Check memory usage
   kubectl top pods -n bazaruto
   
   # Adjust resource limits
   kubectl patch deployment bazaruto -n bazaruto -p '{"spec":{"template":{"spec":{"containers":[{"name":"bazaruto","resources":{"limits":{"memory":"1Gi"}}}]}}}}'
   ```

2. **Database performance**
   ```sql
   -- Check slow queries
   SELECT query, mean_time, calls 
   FROM pg_stat_statements 
   ORDER BY mean_time DESC 
   LIMIT 10;
   ```

For more detailed troubleshooting, refer to the [Operations Guide](ops-guide.md).


