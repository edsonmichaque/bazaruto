# Bazaruto Insurance Platform

A production-grade Go backend service for an insurance marketplace platform, built with modern architecture patterns and inspired by Rails/Laravel conventions.

## 🚀 Features

### Core Platform
- **Layered Architecture**: Clean separation with Store, Services, Handlers, and Router layers
- **RESTful API**: Complete CRUD operations for all insurance entities
- **Authentication & Authorization**: JWT-based auth with RBAC and policy-based authorization
- **Event-Driven Architecture**: Comprehensive event bus system with domain events
- **Job Processing**: Background job system with multiple queue backends
- **Webhook System**: Stripe-inspired persistent retry mechanism for external integrations

### Insurance Domain
- **User Management**: Customer, agent, and admin roles
- **Product Catalog**: Insurance products with dynamic pricing
- **Quote System**: Real-time premium calculations
- **Policy Management**: Policy lifecycle management
- **Claims Processing**: Fraud detection and payout settlement
- **Payment Processing**: Secure payment handling with multiple providers

### Technical Features
- **Database**: PostgreSQL with GORM ORM and UUID primary keys
- **Caching**: Redis integration for sessions and rate limiting
- **Observability**: Structured logging (Zap), Prometheus metrics, OpenTelemetry tracing
- **Rate Limiting**: Token bucket algorithm with Redis and in-memory backends
- **CLI Tools**: Comprehensive command-line interface for management
- **GitHub-style Pagination**: Modern API pagination patterns

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Layer    │    │  Business Logic │    │   Data Layer    │
│                 │    │                 │    │                 │
│  ┌───────────┐  │    │  ┌───────────┐  │    │  ┌───────────┐  │
│  │  Router   │  │───▶│  │ Services  │  │───▶│  │   Store   │  │
│  └───────────┘  │    │  └───────────┘  │    │  └───────────┘  │
│  ┌───────────┐  │    │  ┌───────────┐  │    │  ┌───────────┐  │
│  │ Handlers  │  │    │  │Event Bus  │  │    │  │   GORM    │  │
│  └───────────┘  │    │  └───────────┘  │    │  └───────────┘  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Middleware    │    │   Job System    │    │   PostgreSQL    │
│                 │    │                 │    │                 │
│ • Auth/Authorize│    │ • Email Jobs    │    │ • UUID Primary  │
│ • Rate Limiting │    │ • PDF Generation│    │ • JSONB Fields  │
│ • Logging       │    │ • Webhook Jobs  │    │ • Auto Migrate  │
│ • Metrics       │    │ • Payment Jobs  │    │ • Transactions  │
│ • Tracing       │    │ • Fraud Detection│   │ • Full Text     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📁 Project Structure

```
bazaruto/
├── cmd/
│   └── bazarutod/           # Main application entry point
├── internal/
│   ├── authentication/      # JWT, OIDC, MFA authentication
│   ├── authorization/       # RBAC and policy-based authorization
│   ├── commands/            # CLI commands (serve, worker, queues, jobs)
│   ├── config/              # Business rules configuration (file-based, no DB dependency)
│   ├── events/              # Event bus and domain events
│   ├── eventadapters/       # Event storage adapters
│   ├── eventhandlers/       # Event handlers and webhook system
│   ├── handlers/            # HTTP request handlers
│   ├── job/                 # Job system core (dispatcher, worker, registry)
│   ├── jobadapters/         # Job queue adapters (memory, Redis, database)
│   ├── jobs/                # Job implementations (email, payment, PDF, etc.)
│   ├── logger/              # Zap-based structured logging
│   ├── metrics/             # Prometheus metrics
│   ├── middleware/          # HTTP middleware (auth, rate limiting, etc.)
│   ├── models/              # Domain models and database entities
│   ├── policies/            # Authorization policies
│   ├── router/              # Chi router configuration
│   ├── services/            # Business logic services
│   ├── store/               # Data access layer (GORM repositories)
│   └── tracing/             # OpenTelemetry tracing
├── deploy/                  # Docker Compose and Kubernetes manifests
├── docs/                    # Documentation
├── test/                    # Integration and E2E tests
└── build/                   # CI/CD workflows and Dockerfile
```

