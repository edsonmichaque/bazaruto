# Testing Guide

This directory contains comprehensive tests for the Bazaruto application.

## Test Structure

- `integration/` - Integration tests that test components working together
- `e2e/` - End-to-end tests that test the complete application workflow

## Running Tests

### Unit Tests
```bash
# Run all unit tests
go test ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Integration Tests
```bash
# Run integration tests
go test ./test/integration/...

# Run integration tests with verbose output
go test -v ./test/integration/...
```

### End-to-End Tests
```bash
# Start the application first
make run

# In another terminal, run E2E tests
go test ./test/e2e/...

# Or with custom base URL
E2E_BASE_URL=http://localhost:8080 go test ./test/e2e/...
```

## Test Environment Setup

### Prerequisites
- Go 1.22+
- PostgreSQL (for integration tests)
- Redis (for integration tests)

### Environment Variables
- `BAZARUTO_LOG_LEVEL` - Log level for tests (default: error)
- `BAZARUTO_LOG_FORMAT` - Log format for tests (default: json)
- `E2E_BASE_URL` - Base URL for E2E tests (default: http://localhost:8080)

## Test Categories

### Integration Tests
- **Product CRUD**: Tests complete product lifecycle
- **Quote Workflow**: Tests quote creation and management
- **Health Endpoint**: Tests health check functionality
- **Database Operations**: Tests database interactions

### End-to-End Tests
- **Health Check**: Tests health endpoint in real environment
- **Product Workflow**: Tests complete product workflow
- **Quote Workflow**: Tests complete quote workflow
- **Rate Limiting**: Tests rate limiting functionality

## Test Data

Tests use in-memory SQLite for integration tests and real services for E2E tests. Test data is created and cleaned up automatically.

## Continuous Integration

Tests are automatically run in CI/CD pipeline with:
- PostgreSQL and Redis services
- Code coverage reporting
- Security scanning
- Performance testing

## Writing Tests

### Integration Test Example
```go
func TestProductCRUD(t *testing.T) {
    server := SetupTestServer(t)
    
    // Test implementation
    // ...
}
```

### E2E Test Example
```go
func TestE2EProductWorkflow(t *testing.T) {
    client := NewE2ETestClient(baseURL)
    
    // Test implementation
    // ...
}
```

## Best Practices

1. **Use descriptive test names** that explain what is being tested
2. **Test both success and failure cases**
3. **Use table-driven tests** for multiple scenarios
4. **Clean up test data** after each test
5. **Use mocks** for external dependencies when appropriate
6. **Test error handling** and edge cases
7. **Use assertions** to verify expected behavior
8. **Keep tests independent** and isolated
