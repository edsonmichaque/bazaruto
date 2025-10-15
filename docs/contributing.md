# Contributing Guide

Thank you for your interest in contributing to the Bazaruto Insurance Platform! This guide will help you get started with contributing to the project.

## Getting Started

### Prerequisites

- Go 1.22 or later
- Git
- Docker and Docker Compose
- PostgreSQL 14+
- Redis 6+

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/your-username/bazaruto-insurance.git
   cd bazaruto-insurance
   ```

3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/edsonmichaque/bazaruto-insurance.git
   ```

## Development Setup

### 1. Install Dependencies

```bash
go mod download
```

### 2. Start Dependencies

```bash
docker-compose up -d postgres redis
```

### 3. Configure Application

```bash
cp config.yaml.example config.yaml
# Edit config.yaml with your settings
```

### 4. Run Migrations

```bash
go run cmd/bazarutod/main.go migrate
```

### 5. Start Development Server

```bash
go run cmd/bazarutod/main.go serve
```

## Contributing Process

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### 2. Make Changes

- Write clean, readable code
- Follow Go best practices
- Add tests for new functionality
- Update documentation as needed

### 3. Test Your Changes

```bash
# Run all tests
go test ./...

# Run specific tests
go test ./internal/services/...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 4. Commit Your Changes

```bash
git add .
git commit -m "feat: add new feature"
# or
git commit -m "fix: resolve bug in user service"
```

### 5. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a pull request on GitHub.

## Code Standards

### Go Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` and `goimports` for formatting
- Follow the project's existing code patterns
- Use meaningful variable and function names

### Project Structure

- Keep business logic in the `services` layer
- Use the repository pattern in the `store` layer
- Handle HTTP requests in the `handlers` layer
- Use middleware for cross-cutting concerns

### Error Handling

```go
// Good
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// Bad
if err != nil {
    return err
}
```

### Logging

```go
// Use structured logging
s.logger.Info("User created successfully",
    zap.String("user_id", user.ID.String()),
    zap.String("email", user.Email))
```

## Testing

### Unit Tests

```go
func TestUserService_CreateUser(t *testing.T) {
    // Arrange
    mockStore := &MockUserStore{}
    service := NewUserService(mockStore, logger, eventBus)
    
    // Act
    user, err := service.CreateUser(ctx, req)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
}
```

### Integration Tests

```go
func TestUserAPI_CreateUser(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    // Test API endpoint
    req := httptest.NewRequest("POST", "/users", body)
    w := httptest.NewRecorder()
    
    handler.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusCreated, w.Code)
}
```

### Test Coverage

- Aim for at least 80% test coverage
- Focus on business logic and critical paths
- Test error conditions and edge cases

## Documentation

### Code Documentation

- Document all public functions and types
- Use clear, concise comments
- Include examples for complex functions

```go
// CreateUser creates a new user with the provided information.
// It validates the input, checks for duplicates, and publishes
// a UserCreated event upon successful creation.
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*models.User, error) {
    // Implementation
}
```

### API Documentation

- Update API documentation for new endpoints
- Include request/response examples
- Document error responses

### README Updates

- Update README.md for significant changes
- Add new features to the features list
- Update installation instructions if needed

## Pull Request Process

### Before Submitting

1. **Run Tests**: Ensure all tests pass
2. **Check Linting**: Run `golangci-lint` and fix issues
3. **Update Documentation**: Update relevant documentation
4. **Rebase**: Rebase on latest main branch

### Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests added/updated
```

### Review Process

1. **Automated Checks**: CI/CD pipeline runs tests and linting
2. **Code Review**: At least one maintainer reviews the code
3. **Testing**: Reviewer tests the changes
4. **Approval**: Maintainer approves and merges

## Issue Guidelines

### Bug Reports

```markdown
**Describe the bug**
A clear description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '....'
3. See error

**Expected behavior**
What you expected to happen.

**Environment**
- OS: [e.g. Ubuntu 20.04]
- Go version: [e.g. 1.22.0]
- Application version: [e.g. v1.0.0]

**Additional context**
Any other context about the problem.
```

### Feature Requests

```markdown
**Is your feature request related to a problem?**
A clear description of what the problem is.

**Describe the solution you'd like**
A clear description of what you want to happen.

**Describe alternatives you've considered**
Alternative solutions or features you've considered.

**Additional context**
Any other context or screenshots about the feature request.
```

## Development Tools

### Recommended IDE

- **VS Code** with Go extension
- **GoLand** by JetBrains
- **Vim/Neovim** with vim-go

### Useful Commands

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Build application
go build -o bin/bazarutod cmd/bazarutod/main.go

# Run application
go run cmd/bazarutod/main.go serve
```

## Getting Help

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Documentation**: Check the `/docs` directory

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you agree to uphold this code.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.