## 🚀 Quick Start

### Prerequisites
- Go 1.22+
- PostgreSQL 14+
- Redis 6+
- Docker & Docker Compose (optional)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/edsonmichaque/bazaruto.git
   cd bazaruto
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up the database**
   ```bash
   # Start PostgreSQL and Redis
   docker-compose up -d postgres redis
   
   # Run migrations
   make migrate
   ```

4. **Configure the application**
   ```bash
   cp config.yaml.example config.yaml
   # Edit config.yaml with your settings
   ```

5. **Start the server**
   ```bash
   make run
   # or
   go run cmd/bazarutod/main.go serve
   ```

### Docker Deployment

```bash
# Build and run with Docker Compose
docker-compose up -d

# View logs
docker-compose logs -f bazarutod
```

## 🛠️ Development

### Available Commands

```bash
# Start the API server
make run
go run cmd/bazarutod/main.go serve

# Start job workers
go run cmd/bazarutod/main.go worker

# Run database migrations
make migrate
go run cmd/bazarutod/main.go migrate

# Queue management
go run cmd/bazarutod/main.go queues list
go run cmd/bazarutod/main.go queues pause <queue_name>
go run cmd/bazarutod/main.go queues monitor

# Job management
go run cmd/bazarutod/main.go jobs list
go run cmd/bazarutod/main.go jobs retry --id <job_id>
go run cmd/bazarutod/main.go jobs stats

# Run tests
make test
go test ./...

# Build binary
make build
go build -o bin/bazarutod cmd/bazarutod/main.go
```

### Configuration

The application uses a dual configuration system:

1. **Application Configuration**: Viper-based configuration for infrastructure settings
2. **Business Rules Configuration**: File-based configuration for business logic (independent of database)

```yaml
server:
  addr: ":8080"
  read_timeout: 30s
  write_timeout: 30s

db:
  host: localhost
  port: 5432
  name: bazaruto
  user: postgres
  password: password
  ssl_mode: disable

redis:
  address: localhost:6379
  password: ""
  db: 0

jobs:
  adapter: memory  # memory, redis, database
  queues: ["mailers", "payments", "processing", "notifications"]
  concurrency: 10
  max_retries: 5

log_level: info
log_format: json
metrics_enabled: true
tracing:
  enabled: true
  service_name: bazaruto
  endpoint: http://localhost:14268/api/traces
```

## 📊 API Endpoints

### Authentication
- `POST /auth/login` - User login
- `POST /auth/register` - User registration
- `POST /auth/refresh` - Refresh JWT token
- `POST /auth/logout` - User logout

### Users
- `GET /users` - List users (paginated)
- `GET /users/{id}` - Get user details
- `POST /users` - Create user
- `PUT /users/{id}` - Update user
- `DELETE /users/{id}` - Delete user

### Products
- `GET /products` - List insurance products
- `GET /products/{id}` - Get product details
- `POST /products` - Create product (admin only)
- `PUT /products/{id}` - Update product (admin only)

### Quotes
- `GET /quotes` - List quotes
- `GET /quotes/{id}` - Get quote details
- `POST /quotes` - Create new quote
- `POST /quotes/{id}/calculate` - Calculate premium

### Policies
- `GET /policies` - List policies
- `GET /policies/{id}` - Get policy details
- `POST /policies` - Create policy
- `PUT /policies/{id}` - Update policy

### Claims
- `GET /claims` - List claims
- `GET /claims/{id}` - Get claim details
- `POST /claims` - Submit claim
- `PUT /claims/{id}` - Update claim

### Payments
- `GET /payments` - List payments
- `GET /payments/{id}` - Get payment details
- `POST /payments` - Process payment
- `POST /payments/{id}/refund` - Refund payment

