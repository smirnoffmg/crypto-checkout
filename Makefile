.PHONY: help build test lint clean run up down logs ps test-e2e-kafka

# Default target
help:
	@echo "Available commands:"
	@echo "  build       - Build the application"
	@echo "  test        - Run tests"
	@echo "  lint        - Run linters"
	@echo "  clean       - Clean build artifacts"
	@echo "  run         - Run the application"
	@echo "  dev         - Run with hot reload"
	@echo "  up          - Start all services with Docker Compose"
	@echo "  down        - Stop all Docker Compose services"
	@echo "  logs        - Show logs for all services"
	@echo "  ps          - Show running containers"
	@echo "  test-e2e-kafka - Run Kafka integration E2E test"

# Build the application
build:
	go build -o bin/crypto-checkout ./cmd/crypto-checkout

# Run tests
test:
	gotestsum --hide-summary=output --format-icons=hivis

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
	docker compose --env-file env.dev up -d

down:
	@echo "Stopping all Docker Compose services..."
	docker compose --env-file env.dev down

logs:
	@echo "Showing logs for all services..."
	docker compose --env-file env.dev logs -f

ps:
	@echo "Showing running containers..."
	docker compose --env-file env.dev ps

# E2E tests
test-e2e-kafka:
	@echo "Running Kafka Integration E2E Test..."
	./test/e2e/kafka_integration_test.sh
