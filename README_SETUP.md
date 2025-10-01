# 🚀 Pustaka CMDB Phase 1 Setup Guide

This guide will help you set up and run Pustaka CMDB Phase 1 on your local machine.

## 📋 Prerequisites

Before you begin, ensure you have the following installed:

- **Docker** (version 20.10 or later)
- **Docker Compose** (version 2.0 or later)
- **Go** (version 1.22 or later)
- **Node.js** (version 18 or later)
- **Make** (optional, for using Makefile commands)

## 🛠️ Quick Setup (Recommended)

The easiest way to get started is to use the automated setup script:

```bash
# Clone the repository
git clone <repository-url>
cd pustaka

# Run the setup script
./scripts/setup.sh
```

This script will:
- ✅ Check all prerequisites
- ✅ Set up environment variables
- ✅ Install all dependencies
- ✅ Start all Docker services
- ✅ Run database migrations
- ✅ Create an admin user
- ✅ Build and start the application

## 🏗️ Manual Setup

If you prefer to set up manually, follow these steps:

### 1. Environment Setup

```bash
# Copy environment file
cp .env.example .env

# Review and modify .env as needed
nano .env
```

### 2. Start Database Services

```bash
# Start PostgreSQL, Neo4j, and Redis
docker-compose up -d postgres neo4j redis

# Wait for services to be ready (this may take a minute)
docker-compose logs -f postgres neo4j redis
```

### 3. Install Dependencies

```bash
# Install Go dependencies
go mod download
go mod tidy

# Install Node.js dependencies
cd web
npm install
cd ..
```

### 4. Start Application

```bash
# Build and start API
go build -o bin/api ./cmd/api
./bin/api

# In another terminal, build and start frontend
cd web
npm run build
npm run dev
```

## 🌐 Access Points

Once setup is complete, you can access the application at:

- **Frontend**: http://localhost:3000
- **API**: http://localhost:8080
- **API Health Check**: http://localhost:8080/health
- **Neo4j Browser**: http://localhost:7474

## 🔑 Default Credentials

The setup creates an admin user with the following credentials:

- **Username**: `admin`
- **Password**: `admin123`

⚠️ **Important**: Change the default password immediately after first login!

## 🏛️ What's Included in Phase 1

### Backend Features
- ✅ **FSD-Compliant CI Management** with flexible JSONB attributes
- ✅ **CI Type Schema Management** with validation
- ✅ **Authentication & RBAC** (Admin, Editor, Viewer roles)
- ✅ **Relationship Management** with flexible attributes
- ✅ **Basic Graph Integration** with Neo4j
- ✅ **Audit Logging** for all operations
- ✅ **REST API** with comprehensive endpoints

### Frontend Features
- ✅ **Modern Vue.js 3** interface with Composition API
- ✅ **Authentication** with JWT tokens
- ✅ **CI Management** with dynamic forms
- ✅ **CI Type Management** interface
- ✅ **Basic Graph Visualization**
- ✅ **Responsive Design** with Tailwind CSS

### Database Features
- ✅ **PostgreSQL** with optimized schema
- ✅ **Neo4j** for graph relationships
- ✅ **Redis** for caching
- ✅ **Database Migrations** with default data

## 🛠️ Development Commands

### Docker Commands
```bash
# Start all services
docker-compose up -d

# Stop all services
docker-compose down

# View logs
docker-compose logs -f [service-name]

# Restart a service
docker-compose restart [service-name]
```

### Make Commands
```bash
# Build the application
make build

# Run tests
make test

# Run with hot reload
make dev

# Clean build artifacts
make clean

# Run database migrations
make migrate-up

# Create admin user
make create-admin
```

### Go Commands
```bash
# Run the API
go run ./cmd/api

# Run tests
go test ./...

# Build binary
go build -o bin/api ./cmd/api
```

### Frontend Commands
```bash
cd web

# Start development server
npm run dev

# Build for production
npm run build

# Run tests
npm run test

# Lint code
npm run lint
```

## 📊 Database Access

### PostgreSQL
```bash
# Connect to PostgreSQL
docker-compose exec postgres psql -U pustaka -d pustaka

# View tables
\dt

# View users
SELECT * FROM users;
```

### Neo4j
```bash
# Access Neo4j Browser
# URL: http://localhost:7474
# Username: neo4j
# Password: password
```

### Redis
```bash
# Connect to Redis
docker-compose exec redis redis-cli -a redispassword

# View keys
KEYS *
```

## 🔧 Configuration

### Environment Variables
Key environment variables in `.env`:

```bash
# Database
DATABASE_URL=postgres://pustaka:password@localhost:5432/pustaka?sslmode=disable

# Neo4j
NEO4J_URL=bolt://localhost:7687
NEO4J_USERNAME=neo4j
NEO4J_PASSWORD=password

# JWT
JWT_SECRET=your-secret-key-here

# API Server
SERVER_PORT=8080
```

### Database Schema
The database is initialized with:

- **Users and Roles** with RBAC permissions
- **Default CI Types**: Server, Application, Database
- **Audit Logging** tables
- **Indexes** for performance

## 🧪 Testing

### Backend Tests
```bash
# Run all tests
make test

# Run specific tests
go test ./internal/ci

# Run with coverage
go test -cover ./...
```

### Frontend Tests
```bash
cd web

# Run unit tests
npm run test

# Run with UI
npm run test:ui
```

## 🐛 Troubleshooting

### Common Issues

**1. Port Already in Use**
```bash
# Check what's using the port
lsof -i :8080
lsof -i :3000

# Kill the process
kill -9 <PID>
```

**2. Docker Services Not Starting**
```bash
# Check Docker status
docker ps
docker-compose ps

# View logs
docker-compose logs [service-name]

# Restart services
docker-compose down
docker-compose up -d
```

**3. Database Connection Issues**
```bash
# Check database connectivity
docker-compose exec postgres pg_isready -U pustaka
docker-compose exec neo4j cypher-shell -u neo4j -p password "RETURN 1"
```

**4. Permission Issues**
```bash
# Fix Docker permissions
sudo chown -R $USER:$USER .

# Make scripts executable
chmod +x scripts/*.sh
```

### Getting Help

If you encounter issues:

1. Check the logs: `docker-compose logs -f`
2. Review the troubleshooting steps above
3. Check the [documentation](./FSD.md, ./TSD.md)
4. Open an issue in the repository

## 🚀 Next Steps

After setup is complete:

1. **Login** to the web interface
2. **Explore** the default CI types (Server, Application, Database)
3. **Create** your first configuration items
4. **Build** relationships between CIs
5. **View** the graph visualization
6. **Check** the audit logs

## 📚 Documentation

- [Functional Specification](./FSD.md)
- [Technical Specification](./TSD.md)
- [Implementation Plan](./IMPLEMENTATION_PLAN.md)
- [API Specification](./RBAC_API_SPECIFICATION_PHASE1.md)
- [Coding Standards](./CODING_STANDARDS.md)
- [Security Checklist](./SECURITY_CHECKLIST.md)
- [Test Plan](./TEST_PLAN.md)

---

**Happy CMDB building! 🎉**