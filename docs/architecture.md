# Architecture Overview

## System Architecture

The Bazaruto Insurance Platform follows a layered architecture pattern with clear separation of concerns, inspired by Domain-Driven Design (DDD) principles and Rails/Laravel conventions.

## High-Level Architecture

```mermaid
graph TB
    subgraph "Client Layer"
        WEB[Web Application]
        MOBILE[Mobile App]
        API_CLIENT[API Clients]
    end
    
    subgraph "API Gateway Layer"
        LB[Load Balancer]
        RATE_LIMIT[Rate Limiter]
        AUTH[Authentication]
    end
    
    subgraph "Application Layer"
        ROUTER[Chi Router]
        MIDDLEWARE[Middleware Stack]
        HANDLERS[HTTP Handlers]
    end
    
    subgraph "Business Logic Layer"
        SERVICES[Domain Services]
        EVENTS[Event Bus]
        JOBS[Job System]
    end
    
    subgraph "Data Layer"
        STORE[Repository Layer]
        GORM[GORM ORM]
        DB[(PostgreSQL)]
    end
    
    subgraph "External Services"
        REDIS[(Redis)]
        SMTP[Email Service]
        PAYMENT[Payment Gateway]
        WEBHOOK[Webhook Endpoints]
    end
    
    WEB --> LB
    MOBILE --> LB
    API_CLIENT --> LB
    
    LB --> RATE_LIMIT
    RATE_LIMIT --> AUTH
    AUTH --> ROUTER
    
    ROUTER --> MIDDLEWARE
    MIDDLEWARE --> HANDLERS
    
    HANDLERS --> SERVICES
    SERVICES --> STORE
    SERVICES --> EVENTS
    SERVICES --> JOBS
    
    STORE --> GORM
    GORM --> DB
    
    EVENTS --> REDIS
    JOBS --> REDIS
    SERVICES --> SMTP
    SERVICES --> PAYMENT
    JOBS --> WEBHOOK
```

## Layer Details

### 1. Client Layer
- **Web Application**: React/Vue.js frontend
- **Mobile App**: iOS/Android applications
- **API Clients**: Third-party integrations

### 2. API Gateway Layer
- **Load Balancer**: Distributes traffic across multiple instances
- **Rate Limiter**: Token bucket algorithm with Redis backend
- **Authentication**: JWT token validation and refresh

### 3. Application Layer
- **Chi Router**: High-performance HTTP router
- **Middleware Stack**: 
  - Recovery middleware
  - Logging middleware
  - Authentication middleware
  - Authorization middleware
  - Rate limiting middleware
  - Metrics middleware
  - Tracing middleware
- **HTTP Handlers**: Request/response handling with JSON helpers

### 4. Business Logic Layer
- **Domain Services**: Core business logic
  - UserService
  - ProductService
  - QuoteService
  - PolicyService
  - ClaimService
  - PaymentService
  - WebhookService
- **Event Bus**: Domain event publishing and subscription
- **Job System**: Background job processing with multiple queues

### 5. Data Layer
- **Repository Layer**: Data access abstraction
- **GORM ORM**: Object-relational mapping
- **PostgreSQL**: Primary database with UUID primary keys

### 6. External Services
- **Redis**: Caching, sessions, and job queues
- **Email Service**: SMTP/SendGrid integration
- **Payment Gateway**: Stripe/PayPal integration
- **Webhook Endpoints**: External system notifications

## Domain Model

```mermaid
erDiagram
    User ||--o{ Quote : creates
    User ||--o{ Policy : owns
    User ||--o{ Claim : submits
    User ||--o{ Payment : makes
    
    Product ||--o{ Quote : generates
    Product ||--o{ Policy : covers
    
    Quote ||--o| Policy : becomes
    Quote ||--o{ Payment : requires
    
    Policy ||--o{ Claim : covers
    Policy ||--o{ Subscription : has
    
    Claim ||--o{ Payment : receives
    
    User {
        uuid id PK
        string email
        string full_name
        string role
        timestamp created_at
        timestamp updated_at
    }
    
    Product {
        uuid id PK
        string name
        string category
        decimal base_price
        jsonb coverage_details
        timestamp created_at
        timestamp updated_at
    }
    
    Quote {
        uuid id PK
        uuid user_id FK
        uuid product_id FK
        decimal final_price
        string status
        timestamp created_at
        timestamp updated_at
    }
    
    Policy {
        uuid id PK
        uuid user_id FK
        uuid product_id FK
        uuid quote_id FK
        decimal premium
        string status
        timestamp start_date
        timestamp end_date
        timestamp created_at
        timestamp updated_at
    }
    
    Claim {
        uuid id PK
        uuid user_id FK
        uuid policy_id FK
        decimal amount
        string status
        string description
        timestamp created_at
        timestamp updated_at
    }
    
    Payment {
        uuid id PK
        uuid user_id FK
        uuid policy_id FK
        decimal amount
        string status
        string transaction_id
        timestamp created_at
        timestamp updated_at
    }
```

## Event-Driven Architecture

The platform uses an event-driven architecture for loose coupling and scalability:

```mermaid
sequenceDiagram
    participant U as User
    participant H as Handler
    participant S as Service
    participant E as Event Bus
    participant J as Job System
    participant W as Webhook
    
    U->>H: Create Quote
    H->>S: QuoteService.CreateQuote()
    S->>S: Save to Database
    S->>E: Publish QuoteCreatedEvent
    E->>J: Dispatch CalculatePremiumJob
    E->>J: Dispatch GenerateQuotePDFJob
    E->>W: Send Webhook Notification
    
    Note over J: Background Processing
    J->>S: Calculate Premium
    J->>S: Generate PDF
    J->>W: Deliver Webhook
```

