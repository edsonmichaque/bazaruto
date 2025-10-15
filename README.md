# Bazaruto Insurance Platform

A production-grade Go backend service for an insurance marketplace platform, built with modern architecture patterns and inspired by Rails/Laravel conventions.

## 🚀 Features

- **Layered Architecture**: Clean separation with Store, Services, Handlers, and Router layers
- **RESTful API**: Complete CRUD operations for all insurance entities
- **Authentication & Authorization**: JWT-based auth with RBAC and policy-based authorization
- **Event-Driven Architecture**: Comprehensive event bus system with domain events
- **Job Processing**: Background job system with multiple queue backends
- **Webhook System**: Stripe-inspired persistent retry mechanism for external integrations
- **Insurance Domain**: User management, product catalog, quotes, policies, claims, and payments
- **Observability**: Structured logging, Prometheus metrics, OpenTelemetry tracing

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
├── cmd/bazarutod/           # Main application entry point
├── internal/
│   ├── authentication/      # JWT, OIDC, MFA authentication
│   ├── authorization/       # RBAC and policy-based authorization
│   ├── commands/            # CLI commands (serve, worker, queues, jobs)
│   ├── config/              # Business rules configuration (file-based, no DB dependency)
│   ├── events/              # Event bus and domain events
│   ├── handlers/            # HTTP request handlers
│   ├── jobs/                # Job implementations (email, payment, PDF, etc.)
│   ├── middleware/          # HTTP middleware (auth, rate limiting, etc.)
│   ├── models/              # Domain models and database entities
│   ├── services/            # Business logic services
│   └── store/               # Data access layer (GORM repositories)
├── pkg/
│   ├── event/               # General-purpose event system
│   └── job/                 # General-purpose job system
├── deploy/                  # Docker Compose and Kubernetes manifests
├── docs/                    # Documentation
└── test/                    # Integration and E2E tests
```

## 🚀 Quick Start

### Prerequisites
- Go 1.22+
- PostgreSQL 14+
- Redis 6+

### Installation

1. **Clone and setup**
   ```bash
   git clone https://github.com/edsonmichaque/bazaruto-insurance.git
   cd bazaruto-insurance
   go mod download
   ```

2. **Start dependencies**
   ```bash
   docker-compose up -d postgres redis
   ```

3. **Configure and run**
   ```bash
   cp config.yaml.example config.yaml
   # Edit config.yaml with your settings
   go run cmd/bazarutod/main.go serve
   ```

### Docker Deployment

```bash
docker-compose up -d
```

## 🛠️ Development

```bash
# Start the API server
go run cmd/bazarutod/main.go serve

# Start job workers
go run cmd/bazarutod/main.go worker

# Run database migrations
go run cmd/bazarutod/main.go migrate

# Queue management
go run cmd/bazarutod/main.go queues list
go run cmd/bazarutod/main.go jobs list

# Run tests
go test ./...

# Build binary
go build -o bin/bazarutod cmd/bazarutod/main.go
```

## 📊 API Endpoints

### Core Resources
- **Authentication**: `/auth/login`, `/auth/register`, `/auth/refresh`
- **Users**: `/users` (CRUD operations)
- **Products**: `/products` (Insurance products)
- **Quotes**: `/quotes` (Premium calculations)
- **Policies**: `/policies` (Policy management)
- **Claims**: `/claims` (Claims processing)
- **Payments**: `/payments` (Payment processing)
- **Webhooks**: `/webhooks/configs`, `/webhooks/deliveries`

### Monitoring
- **Health**: `/healthz`
- **Metrics**: `/metrics`

## 🔧 Configuration

The application uses a dual configuration system:

1. **Application Configuration**: Viper-based configuration for infrastructure settings
2. **Business Rules Configuration**: File-based configuration for business logic (independent of database)

See [docs/configuration.md](docs/configuration.md) for detailed configuration options.

## 🔐 Security

- **JWT Authentication**: Secure token-based authentication
- **RBAC Authorization**: Role-based access control (Admin, Agent, Customer)
- **Rate Limiting**: Configurable rate limiting with Redis and in-memory backends
- **Input Validation**: Comprehensive request validation
- **SQL Injection Protection**: GORM ORM with parameterized queries

## 📈 Monitoring & Observability

- **Structured Logging**: Zap-based JSON logging
- **Prometheus Metrics**: Custom business metrics and HTTP metrics
- **OpenTelemetry Tracing**: Distributed tracing support
- **Health Checks**: Comprehensive health monitoring

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run specific test suites
go test ./internal/services/...
go test ./internal/handlers/...

# Generate test coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🚀 Deployment

### Docker
```bash
docker build -t bazaruto:latest .
docker-compose up -d
```

### Kubernetes
```bash
kubectl apply -f deploy/kubernetes/
```

## 📚 Documentation

- [Architecture](docs/architecture.md) - System architecture and design patterns
- [Configuration](docs/configuration.md) - Configuration options and business rules
- [Deployment](docs/deployment.md) - Deployment instructions
- [Contributing](docs/contributing.md) - Contribution guidelines
- [Operations Guide](docs/ops-guide.md) - Production operations

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Bazaruto Insurance Platform** - Building the future of insurance technology 🚀