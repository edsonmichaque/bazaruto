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

1. **Clone and setup**
   ```bash
   git clone https://github.com/edsonmichaque/bazaruto-insurance.git
   cd bazaruto-insurance
   docker-compose up -d postgres redis
   ```

2. **Run migrations and start**
   ```bash
   go run cmd/bazarutod/main.go migrate
   go run cmd/bazarutod/main.go serve
   ```

### Using Local Services

1. **Install PostgreSQL and Redis locally**
2. **Create database**
   ```sql
   CREATE DATABASE bazaruto;
   ```
3. **Configure and run**
   ```bash
   cp config.yaml.example config.yaml
   # Edit config.yaml with your settings
   go run cmd/bazarutod/main.go migrate
   go run cmd/bazarutod/main.go serve
   ```

## Docker Deployment

### Build and Run

```bash
# Build Docker image
docker build -t bazaruto:latest .

# Run with Docker Compose
docker-compose up -d

# View logs
docker-compose logs -f bazarutod
```

### Docker Compose Configuration

```yaml
version: '3.8'
services:
  bazarutod:
    build: .
    ports:
      - "8080:8080"
    environment:
      - BAZARUTO_DB_HOST=postgres
      - BAZARUTO_REDIS_ADDRESS=redis:6379
    depends_on:
      - postgres
      - redis
    volumes:
      - ./config:/app/config

  postgres:
    image: postgres:14
    environment:
      POSTGRES_DB: bazaruto
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  redis:
    image: redis:6-alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:
```

## Kubernetes Deployment

### Prerequisites

- Kubernetes cluster (1.20+)
- kubectl configured
- Helm (optional)

### Deploy with kubectl

```bash
# Create namespace
kubectl create namespace bazaruto

# Apply configurations
kubectl apply -f deploy/kubernetes/namespace.yaml
kubectl apply -f deploy/kubernetes/configmap.yaml
kubectl apply -f deploy/kubernetes/deployment.yaml
kubectl apply -f deploy/kubernetes/service.yaml
kubectl apply -f deploy/kubernetes/ingress.yaml

# Check deployment
kubectl get pods -n bazaruto
kubectl get services -n bazaruto
```

### Kubernetes Manifests

#### Namespace
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: bazaruto
```

#### ConfigMap
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: bazaruto-config
  namespace: bazaruto
data:
  config.yaml: |
    server:
      addr: ":8080"
    db:
      host: postgres-service
      port: 5432
      name: bazaruto
    redis:
      address: redis-service:6379
```

#### Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bazarutod
  namespace: bazaruto
spec:
  replicas: 3
  selector:
    matchLabels:
      app: bazarutod
  template:
    metadata:
      labels:
        app: bazarutod
    spec:
      containers:
      - name: bazarutod
        image: bazaruto:latest
        ports:
        - containerPort: 8080
        env:
        - name: BAZARUTO_DB_HOST
          value: "postgres-service"
        - name: BAZARUTO_REDIS_ADDRESS
          value: "redis-service:6379"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

#### Service
```yaml
apiVersion: v1
kind: Service
metadata:
  name: bazarutod-service
  namespace: bazaruto
spec:
  selector:
    app: bazarutod
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
```

#### Ingress
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: bazarutod-ingress
  namespace: bazaruto
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: api.bazaruto.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: bazarutod-service
            port:
              number: 80
```

## Production Deployment

### Environment Setup

1. **Database Setup**
   ```bash
   # Use managed PostgreSQL service (AWS RDS, Google Cloud SQL, etc.)
   # Configure connection pooling
   # Set up read replicas for scaling
   ```

2. **Redis Setup**
   ```bash
   # Use managed Redis service (AWS ElastiCache, Google Cloud Memorystore, etc.)
   # Configure clustering for high availability
   ```

3. **Load Balancer**
   ```bash
   # Configure load balancer (AWS ALB, Google Cloud Load Balancer, etc.)
   # Set up SSL/TLS termination
   # Configure health checks
   ```

### Configuration

1. **Environment Variables**
   ```bash
   export BAZARUTO_SERVER_ADDR=":8080"
   export BAZARUTO_DB_HOST="production-db.example.com"
   export BAZARUTO_DB_PASSWORD="secure-password"
   export BAZARUTO_REDIS_ADDRESS="production-redis.example.com:6379"
   export BAZARUTO_JWT_SECRET="very-secure-secret"
   export BAZARUTO_LOG_LEVEL="info"
   ```

2. **Business Rules Configuration**
   ```bash
   # Create environment-specific business rules
   cp config/business_rules.json config/production.json
   # Edit production.json with production values
   ```

### Monitoring Setup

1. **Prometheus Metrics**
   ```yaml
   # Add to deployment
   - name: metrics
     containerPort: 9090
   ```

2. **Logging**
   ```bash
   # Configure log aggregation (ELK stack, Fluentd, etc.)
   # Set up log rotation
   # Configure log levels
   ```

3. **Tracing**
   ```bash
   # Set up Jaeger or similar tracing system
   # Configure sampling rates
   # Set up alerting
   ```

### Security Considerations

1. **Network Security**
   - Use private subnets for databases
   - Configure security groups/firewalls
   - Enable VPC endpoints

2. **Secrets Management**
   - Use Kubernetes secrets or external secret management
   - Rotate secrets regularly
   - Encrypt secrets at rest

3. **SSL/TLS**
   - Use Let's Encrypt or similar for certificates
   - Configure HSTS headers
   - Use strong cipher suites

## Scaling

### Horizontal Scaling

1. **Application Scaling**
   ```bash
   # Scale deployment
   kubectl scale deployment bazarutod --replicas=5 -n bazaruto
   ```

2. **Database Scaling**
   - Set up read replicas
   - Configure connection pooling
   - Use database sharding if needed

3. **Job Processing Scaling**
   ```bash
   # Scale job workers
   kubectl scale deployment bazarutod-workers --replicas=10 -n bazaruto
   ```

### Vertical Scaling

1. **Resource Limits**
   ```yaml
   resources:
     requests:
       memory: "512Mi"
       cpu: "500m"
     limits:
       memory: "1Gi"
       cpu: "1000m"
   ```

## Backup and Recovery

### Database Backup

```bash
# Automated backup script
#!/bin/bash
pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME > backup_$(date +%Y%m%d_%H%M%S).sql
```

### Configuration Backup

```bash
# Backup business rules configuration
cp config/production.json backups/production_$(date +%Y%m%d_%H%M%S).json
```

### Disaster Recovery

1. **RTO/RPO Requirements**
   - Recovery Time Objective: 4 hours
   - Recovery Point Objective: 1 hour

2. **Recovery Procedures**
   - Database restore from backup
   - Application deployment
   - Configuration restoration
   - Health checks and validation

## Troubleshooting

### Common Issues

1. **Database Connection Issues**
   ```bash
   # Check database connectivity
   kubectl exec -it bazarutod-pod -- nc -zv postgres-service 5432
   ```

2. **Redis Connection Issues**
   ```bash
   # Check Redis connectivity
   kubectl exec -it bazarutod-pod -- redis-cli -h redis-service ping
   ```

3. **Memory Issues**
   ```bash
   # Check memory usage
   kubectl top pods -n bazaruto
   ```

### Health Checks

```bash
# Application health
curl http://localhost:8080/healthz

# Database health
curl http://localhost:8080/healthz/db

# Redis health
curl http://localhost:8080/healthz/redis
```

### Logs

```bash
# Application logs
kubectl logs -f deployment/bazarutod -n bazaruto

# Job worker logs
kubectl logs -f deployment/bazarutod-workers -n bazaruto
```