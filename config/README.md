# Configuration Files

This directory contains configuration files and examples for the Bazaruto Insurance Platform.

## Files

### Application Configuration

- `config.yaml.example` - Main application configuration template
- `development.yaml` - Development environment configuration
- `staging.yaml` - Staging environment configuration  
- `production.yaml` - Production environment configuration

### Business Rules Configuration

- `business_rules.json.example` - Business rules configuration template
- `business_rules.production.json` - Production business rules configuration

## Usage

### Development

For local development, copy the example files and modify as needed:

```bash
# Copy the main config
cp config.yaml.example config.yaml

# Copy business rules
cp config/business_rules.json.example config/business_rules.json
```

### Production

For production deployment, use environment-specific configurations:

```bash
# Use production config
cp config/production.yaml config.yaml
cp config/business_rules.production.json config/business_rules.json
```

### Environment Variables

Production configurations use environment variables for sensitive data:

```bash
# Database
export DATABASE_URL="postgres://user:pass@host:5432/db"

# Redis
export REDIS_URL="redis://host:6379"
export REDIS_PASSWORD="your-redis-password"

# Email
export EMAIL_PROVIDER="smtp"
export SMTP_HOST="smtp.example.com"
export SMTP_PORT="587"
export SMTP_USERNAME="your-username"
export SMTP_PASSWORD="your-password"
export EMAIL_FROM="noreply@bazaruto.com"

# Security
export CSRF_SECRET="your-csrf-secret"
export ALLOWED_ORIGINS="https://app.bazaruto.com,https://admin.bazaruto.com"

# Observability
export OTEL_EXPORTER_OTLP_ENDPOINT="http://jaeger:14268/api/traces"
```

## Configuration Structure

### Application Configuration

The main application configuration (`config.yaml`) includes:

- **Server**: HTTP server settings (address, timeouts)
- **Database**: Database connection and pool settings
- **Redis**: Redis connection settings (optional)
- **Rate Limiting**: API rate limiting configuration
- **Observability**: Logging, metrics, and tracing settings
- **Email**: Email service configuration
- **Webhooks**: Webhook delivery settings
- **Security**: CORS, CSRF, and security headers

### Business Rules Configuration

The business rules configuration (`business_rules.json`) includes:

- **Fraud Detection**: Risk assessment and fraud detection rules
- **Risk Assessment**: Customer risk evaluation criteria
- **Pricing**: Insurance product pricing rules and discounts
- **Underwriting**: Policy underwriting criteria and requirements
- **Commission**: Agent commission rates and tiers
- **Compliance**: KYC/AML requirements and data retention policies
- **Policy Lifecycle**: Policy management rules and timeframes
- **Claim Processing**: Claim processing rules and requirements

## Validation

The application validates configuration files on startup and provides detailed error messages for:

- Invalid configuration values
- Missing required fields
- Inconsistent settings
- Security misconfigurations

## Hot Reloading

Business rules configuration supports hot reloading without application restart:

```bash
# Update business rules
curl -X POST http://localhost:8080/v1/admin/config/reload
```

## Security

- Never commit actual configuration files with sensitive data
- Use environment variables for production secrets
- Regularly rotate secrets and API keys
- Monitor configuration changes in production
- Use least-privilege access for configuration management

## Troubleshooting

### Common Issues

1. **Database Connection**: Check DSN format and credentials
2. **Redis Connection**: Verify Redis is running and accessible
3. **Rate Limiting**: Adjust limits based on traffic patterns
4. **Email Delivery**: Check SMTP settings and credentials
5. **Business Rules**: Validate JSON syntax and required fields

### Configuration Validation

The application provides detailed validation errors:

```bash
# Check configuration
./bazarutod config validate

# Test database connection
./bazarutod config test-db

# Test Redis connection
./bazarutod config test-redis
```
