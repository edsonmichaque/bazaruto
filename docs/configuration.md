# Configuration Guide

This document describes all configuration options available in the Bazaruto Insurance Platform.

## Configuration Files

The application uses two main configuration systems:

### 1. Application Configuration
Uses Viper for configuration management and supports multiple formats:
- YAML (recommended)
- JSON
- TOML
- Environment variables

Default configuration file: `config.yaml`

### 2. Business Rules Configuration
Uses a dedicated file-based configuration system for business rules:
- JSON format (recommended)
- File-based storage with in-memory caching
- Environment-specific configuration files
- Automatic fallback to sensible defaults

Default business rules file: `config/business_rules.json`

## Configuration Structure

```yaml
# Server Configuration
server:
  addr: ":8080"                    # Server address and port
  read_timeout: 30s                # HTTP read timeout
  write_timeout: 30s               # HTTP write timeout
  idle_timeout: 120s               # HTTP idle timeout
  max_header_bytes: 1048576        # Maximum header size (1MB)

# Database Configuration
db:
  host: localhost                  # Database host
  port: 5432                       # Database port
  name: bazaruto                   # Database name
  user: postgres                   # Database user
  password: password               # Database password
  ssl_mode: disable                # SSL mode (disable, require, verify-ca, verify-full)
  max_connections: 25              # Maximum database connections
  min_connections: 5               # Minimum database connections
  connect_timeout: 30s             # Connection timeout
  acquire_timeout: 30s             # Connection acquire timeout
  max_lifetime: 1h                 # Connection max lifetime
  idle_timeout: 30m                # Connection idle timeout

# Redis Configuration
redis:
  address: localhost:6379          # Redis server address
  password: ""                     # Redis password
  db: 0                            # Redis database number
  max_connections: 10              # Maximum Redis connections
  min_connections: 1               # Minimum Redis connections
  connect_timeout: 5s              # Connection timeout
  acquire_timeout: 5s              # Connection acquire timeout
  max_lifetime: 1h                 # Connection max lifetime
  idle_timeout: 5m                 # Connection idle timeout

# Rate Limiting Configuration
rate:
  enabled: true                    # Enable rate limiting
  provider: memory                 # Rate limiter provider (memory, redis)
  requests_per_minute: 60          # Requests per minute limit
  burst: 10                        # Burst capacity
  cleanup_interval: 1m             # Cleanup interval for memory provider

# Job System Configuration
jobs:
  adapter: memory                  # Job queue adapter (memory, redis, database)
  queues:                          # Available queues
    - mailers                      # Email processing queue
    - payments                     # Payment processing queue
    - processing                   # General processing queue
    - notifications                # Notification queue
    - heavy                        # Heavy processing queue
  concurrency: 10                  # Worker concurrency
  poll_interval: 1s                # Poll interval for jobs
  max_retries: 5                   # Maximum retry attempts
  timeout: 30m                     # Job timeout
  redis:                           # Redis-specific job config
    address: localhost:6379
    password: ""
    db: 1
  database:                        # Database-specific job config
    table_name: jobs
    max_connections: 5

# Authentication Configuration
auth:
  jwt:
    secret: "your-secret-key"      # JWT signing secret
    expires_in: 15m                # Access token expiration
    refresh_expires_in: 7d         # Refresh token expiration
    issuer: "bazaruto"             # JWT issuer
    audience: "bazaruto-api"       # JWT audience
  password:
    min_length: 8                  # Minimum password length
    require_uppercase: true        # Require uppercase letters
    require_lowercase: true        # Require lowercase letters
    require_numbers: true          # Require numbers
    require_symbols: false         # Require special symbols
  mfa:
    enabled: true                  # Enable MFA
    issuer: "Bazaruto Insurance"   # TOTP issuer name
    window: 1                      # TOTP time window

# Email Configuration
email:
  provider: smtp                   # Email provider (smtp, sendgrid, ses)
  smtp:
    host: localhost                # SMTP host
    port: 587                      # SMTP port
    username: ""                   # SMTP username
    password: ""                   # SMTP password
    from: "noreply@bazaruto.com"   # Default from address
    tls: true                      # Use TLS
  sendgrid:
    api_key: ""                    # SendGrid API key
    from: "noreply@bazaruto.com"   # Default from address
  ses:
    region: us-east-1              # AWS region
    access_key: ""                 # AWS access key
    secret_key: ""                 # AWS secret key
    from: "noreply@bazaruto.com"   # Default from address

# Payment Configuration
payment:
  provider: stripe                 # Payment provider (stripe, paypal)
  stripe:
    public_key: ""                 # Stripe public key
    secret_key: ""                 # Stripe secret key
    webhook_secret: ""             # Stripe webhook secret
  paypal:
    client_id: ""                  # PayPal client ID
    client_secret: ""              # PayPal client secret
    sandbox: true                  # Use PayPal sandbox

# Webhook Configuration
webhook:
  timeout: 30s                     # Webhook request timeout
  max_retries: 10                  # Maximum retry attempts
  retry_backoff: 2s                # Base retry backoff
  signature_header: "X-Signature"  # Signature header name
  secret: ""                       # Webhook signing secret

# Observability Configuration
log_level: info                    # Log level (debug, info, warn, error, fatal, panic)
log_format: json                   # Log format (json, text)
metrics_enabled: true              # Enable Prometheus metrics
metrics_path: /metrics             # Metrics endpoint path
tracing:
  enabled: true                    # Enable distributed tracing
  service_name: bazaruto           # Service name for tracing
  endpoint: http://localhost:14268/api/traces  # Jaeger endpoint
  sample_rate: 0.1                 # Sampling rate (0.0 to 1.0)

# CORS Configuration
cors:
  enabled: true                    # Enable CORS
  allowed_origins:                 # Allowed origins
    - "http://localhost:3000"
    - "https://bazaruto.com"
  allowed_methods:                 # Allowed HTTP methods
    - GET
    - POST
    - PUT
    - DELETE
    - OPTIONS
  allowed_headers:                 # Allowed headers
    - Content-Type
    - Authorization
    - X-Requested-With
  allow_credentials: true          # Allow credentials
  max_age: 86400                   # Max age for preflight requests

# Security Configuration
security:
  bcrypt_cost: 12                  # Bcrypt hashing cost
  session_timeout: 24h             # Session timeout
  max_login_attempts: 5            # Maximum login attempts
  lockout_duration: 15m            # Account lockout duration
  password_reset_timeout: 1h       # Password reset token timeout
  email_verification_timeout: 24h  # Email verification timeout
```

