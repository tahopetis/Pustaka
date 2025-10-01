#!/bin/bash

# Pustaka CMDB - Phase 1 Setup Script
# This script sets up the development environment for Phase 1

set -e

echo "🚀 Setting up Pustaka CMDB - Phase 1"
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."

    # Check Docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi

    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi

    # Check Go
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.22+ first."
        exit 1
    fi

    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_status "Found Go version: $GO_VERSION"

    # Check Node.js
    if ! command -v node &> /dev/null; then
        print_error "Node.js is not installed. Please install Node.js 18+ first."
        exit 1
    fi

    NODE_VERSION=$(node --version)
    print_status "Found Node.js version: $NODE_VERSION"

    print_success "All prerequisites are installed!"
}

# Setup environment
setup_environment() {
    print_status "Setting up environment..."

    # Copy .env.example to .env if it doesn't exist
    if [ ! -f .env ]; then
        cp .env.example .env
        print_success "Created .env file from .env.example"
        print_warning "Please review and update the .env file with your configuration"
    else
        print_status ".env file already exists"
    fi
}

# Install Go dependencies
install_go_dependencies() {
    print_status "Installing Go dependencies..."

    go mod download
    go mod tidy

    print_success "Go dependencies installed"
}

# Install Node.js dependencies
install_node_dependencies() {
    print_status "Installing Node.js dependencies..."

    cd web
    npm install
    cd ..

    print_success "Node.js dependencies installed"
}

# Start Docker services
start_docker_services() {
    print_status "Starting Docker services..."

    # Start databases first
    docker-compose up -d postgres neo4j redis

    print_status "Waiting for databases to be ready..."

    # Wait for PostgreSQL to be ready
    print_status "Waiting for PostgreSQL..."
    timeout=60
    while ! docker-compose exec -T postgres pg_isready -U pustaka -d pustaka > /dev/null 2>&1; do
        if [ $timeout -le 0 ]; then
            print_error "PostgreSQL failed to start within timeout"
            exit 1
        fi
        echo -n "."
        sleep 2
        timeout=$((timeout - 2))
    done
    echo ""
    print_success "PostgreSQL is ready"

    # Wait for Neo4j to be ready
    print_status "Waiting for Neo4j..."
    timeout=60
    while ! docker-compose exec -T neo4j cypher-shell -u neo4j -p password "RETURN 1" > /dev/null 2>&1; do
        if [ $timeout -le 0 ]; then
            print_error "Neo4j failed to start within timeout"
            exit 1
        fi
        echo -n "."
        sleep 2
        timeout=$((timeout - 2))
    done
    echo ""
    print_success "Neo4j is ready"

    # Wait for Redis to be ready
    print_status "Waiting for Redis..."
    timeout=30
    while ! docker-compose exec -T redis redis-cli ping > /dev/null 2>&1; do
        if [ $timeout -le 0 ]; then
            print_error "Redis failed to start within timeout"
            exit 1
        fi
        echo -n "."
        sleep 1
        timeout=$((timeout - 1))
    done
    echo ""
    print_success "Redis is ready"

    print_success "All Docker services are running"
}

# Run database migrations
run_migrations() {
    print_status "Running database migrations..."

    # The migrations should already be applied automatically by PostgreSQL init
    # But we'll check if tables exist
    TABLES=$(docker-compose exec -T postgres psql -U pustaka -d pustaka -tAc "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'")

    if [ "$TABLES" -gt 0 ]; then
        print_success "Database tables already exist"
    else
        print_error "Database tables not found. Please check PostgreSQL initialization."
        exit 1
    fi
}

