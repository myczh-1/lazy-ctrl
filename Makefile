.PHONY: help install build start stop clean test proto

# Default target
help:
	@echo "Lazy Control System - Available commands:"
	@echo "  install    - Install all dependencies"
	@echo "  build      - Build all services"
	@echo "  start      - Start all services"
	@echo "  stop       - Stop all services"  
	@echo "  clean      - Clean build artifacts"
	@echo "  test       - Run tests"
	@echo "  proto      - Generate protobuf files"
	@echo "  docker     - Build and start with Docker"

# Install dependencies
install:
	@echo "Installing Go dependencies..."
	cd controller-agent && go mod tidy
	@echo "Installing config service dependencies..."
	cd config-service && npm install
	@echo "Installing frontend dependencies..."
	cd frontend-app && npm install
	@echo "Installing cloud service dependencies..."
	cd cloud-service && npm install

# Build all services
build:
	@echo "Building controller agent..."
	cd controller-agent && go build -o bin/controller-agent cmd/main.go
	@echo "Building config service..."
	cd config-service && npm run build
	@echo "Building frontend..."
	cd frontend-app && npm run build:h5

# Generate protobuf files
proto:
	@echo "Generating protobuf files..."
	cd controller-agent && protoc --go_out=. --go-grpc_out=. api/proto/controller.proto

# Start development servers
start:
	@echo "Starting all services in development mode..."
	@echo "Starting controller agent..."
	cd controller-agent && go run cmd/main.go --config configs/commands.json &
	@echo "Starting config service..."
	cd config-service && npm run dev &
	@echo "Starting cloud service..."
	cd cloud-service && npm run dev &
	@echo "All services started. Check logs for details."

# Stop all services
stop:
	@echo "Stopping all services..."
	pkill -f "controller-agent" || true
	pkill -f "config-service" || true
	pkill -f "cloud-service" || true

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf controller-agent/bin/
	rm -rf config-service/dist/
	rm -rf frontend-app/dist/
	rm -rf cloud-service/dist/
	rm -rf */node_modules/

# Run tests
test:
	@echo "Running Go tests..."
	cd controller-agent && go test ./...
	@echo "Running Node.js tests..."
	cd config-service && npm test
	cd cloud-service && npm test

# Docker commands
docker:
	@echo "Building and starting with Docker..."
	docker-compose up --build -d

docker-stop:
	@echo "Stopping Docker containers..."
	docker-compose down

docker-logs:
	@echo "Showing Docker logs..."
	docker-compose logs -f

# Development helpers
dev-controller:
	cd controller-agent && go run cmd/main.go --config configs/commands.json

dev-config:
	cd config-service && npm run dev

dev-frontend:
	cd frontend-app && npm run dev:h5

dev-cloud:
	cd cloud-service && npm run dev