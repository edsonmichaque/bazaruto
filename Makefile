.PHONY: run migrate tidy lint test build clean docker

# Default target
all: tidy lint test build

# Run the application
run:
	@go run ./cmd/bazarutod serve

# Run database migrations
migrate:
	@go run ./cmd/bazarutod db migrate

# Reset database (dangerous!)
reset-db:
	@go run ./cmd/bazarutod db reset

# Seed admin data
seed:
	@go run ./cmd/bazarutod admin seed

# Validate configuration
lint-config:
	@go run ./cmd/bazarutod lint

# Tidy dependencies
tidy:
	@go mod tidy

# Run linter
lint:
	@golangci-lint run

# Run tests
test:
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

# Build binary with version info
build:
	@go build -ldflags "-X github.com/edsonmichaque/bazaruto/internal/version.Version=$(shell git describe --tags --always --dirty) -X github.com/edsonmichaque/bazaruto/internal/version.Commit=$(shell git rev-parse HEAD) -X github.com/edsonmichaque/bazaruto/internal/version.BuildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" -o bin/bazarutod ./cmd/bazarutod

# Build for multiple platforms with version info
build-all:
	@GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/edsonmichaque/bazaruto/internal/version.Version=$(shell git describe --tags --always --dirty) -X github.com/edsonmichaque/bazaruto/internal/version.Commit=$(shell git rev-parse HEAD) -X github.com/edsonmichaque/bazaruto/internal/version.BuildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" -o bin/bazarutod-linux-amd64 ./cmd/bazarutod
	@GOOS=darwin GOARCH=amd64 go build -ldflags "-X github.com/edsonmichaque/bazaruto/internal/version.Version=$(shell git describe --tags --always --dirty) -X github.com/edsonmichaque/bazaruto/internal/version.Commit=$(shell git rev-parse HEAD) -X github.com/edsonmichaque/bazaruto/internal/version.BuildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" -o bin/bazarutod-darwin-amd64 ./cmd/bazarutod
	@GOOS=darwin GOARCH=arm64 go build -ldflags "-X github.com/edsonmichaque/bazaruto/internal/version.Version=$(shell git describe --tags --always --dirty) -X github.com/edsonmichaque/bazaruto/internal/version.Commit=$(shell git rev-parse HEAD) -X github.com/edsonmichaque/bazaruto/internal/version.BuildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" -o bin/bazarutod-darwin-arm64 ./cmd/bazarutod
	@GOOS=windows GOARCH=amd64 go build -ldflags "-X github.com/edsonmichaque/bazaruto/internal/version.Version=$(shell git describe --tags --always --dirty) -X github.com/edsonmichaque/bazaruto/internal/version.Commit=$(shell git rev-parse HEAD) -X github.com/edsonmichaque/bazaruto/internal/version.BuildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" -o bin/bazarutod-windows-amd64.exe ./cmd/bazarutod

# Clean build artifacts
clean:
	@rm -rf bin/
	@rm -f coverage.out coverage.html

# Build Docker image
docker:
	@docker build -t bazaruto:latest .

# Run with Docker Compose
docker-up:
	@docker-compose -f deploy/docker-compose.yaml up -d

# Stop Docker Compose
docker-down:
	@docker-compose -f deploy/docker-compose.yaml down

# Cache management
cache-inspect:
	@go run ./cmd/bazarutod cache inspect

cache-purge:
	@go run ./cmd/bazarutod cache purge

cache-warm:
	@go run ./cmd/bazarutod cache warm

# Development helpers
dev-setup: tidy
	@echo "Setting up development environment..."
	@docker-compose -f deploy/docker-compose.yaml up -d postgres redis
	@sleep 5
	@make migrate
	@make seed

dev-clean:
	@docker-compose -f deploy/docker-compose.yaml down -v
	@make clean
