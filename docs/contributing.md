# Contributing Guide

Thank you for your interest in contributing to the Bazaruto Insurance Platform! This guide will help you get started with contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contributing Process](#contributing-process)
- [Code Standards](#code-standards)
- [Testing](#testing)
- [Documentation](#documentation)
- [Pull Request Process](#pull-request-process)
- [Issue Guidelines](#issue-guidelines)

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you agree to uphold this code.

## Getting Started

### Prerequisites

- Go 1.22 or later
- Git
- Docker and Docker Compose
- PostgreSQL 14+
- Redis 6+
- Make

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/your-username/bazaruto.git
   cd bazaruto
   ```

3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/edsonmichaque/bazaruto.git
   ```

## Development Setup

### 1. Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
go install github.com/air-verse/air@latest
```

### 2. Set Up Environment

```bash
# Copy configuration template
cp config.yaml.example config.yaml

# Edit configuration
nano config.yaml
```

### 3. Start Dependencies

```bash
# Start PostgreSQL and Redis
docker-compose up -d postgres redis

# Run database migrations
make migrate
```

### 4. Run the Application

```bash
# Development mode with hot reload
make dev

# Or run directly
make run
```

## Contributing Process

### 1. Create a Branch

```bash
# Create a new branch for your feature
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/your-bug-description
```

### 2. Make Changes

- Write your code following the [Code Standards](#code-standards)
- Add tests for new functionality
- Update documentation as needed
- Ensure all tests pass

### 3. Commit Changes

Use [Conventional Commits](https://www.conventionalcommits.org/) format:

```bash
# Feature
git commit -m "feat: add webhook retry mechanism"

# Bug fix
git commit -m "fix: resolve memory leak in job processing"

# Documentation
git commit -m "docs: update API documentation"

# Refactoring
git commit -m "refactor: improve error handling"

# Performance
git commit -m "perf: optimize database queries"
```

### 4. Push and Create Pull Request

```bash
# Push your branch
git push origin feature/your-feature-name

# Create pull request on GitHub
```

## Code Standards

### Go Code Style

We follow the standard Go formatting and style guidelines:

```bash
# Format code
make fmt

# Lint code
make lint

# Run security scan
make security
```

### Code Organization

Follow the established project structure:

```
internal/
├── authentication/    # Authentication logic
├── authorization/     # Authorization logic
├── commands/          # CLI commands
├── config/           # Configuration management
├── events/           # Event bus system
├── handlers/         # HTTP handlers
├── job/              # Job system core
├── jobs/             # Job implementations
├── logger/           # Logging utilities
├── middleware/       # HTTP middleware
├── models/           # Domain models
├── services/         # Business logic
├── store/            # Data access layer
└── router/           # HTTP routing
```

### Naming Conventions

- **Packages**: lowercase, single word
- **Files**: snake_case
- **Types**: PascalCase
- **Functions**: PascalCase (exported), camelCase (private)
- **Variables**: camelCase
- **Constants**: PascalCase or UPPER_SNAKE_CASE

### Error Handling

```go
// Good: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// Good: Use custom error types
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation failed for field %s: %s", e.Field, e.Message)
}
```

### Logging

```go
// Use structured logging
logger.Info("User created successfully",
    zap.String("user_id", user.ID.String()),
    zap.String("email", user.Email))

logger.Error("Failed to process payment",
    zap.Error(err),
    zap.String("payment_id", payment.ID.String()))
```

### Database Operations

```go
// Use transactions for multiple operations
func (s *UserService) CreateUserWithProfile(ctx context.Context, user *models.User, profile *models.Profile) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(user).Error; err != nil {
            return fmt.Errorf("failed to create user: %w", err)
        }
        
        profile.UserID = user.ID
        if err := tx.Create(profile).Error; err != nil {
            return fmt.Errorf("failed to create profile: %w", err)
        }
        
        return nil
    })
}
```

## Testing

### Test Structure

```go
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        user    *models.User
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid user",
            user: &models.User{
                Email:    "test@example.com",
                FullName: "Test User",
            },
            wantErr: false,
        },
        {
            name: "invalid email",
            user: &models.User{
                Email:    "invalid-email",
                FullName: "Test User",
            },
            wantErr: true,
            errMsg:  "invalid email format",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Running Tests

```bash
# Run all tests
make test

# Run specific package tests
go test ./internal/services/...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run integration tests
go test ./test/integration/...

# Run E2E tests
go test ./test/e2e/...
```

### Test Requirements

- **Unit Tests**: > 80% coverage
- **Integration Tests**: Cover all external dependencies
- **E2E Tests**: Cover critical user journeys
- **Performance Tests**: For performance-critical code

## Documentation

### Code Documentation

```go
// UserService provides user management functionality
type UserService struct {
    store  store.UserStore
    logger *logger.Logger
}

// CreateUser creates a new user with the provided information.
// It validates the user data and returns an error if validation fails.
func (s *UserService) CreateUser(ctx context.Context, user *models.User) error {
    // Implementation
}
```

### API Documentation

Use OpenAPI/Swagger annotations:

```go
// CreateUser creates a new user
// @Summary Create user
// @Description Create a new user account
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.User true "User information"
// @Success 201 {object} models.User
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

### README Updates

Update relevant README files when adding new features:

- Main README.md
- Package-specific README files
- API documentation
- Configuration examples

## Pull Request Process

### Before Submitting

- [ ] Code follows style guidelines
- [ ] All tests pass
- [ ] Code is properly documented
- [ ] No security vulnerabilities
- [ ] Performance impact considered
- [ ] Breaking changes documented

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
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No breaking changes (or documented)

## Related Issues
Closes #123
```

### Review Process

1. **Automated Checks**
   - CI/CD pipeline runs
   - Code quality checks
   - Security scans
   - Test coverage

2. **Manual Review**
   - Code review by maintainers
   - Architecture review for major changes
   - Security review for sensitive changes

3. **Approval**
   - At least one maintainer approval required
   - All checks must pass
   - No unresolved discussions

## Issue Guidelines

### Bug Reports

Use the bug report template:

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

**Environment:**
- OS: [e.g. Ubuntu 20.04]
- Go version: [e.g. 1.22.0]
- Version: [e.g. 1.0.0]

**Additional context**
Any other context about the problem.
```

### Feature Requests

Use the feature request template:

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

### Issue Labels

- `bug`: Something isn't working
- `enhancement`: New feature or request
- `documentation`: Improvements or additions to documentation
- `good first issue`: Good for newcomers
- `help wanted`: Extra attention is needed
- `priority: high`: High priority
- `priority: medium`: Medium priority
- `priority: low`: Low priority

## Development Workflow

### Daily Workflow

1. **Start of day**
   ```bash
   git checkout main
   git pull upstream main
   ```

2. **Create feature branch**
   ```bash
   git checkout -b feature/your-feature
   ```

3. **Make changes and commit**
   ```bash
   git add .
   git commit -m "feat: add your feature"
   ```

4. **Push and create PR**
   ```bash
   git push origin feature/your-feature
   ```

### Weekly Workflow

1. **Update dependencies**
   ```bash
   go mod tidy
   go mod verify
   ```

2. **Run full test suite**
   ```bash
   make test-all
   ```

3. **Update documentation**
   - Review and update docs
   - Check for outdated information

## Performance Guidelines

### Database Performance

- Use indexes appropriately
- Avoid N+1 queries
- Use connection pooling
- Implement query timeouts

### Memory Management

- Avoid memory leaks
- Use proper resource cleanup
- Monitor memory usage
- Implement proper caching

### API Performance

- Implement proper pagination
- Use compression
- Cache frequently accessed data
- Optimize response sizes

## Security Guidelines

### Input Validation

```go
// Validate all inputs
func validateUser(user *models.User) error {
    if user.Email == "" {
        return errors.New("email is required")
    }
    
    if !isValidEmail(user.Email) {
        return errors.New("invalid email format")
    }
    
    return nil
}
```

### Authentication

- Use secure JWT secrets
- Implement proper token expiration
- Use HTTPS in production
- Implement rate limiting

### Data Protection

- Encrypt sensitive data
- Use parameterized queries
- Implement proper access controls
- Log security events

## Troubleshooting

### Common Issues

1. **Build failures**
   ```bash
   # Clean and rebuild
   make clean
   go mod download
   make build
   ```

2. **Test failures**
   ```bash
   # Run tests with verbose output
   go test -v ./...
   
   # Check test coverage
   go test -cover ./...
   ```

3. **Database issues**
   ```bash
   # Reset database
   make db-reset
   make migrate
   ```

### Getting Help

- Check existing issues and discussions
- Join our community chat
- Contact maintainers
- Read the documentation

## Recognition

Contributors are recognized in:

- CONTRIBUTORS.md file
- Release notes
- Project documentation
- Community acknowledgments

Thank you for contributing to the Bazaruto Insurance Platform! 🚀


