# Configuration Guide

This document describes the configuration options for the Bazaruto Insurance Platform.

## Configuration Systems

The application uses two configuration systems:

1. **Application Configuration**: Viper-based configuration for infrastructure settings
2. **Business Rules Configuration**: File-based configuration for business logic (independent of database)

## Application Configuration

Default file: `config.yaml`

```yaml
# Server Configuration
server:
  addr: ":8080"
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s

# Database Configuration
db:
  host: localhost
  port: 5432
  name: bazaruto
  user: postgres
  password: password
  ssl_mode: disable
  max_connections: 25
  min_connections: 5

# Redis Configuration
redis:
  address: localhost:6379
  password: ""
  db: 0
  max_connections: 10

# Authentication Configuration
auth:
  jwt_secret: "your-secret-key"
  token_expiry: 24h
  refresh_token_expiry: 168h
  password_min_length: 8

# Job System Configuration
jobs:
  adapter: memory  # memory, redis, database
  queues: ["mailers", "payments", "processing", "notifications"]
  concurrency: 10
  max_retries: 5

# Logging Configuration
log_level: info
log_format: json

# Metrics and Tracing
metrics_enabled: true
tracing:
  enabled: true
  service_name: bazaruto
  endpoint: http://localhost:14268/api/traces
```

## Business Rules Configuration

Default file: `config/business_rules.json`

The business rules configuration provides dynamic, file-based configuration for all business logic components. This system is completely independent of the database and provides fast, in-memory access to configuration data.

### Configuration Structure

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

## Configuration Management

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

## Configuration Features

- **File-based Storage**: Configuration is stored in JSON files for easy editing and version control
- **In-memory Caching**: Fast access to configuration data without file I/O
- **Automatic Fallback**: Uses sensible defaults if configuration file is missing
- **Thread-safe**: Safe for concurrent access across multiple goroutines
- **Environment-specific**: Different configuration files for different environments
- **Hot Reloading**: Configuration can be updated without restarting the application

## Environment-specific Configuration

You can use different configuration files for different environments:

```bash
# Development
configManager := config.NewConfigManager(logger, "config/development.json")

# Staging
configManager := config.NewConfigManager(logger, "config/staging.json")

# Production
configManager := config.NewConfigManager(logger, "config/production.json")
```

## Configuration Validation

The configuration system includes built-in validation:

- Risk thresholds must be in ascending order
- Factor weights must sum to 1.0
- Time durations must be valid
- Required fields cannot be empty

## Environment Variables

You can override configuration values using environment variables:

```bash
export BAZARUTO_SERVER_ADDR=":9090"
export BAZARUTO_DB_HOST="production-db.example.com"
export BAZARUTO_REDIS_ADDRESS="production-redis.example.com:6379"
export BAZARUTO_LOG_LEVEL="debug"
```

## Troubleshooting

### Common Issues

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

4. **Configuration Validation Errors**
   - Check for invalid values in business rules
   - Verify required fields are present
   - Ensure proper data types

### Configuration Validation

The application provides detailed error messages for configuration validation failures. Common error messages include:

- `invalid server address format`
- `database host cannot be empty`
- `invalid SSL mode`
- `JWT secret cannot be empty`
- `fraud detection low threshold must be between 0 and 100`
- `fraud detection medium threshold must be >= low threshold`