# Pustaka CI/CD Platform Makefile

.PHONY: help build run test clean docker-build docker-up docker-down migrate fmt lint dev prod deps security create-admin

# Default target
help:
	@echo "Pustaka CI/CD Platform"
	@echo ""
	@echo "Available commands:"
	@echo "  build          Build the API binary"
	@echo "  run            Run the API server"
	@echo "  test           Run all tests"
	@echo "  clean          Clean build artifacts"
	@echo "  docker-build   Build Docker images"
	@echo "  docker-up      Start services with Docker Compose"
	@echo "  docker-down    Stop services with Docker Compose"
	@echo "  migrate        Run database migrations"
	@echo "  fmt            Format Go code"
	@echo "  lint           Run linter"
	@echo "  dev            Start development environment"
	@echo "  prod           Start production environment"
	@echo "  deps           Install dependencies"
	@echo "  security       Run security scan"
	@echo "  create-admin   Create initial admin user"

# Build the API binary
build:
	@echo "Building API binary..."
	go build -o bin/api cmd/api/main.go
	@echo "✅ Build complete"

# Run the API server
run:
	@echo "Starting API server..."
	go run cmd/api/main.go

# Run all tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean -cache
	@echo "✅ Clean complete"

# Build Docker images
# Sequential on purpose: the Go and Vite build stages each peak around 1-1.5 GB
# RAM, and running them in parallel OOMs on low-memory hosts.
docker-build:
	@echo "Building Docker images (sequential)..."
	docker-compose build
	@echo "✅ Docker build complete"

# Start services with Docker Compose
docker-up:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d
	@echo "✅ Services started"
	@echo "🌐 Frontend: http://localhost:3001"
	@echo "🚀 API: http://localhost:8081"
	@echo "🐘 PostgreSQL: localhost:5433"
	@echo "🕸️  Neo4j: http://localhost:7475"
	@echo "🔴 Redis: localhost:6380"

# Stop services with Docker Compose
docker-down:
	@echo "Stopping services..."
	docker-compose down
	@echo "✅ Services stopped"

# Run database migrations
migrate:
	@echo "Running database migrations..."
	@echo "Note: Make sure PostgreSQL is running before running migrations"
	@if command -v migrate >/dev/null 2>&1; then \
		migrate -path cmd/migrations -database "postgres://pustaka:password@localhost:5432/pustaka?sslmode=disable" up; \
	else \
		echo "❌ migrate command not found. Please install golang-migrate: https://github.com/golang-migrate/migrate"; \
	fi

# Format Go code
fmt:
	@echo "Formatting Go code..."
	go fmt ./...
	@echo "✅ Code formatted"

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "❌ golangci-lint not found. Please install: https://golangci-lint.run/"; \
	fi

# Start development environment
dev: docker-up migrate
	@echo "🚀 Development environment ready!"
	@echo ""
	@echo "Services:"
	@echo "  Frontend: http://localhost:3000"
	@echo "  API: http://localhost:8080"
	@echo "  API Health: http://localhost:8080/health"
	@echo "  Neo4j Browser: http://localhost:7474"
	@echo ""
	@echo "Default credentials:"
	@echo "  PostgreSQL: pustaka/password"
	@echo "  Neo4j: neo4j/password"
	@echo ""
	@echo "To view logs: docker-compose logs -f [service-name]"
	@echo "To stop: make docker-down"

# Start production environment
prod:
	@echo "🚀 Starting production environment..."
	@if [ ! -f .env ]; then \
		echo "❌ .env file not found. Copy .env.example to .env and update values."; \
		exit 1; \
	fi
	docker-compose -f docker-compose.yml --env-file .env up -d
	@echo "✅ Production environment started"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "✅ Dependencies installed"

# Security scan
security:
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "❌ gosec not found. Please install: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Create initial admin user
create-admin:
	@echo "Creating initial admin user..."
	@echo "Note: Make sure the API is running before executing this"
	curl -X POST http://localhost:8080/api/v1/users/admin \
		-H "Content-Type: application/json" \
		-d '{
			"username": "admin",
			"email": "admin@example.com",
			"password": "admin123",
			"role": "admin"
		}'
	@echo ""
	@echo "✅ Admin user created (username: admin, password: admin123)"
	@echo "⚠️  Please change the default password after first login"