## Job System Architecture

The job system supports multiple backends and provides reliable processing:

```mermaid
graph TB
    subgraph "Job Dispatcher"
        DISPATCHER[Job Dispatcher]
        REGISTRY[Job Registry]
    end
    
    subgraph "Queue Backends"
        MEMORY[Memory Queue]
        REDIS_Q[Redis Queue]
        DB_Q[Database Queue]
    end
    
    subgraph "Workers"
        WORKER1[Worker Pool 1]
        WORKER2[Worker Pool 2]
        WORKER3[Worker Pool 3]
    end
    
    subgraph "Job Types"
        EMAIL[Email Jobs]
        PAYMENT[Payment Jobs]
        PDF[PDF Jobs]
        WEBHOOK[Webhook Jobs]
    end
    
    DISPATCHER --> REGISTRY
    DISPATCHER --> MEMORY
    DISPATCHER --> REDIS_Q
    DISPATCHER --> DB_Q
    
    MEMORY --> WORKER1
    REDIS_Q --> WORKER2
    DB_Q --> WORKER3
    
    WORKER1 --> EMAIL
    WORKER2 --> PAYMENT
    WORKER3 --> PDF
    WORKER3 --> WEBHOOK
```

## Security Architecture

```mermaid
graph TB
    subgraph "Authentication"
        JWT[JWT Tokens]
        REFRESH[Refresh Tokens]
        MFA[MFA Support]
    end
    
    subgraph "Authorization"
        RBAC[Role-Based Access Control]
        POLICIES[Resource Policies]
        GATES[Custom Gates]
    end
    
    subgraph "Security Middleware"
        RATE_LIMIT[Rate Limiting]
        CORS[CORS Protection]
        VALIDATION[Input Validation]
        SANITIZATION[Data Sanitization]
    end
    
    JWT --> RBAC
    REFRESH --> JWT
    MFA --> JWT
    
    RBAC --> POLICIES
    POLICIES --> GATES
    
    RATE_LIMIT --> VALIDATION
    CORS --> SANITIZATION
```

## Deployment Architecture

```mermaid
graph TB
    subgraph "Load Balancer"
        LB[NGINX/HAProxy]
    end
    
    subgraph "Application Tier"
        APP1[App Instance 1]
        APP2[App Instance 2]
        APP3[App Instance 3]
    end
    
    subgraph "Data Tier"
        DB_MASTER[(PostgreSQL Master)]
        DB_REPLICA[(PostgreSQL Replica)]
        REDIS_CLUSTER[(Redis Cluster)]
    end
    
    subgraph "Monitoring"
        PROMETHEUS[Prometheus]
        GRAFANA[Grafana]
        JAEGER[Jaeger]
    end
    
    LB --> APP1
    LB --> APP2
    LB --> APP3
    
    APP1 --> DB_MASTER
    APP2 --> DB_MASTER
    APP3 --> DB_MASTER
    
    APP1 --> DB_REPLICA
    APP2 --> DB_REPLICA
    APP3 --> DB_REPLICA
    
    APP1 --> REDIS_CLUSTER
    APP2 --> REDIS_CLUSTER
    APP3 --> REDIS_CLUSTER
    
    APP1 --> PROMETHEUS
    APP2 --> PROMETHEUS
    APP3 --> PROMETHEUS
    
    PROMETHEUS --> GRAFANA
    APP1 --> JAEGER
    APP2 --> JAEGER
    APP3 --> JAEGER
```

## Key Design Principles

### 1. Separation of Concerns
- Each layer has a specific responsibility
- Clear boundaries between layers
- Dependency injection for testability

### 2. Domain-Driven Design
- Business logic encapsulated in services
- Rich domain models
- Event-driven communication

### 3. SOLID Principles
- Single Responsibility Principle
- Open/Closed Principle
- Liskov Substitution Principle
- Interface Segregation Principle
- Dependency Inversion Principle

### 4. Clean Architecture
- Independent of frameworks
- Testable business logic
- Independent of UI
- Independent of database
- Independent of external services

### 5. Event-Driven Architecture
- Loose coupling between components
- Scalable and maintainable
- Asynchronous processing
- Event sourcing capabilities

## Technology Stack

### Backend
- **Language**: Go 1.22+
- **Framework**: Chi v5 (HTTP router)
- **ORM**: GORM v2
- **Database**: PostgreSQL 14+
- **Cache**: Redis 6+
- **Authentication**: JWT with refresh tokens
- **Logging**: Zap (structured logging)
- **Metrics**: Prometheus
- **Tracing**: OpenTelemetry + Jaeger

### Infrastructure
- **Containerization**: Docker
- **Orchestration**: Kubernetes
- **CI/CD**: GitHub Actions
- **Monitoring**: Prometheus + Grafana
- **Log Aggregation**: ELK Stack (optional)

### External Services
- **Email**: SMTP/SendGrid
- **Payments**: Stripe/PayPal
- **File Storage**: AWS S3 (optional)
- **CDN**: CloudFlare (optional)

## Scalability Considerations

### Horizontal Scaling
- Stateless application instances
- Load balancer distribution
- Database read replicas
- Redis clustering

### Performance Optimization
- Connection pooling
- Query optimization
- Caching strategies
- Background job processing

### Monitoring and Observability
- Application metrics
- Business metrics
- Distributed tracing
- Health checks
- Alerting

This architecture provides a solid foundation for building a scalable, maintainable, and secure insurance platform.