## Environment Variables

All configuration options can be overridden using environment variables. Environment variables use the `BAZARUTO_` prefix and nested keys are separated by underscores.

### Examples

```bash
# Server configuration
export BAZARUTO_SERVER_ADDR=":8080"
export BAZARUTO_SERVER_READ_TIMEOUT="30s"

# Database configuration
export BAZARUTO_DB_HOST="localhost"
export BAZARUTO_DB_PORT="5432"
export BAZARUTO_DB_NAME="bazaruto"
export BAZARUTO_DB_USER="postgres"
export BAZARUTO_DB_PASSWORD="password"

# Redis configuration
export BAZARUTO_REDIS_ADDRESS="localhost:6379"
export BAZARUTO_REDIS_PASSWORD=""
export BAZARUTO_REDIS_DB="0"

# Authentication
export BAZARUTO_AUTH_JWT_SECRET="your-secret-key"
export BAZARUTO_AUTH_JWT_EXPIRES_IN="15m"

# Logging
export BAZARUTO_LOG_LEVEL="info"
export BAZARUTO_LOG_FORMAT="json"

# Metrics and Tracing
export BAZARUTO_METRICS_ENABLED="true"
export BAZARUTO_TRACING_ENABLED="true"
export BAZARUTO_TRACING_SERVICE_NAME="bazaruto"
```

## Configuration Validation

The application validates configuration on startup and will fail with descriptive error messages if invalid values are provided.

### Validation Rules

1. **Server Configuration**
   - `addr` must be a valid address format
   - Timeouts must be positive durations
   - `max_header_bytes` must be positive

2. **Database Configuration**
   - `host` must not be empty
   - `port` must be between 1 and 65535
   - `name` must not be empty
   - `user` must not be empty
   - `password` must not be empty
   - `ssl_mode` must be one of: disable, require, verify-ca, verify-full
   - Connection pool settings must be positive

3. **Redis Configuration**
   - `address` must be a valid address format
   - `db` must be between 0 and 15
   - Connection pool settings must be positive

4. **Authentication Configuration**
   - `jwt.secret` must not be empty
   - Expiration times must be positive durations
   - Password requirements must be valid

5. **Email Configuration**
   - Provider must be one of: smtp, sendgrid, ses
   - SMTP settings must be valid when using SMTP provider
   - API keys must not be empty when using cloud providers

6. **Payment Configuration**
   - Provider must be one of: stripe, paypal
   - API keys must not be empty when using payment providers