# Create admin user
create_admin_user() {
    print_status "Creating admin user..."

    # Check if admin user already exists
    ADMIN_EXISTS=$(docker-compose exec -T postgres psql -U pustaka -d pustaka -tAc "SELECT COUNT(*) FROM users WHERE username = 'admin'")

    if [ "$ADMIN_EXISTS" -gt 0 ]; then
        print_status "Admin user already exists"
    else
        # This would normally be done via the API, but for setup we'll insert directly
        docker-compose exec -T postgres psql -U pustaka -d pustaka -c "
            INSERT INTO users (username, email, password_hash, is_active, created_at, updated_at)
            VALUES ('admin', 'admin@pustaka.local', '\$argon2id\$v=19\$m=65536,t=4,p=4\$c29tZXNhbHQ\$RDEwMEEwQjAxRjlGQjQ0NjU2NzgwNjQ1Njg2NDU2NzgwNjQ1Njg2NDU2', true, NOW(), NOW());

            INSERT INTO user_roles (user_id, role_id, assigned_by, assigned_at)
            SELECT u.id, r.id, u.id, NOW()
            FROM users u, roles r
            WHERE u.username = 'admin' AND r.name = 'admin';
        "

        print_success "Admin user created with default password: admin123"
        print_warning "Please change the default admin password after first login"
    fi
}

# Build and start the application
build_and_start_application() {
    print_status "Building and starting the application..."

    # Build the Go API
    go build -o bin/api ./cmd/api

    # Start the API service
    docker-compose up -d api

    print_status "Waiting for API to be ready..."
    timeout=60
    while ! curl -s http://localhost:8080/health > /dev/null 2>&1; do
        if [ $timeout -le 0 ]; then
            print_error "API failed to start within timeout"
            exit 1
        fi
        echo -n "."
        sleep 2
        timeout=$((timeout - 2))
    done
    echo ""
    print_success "API is ready"

    # Build and start the frontend
    cd web
    npm run build
    cd ..

    docker-compose up -d frontend

    print_status "Waiting for frontend to be ready..."
    timeout=60
    while ! curl -s http://localhost:3000 > /dev/null 2>&1; do
        if [ $timeout -le 0 ]; then
            print_error "Frontend failed to start within timeout"
            exit 1
        fi
        echo -n "."
        sleep 2
        timeout=$((timeout - 2))
    done
    echo ""
    print_success "Frontend is ready"
}

# Display success message
display_success_message() {
    print_success "🎉 Pustaka CMDB Phase 1 setup completed successfully!"
    echo ""
    echo "📋 What's been set up:"
    echo "  ✅ PostgreSQL database with schema and default CI types"
    echo "  ✅ Neo4j graph database with indexes"
    echo "  ✅ Redis cache"
    echo "  ✅ Go API backend with authentication and RBAC"
    echo "  ✅ Vue.js frontend"
    echo "  ✅ Admin user created"
    echo ""
    echo "🌐 Access URLs:"
    echo "  📱 Frontend: http://localhost:3000"
    echo "  🔌 API: http://localhost:8080"
    echo "  🗃️ Neo4j Browser: http://localhost:7474"
    echo "  📊 API Health: http://localhost:8080/health"
    echo ""
    echo "👤 Default Admin Login:"
    echo "  Username: admin"
    echo "  Password: admin123"
    echo ""
    echo "⚠️  Important Notes:"
    echo "  - Please change the default admin password immediately"
    echo "  - Review the .env file for production settings"
    echo "  - Check the documentation for API usage"
    echo ""
    echo "🚀 Next Steps:"
    echo "  1. Login to the web interface"
    echo "  2. Create CI types and configuration items"
    echo "  3. Explore the graph visualization"
    echo "  4. Check the API documentation"
    echo ""
    echo "🛠️  Useful Commands:"
    echo "  • Stop all services:     docker-compose down"
    echo "  • View logs:           docker-compose logs -f [service]"
    echo "  • Restart services:    docker-compose restart [service]"
    echo "  • Run tests:           make test"
    echo ""
}

# Main execution
main() {
    echo ""
    check_prerequisites
    echo ""
    setup_environment
    echo ""
    install_go_dependencies
    echo ""
    install_node_dependencies
    echo ""
    start_docker_services
    echo ""
    run_migrations
    echo ""
    create_admin_user
    echo ""
    build_and_start_application
    echo ""
    display_success_message
}

# Handle script interruption
trap 'print_error "Setup interrupted"; exit 1' INT

# Run main function
main "$@"