### Webhooks
- `GET /webhooks/configs` - List webhook configurations
- `POST /webhooks/configs` - Create webhook configuration
- `GET /webhooks/deliveries` - List webhook deliveries
- `GET /webhooks/deliveries/{id}` - Get delivery details

### Health & Monitoring
- `GET /healthz` - Health check
- `GET /metrics` - Prometheus metrics

## 🔧 Job System

The platform includes a comprehensive job processing system:

### Job Types
- **Email Jobs**: Welcome emails, password resets, notifications
- **Payment Jobs**: Payment processing, refunds, payouts
- **PDF Jobs**: Quote and policy document generation
- **Notification Jobs**: Push notifications, SMS
- **Webhook Jobs**: External system integrations with retry logic
- **Processing Jobs**: Premium calculations, fraud detection

### Queue Management
```bash
# List all queues
bazarutod queues list

# Pause a queue
bazarutod queues pause mailers

# Resume a queue
bazarutod queues resume mailers

# Monitor queue health
bazarutod queues monitor --interval 30s --duration 5m
```

### Job Management
```bash
# List jobs with filters
bazarutod jobs list --queue mailers --status failed --limit 50

# Retry failed jobs
bazarutod jobs retry --id <job_id>
bazarutod jobs retry --all --queue mailers

# Cancel jobs
bazarutod jobs cancel --id <job_id>
bazarutod jobs cancel --all --queue processing

# View job statistics
bazarutod jobs stats --queue payments --duration 24h

# Clean up old jobs
bazarutod jobs cleanup --older-than 30d --dry-run
```

## 🔐 Security

- **JWT Authentication**: Secure token-based authentication
- **RBAC Authorization**: Role-based access control (Admin, Agent, Customer)
- **Policy-based Authorization**: Resource-specific authorization policies
- **Rate Limiting**: Configurable rate limiting with Redis and in-memory backends
- **Input Validation**: Comprehensive request validation
- **SQL Injection Protection**: GORM ORM with parameterized queries
- **CORS Configuration**: Configurable cross-origin resource sharing

## 📈 Monitoring & Observability

### Logging
- **Structured Logging**: Zap-based JSON logging
- **Log Levels**: Debug, Info, Warn, Error, Fatal, Panic
- **Request Tracing**: Full request/response logging with correlation IDs

### Metrics
- **Prometheus Integration**: Custom business metrics
- **HTTP Metrics**: Request duration, status codes, throughput
- **Job Metrics**: Job processing times, success/failure rates
- **Database Metrics**: Connection pool, query performance

### Tracing
- **OpenTelemetry**: Distributed tracing support
- **Jaeger Integration**: Trace visualization and analysis
- **Custom Spans**: Business logic tracing

## 🧪 Testing

```bash
# Run all tests
make test

# Run specific test suites
go test ./internal/services/...
go test ./internal/handlers/...

# Run integration tests
go test ./test/integration/...

# Run E2E tests
go test ./test/e2e/...

# Generate test coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🚀 Deployment

### Docker
```bash
# Build Docker image
docker build -t bazaruto:latest .

# Run with Docker Compose
docker-compose up -d
```

### Kubernetes
```bash
# Apply Kubernetes manifests
kubectl apply -f deploy/k8s/

# Check deployment status
kubectl get pods -l app=bazaruto
```

### Production Considerations
- Use external PostgreSQL and Redis instances
- Configure proper logging levels
- Set up monitoring and alerting
- Use secrets management for sensitive data
- Configure proper resource limits
- Set up backup and disaster recovery

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: Check the `/docs` directory for detailed documentation
- **Issues**: Report bugs and request features via GitHub Issues
- **Discussions**: Join community discussions in GitHub Discussions

## 🏆 Acknowledgments

- Inspired by Rails/Laravel conventions and patterns
- Built with modern Go best practices
- Webhook retry mechanism inspired by Stripe's implementation
- Architecture patterns from Domain-Driven Design (DDD)

---

**Bazaruto Insurance Platform** - Building the future of insurance technology 🚀