## Configuration Loading Order

Configuration is loaded in the following order (later values override earlier ones):

1. Default values (hardcoded in application)
2. Configuration file (`config.yaml`)
3. Environment variables
4. Command-line flags

## Production Configuration

### Security Considerations

1. **Secrets Management**
   - Use environment variables for sensitive data
   - Consider using a secrets management service (AWS Secrets Manager, HashiCorp Vault)
   - Never commit secrets to version control

2. **Database Security**
   - Use SSL connections in production
   - Implement connection pooling
   - Use read replicas for read-heavy workloads

3. **Authentication Security**
   - Use strong JWT secrets (32+ characters)
   - Implement proper password policies
   - Enable MFA for admin users

4. **Network Security**
   - Configure proper CORS settings
   - Use HTTPS in production
   - Implement rate limiting

### Performance Tuning

1. **Database Configuration**
   ```yaml
   db:
     max_connections: 50
     min_connections: 10
     max_lifetime: 1h
     idle_timeout: 30m
   ```

2. **Redis Configuration**
   ```yaml
   redis:
     max_connections: 20
     min_connections: 5
     max_lifetime: 1h
     idle_timeout: 10m
   ```

3. **Job System Configuration**
   ```yaml
   jobs:
     concurrency: 20
     poll_interval: 500ms
     timeout: 10m
   ```

### Monitoring Configuration

1. **Logging**
   ```yaml
   log_level: info
   log_format: json
   ```

2. **Metrics**
   ```yaml
   metrics_enabled: true
   metrics_path: /metrics
   ```

3. **Tracing**
   ```yaml
   tracing:
     enabled: true
     service_name: bazaruto-prod
     endpoint: http://jaeger:14268/api/traces
     sample_rate: 0.01
   ```

## Configuration Examples

### Development Configuration

```yaml
server:
  addr: ":8080"

db:
  host: localhost
  port: 5432
  name: bazaruto_dev
  user: postgres
  password: password
  ssl_mode: disable

redis:
  address: localhost:6379
  password: ""
  db: 0

log_level: debug
log_format: text
metrics_enabled: false
tracing:
  enabled: false
```

### Production Configuration

```yaml
server:
  addr: ":8080"
  read_timeout: 30s
  write_timeout: 30s

db:
  host: db.internal
  port: 5432
  name: bazaruto_prod
  user: bazaruto_user
  password: ${DB_PASSWORD}
  ssl_mode: require
  max_connections: 50
  min_connections: 10

redis:
  address: redis.internal:6379
  password: ${REDIS_PASSWORD}
  db: 0
  max_connections: 20

auth:
  jwt:
    secret: ${JWT_SECRET}
    expires_in: 15m
    refresh_expires_in: 7d

log_level: info
log_format: json
metrics_enabled: true
tracing:
  enabled: true
  service_name: bazaruto-prod
  endpoint: http://jaeger:14268/api/traces
  sample_rate: 0.01
```

### Docker Configuration

```yaml
server:
  addr: ":8080"

db:
  host: postgres
  port: 5432
  name: bazaruto
  user: postgres
  password: password
  ssl_mode: disable

redis:
  address: redis:6379
  password: ""
  db: 0

log_level: info
log_format: json
```

## Business Rules Configuration

The business rules configuration system provides dynamic, file-based configuration for all business logic components. This system is completely independent of the database and provides fast, in-memory access to configuration data.

### Configuration File Structure

The business rules configuration is stored in JSON format with the following structure:

