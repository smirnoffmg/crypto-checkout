.PHONY: help build test lint clean run docker-test docker-up docker-down

# Default target
help:
	@echo "Available commands:"
	@echo "  build       - Build the application"
	@echo "  test        - Run tests"
	@echo "  lint        - Run linters"
	@echo "  clean       - Clean build artifacts"
	@echo "  run         - Run the application"
	@echo "  dev         - Run with hot reload"
	@echo "  up   - Start all services with Docker Compose"
	@echo "  down - Stop all Docker Compose services"

# Build the application
build:
	go build -o bin/crypto-checkout ./cmd/crypto-checkout

# Run tests
test:
	go test -v -race -cover ./...

# Run tests in parallel
test-parallel:
	go test -v -race -cover -p 1 ./...

# Run linters
lint:
	golangci-lint run

fmt:
	go fmt ./...
	golangci-lint fmt

# Clean build artifacts
clean:
	rm -rf bin/
	go clean -cache

# Run the application
run: build
	./bin/crypto-checkout

up:
	@echo "Starting all services with Docker Compose..."
	docker compose up -d

down:
	@echo "Stopping all Docker Compose services..."
	docker compose down
	docker compose -f docker-compose.test.yml down