```json
{
  "fraud_detection": {
    "enabled": true,
    "version": "1.0",
    "risk_thresholds": {
      "low": 30.0,
      "medium": 60.0,
      "high": 80.0,
      "critical": 90.0
    },
    "factor_weights": {
      "claim_frequency": 0.3,
      "claim_amount": 0.25,
      "geographic_risk": 0.2,
      "payment_history": 0.15,
      "policy_duration": 0.1
    },
    "timing_rules": {
      "new_account_threshold": "4320h",
      "policy_start_threshold": "168h",
      "reporting_delay_threshold": "720h",
      "weekend_multiplier": 1.2,
      "business_hours_multiplier": 0.9
    },
    "amount_rules": {
      "high_value_threshold": 10000.0,
      "very_high_value_threshold": 100000.0,
      "round_number_penalty": 15.0,
      "coverage_ratio_threshold": 0.95
    },
    "document_rules": {
      "min_document_count": 2,
      "min_file_size": 1024,
      "max_file_size": 10485760,
      "required_document_types": ["id", "proof_of_address", "financial_statement"]
    },
    "geographic_rules": {
      "high_risk_countries": ["AF", "IR", "KP", "SY"],
      "high_risk_regions": ["middle_east", "africa"],
      "country_risk_multiplier": {
        "AF": 2.0,
        "IR": 2.5,
        "KP": 3.0,
        "SY": 2.0
      }
    },
    "behavioral_rules": {
      "min_description_length": 50,
      "max_description_length": 1000,
      "description_length_penalty": 10.0
    },
    "confidence_thresholds": {
      "low": 0.3,
      "medium": 0.5,
      "high": 0.7,
      "very_high": 0.9
    },
    "auto_review_thresholds": {
      "score_threshold": 70.0,
      "critical_factor_count": 1,
      "high_severity_count": 2
    }
  },
  "risk_assessment": {
    "enabled": true,
    "version": "1.0",
    "factor_weights": {
      "age": 0.2,
      "occupation": 0.25,
      "health_status": 0.3,
      "lifestyle": 0.15,
      "family_history": 0.1
    }
  },
  "pricing": {
    "enabled": true,
    "version": "1.0",
    "base_rates": {
      "life_insurance": 50.0,
      "health_insurance": 200.0,
      "auto_insurance": 100.0,
      "home_insurance": 150.0,
      "travel_insurance": 25.0,
      "disability_insurance": 75.0
    }
  },
  "underwriting": {
    "enabled": true,
    "version": "1.0"
  },
  "commission": {
    "enabled": true,
    "version": "1.0"
  },
  "compliance": {
    "enabled": true,
    "version": "1.0"
  },
  "policy_lifecycle": {
    "enabled": true,
    "version": "1.0"
  },
  "claim_processing": {
    "enabled": true,
    "version": "1.0"
  }
}
```

### Configuration Management

The business rules configuration is managed through the `ConfigManager` service:

```go
// Create config manager
configManager := config.NewConfigManager(logger, "config/production.json")

// Load configuration
if err := configManager.LoadConfig(ctx); err != nil {
    log.Fatal("Failed to load config:", err)
}

// Get current configuration
businessRules := configManager.GetConfig()

// Update configuration
businessRules.FraudDetection.Enabled = false
if err := configManager.UpdateConfig(ctx, businessRules); err != nil {
    log.Error("Failed to update config:", err)
}
```

### Configuration Features

- **File-based Storage**: Configuration is stored in JSON files for easy editing and version control
- **In-memory Caching**: Fast access to configuration data without file I/O
- **Automatic Fallback**: Uses sensible defaults if configuration file is missing
- **Thread-safe**: Safe for concurrent access across multiple goroutines
- **Environment-specific**: Different configuration files for different environments
- **Hot Reloading**: Configuration can be updated without restarting the application

### Environment-specific Configuration

You can use different configuration files for different environments:

```bash
# Development
configManager := config.NewConfigManager(logger, "config/development.json")

# Staging
configManager := config.NewConfigManager(logger, "config/staging.json")

# Production
configManager := config.NewConfigManager(logger, "config/production.json")
```

### Configuration Validation

The configuration system includes built-in validation:

- Risk thresholds must be in ascending order
- Factor weights must sum to 1.0
- Time durations must be valid
- Required fields cannot be empty

## Troubleshooting

### Common Configuration Issues

1. **Database Connection Issues**
   - Check host, port, and credentials
   - Verify SSL mode settings
   - Ensure database exists

2. **Redis Connection Issues**
   - Check address and port
   - Verify password if required
   - Ensure Redis is running

3. **Authentication Issues**
   - Verify JWT secret is set
   - Check token expiration settings
   - Ensure proper password policies

4. **Email Configuration Issues**
   - Verify SMTP settings
   - Check API keys for cloud providers
   - Test email delivery

5. **Payment Configuration Issues**
   - Verify API keys
   - Check webhook secrets
   - Ensure proper environment settings

### Configuration Validation Errors

The application provides detailed error messages for configuration validation failures. Common error messages include:

- `invalid server address format`
- `database host cannot be empty`
- `invalid SSL mode`
- `JWT secret cannot be empty`
- `invalid log level`
- `invalid email provider`

### Debugging Configuration

To debug configuration loading:

1. Enable debug logging: `log_level: debug`
2. Check configuration loading logs
3. Use configuration validation endpoints
4. Test individual components

For more detailed troubleshooting, refer to the [Operations Guide](ops-guide.md).