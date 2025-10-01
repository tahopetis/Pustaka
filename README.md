# Pustaka CI/CD Platform

A comprehensive CI/CD platform for modern software development teams, built with Go and Vue.js.

## 🚀 Features

- **CI Pipeline Management**: Create, monitor, and manage CI/CD pipelines
- **Real-time Updates**: Live status updates and notifications
- **Graph Visualization**: Interactive dependency graphs for pipeline relationships
- **User Management**: Role-based access control (RBAC) with authentication
- **Audit Logging**: Comprehensive audit trail for compliance and security
- **Multi-database Support**: PostgreSQL for primary data, Neo4j for relationships
- **Redis Caching**: Fast caching and session management
- **Docker Ready**: Complete containerization with Docker Compose

## 🏗️ Architecture

### Backend (Go)
- **Framework**: Chi v5 HTTP router
- **Database**: PostgreSQL with pgx v5 driver
- **Graph Database**: Neo4j for relationship mapping
- **Cache**: Redis for session and caching
- **Authentication**: JWT-based with role-based access control
- **Logging**: Structured logging with zerolog
- **Audit**: Comprehensive audit logging system

### Frontend (Vue.js)
- **Framework**: Vue.js 3 with TypeScript
- **UI Components**: Modern component library
- **State Management**: Pinia for state management
- **Routing**: Vue Router for navigation
- **HTTP Client**: Axios for API communication
- **Graph Visualization**: vis-network for interactive graphs

## 📋 Prerequisites

- Docker and Docker Compose
- Go 1.22+ (for local development)
- Node.js 18+ (for frontend development)
- Make (optional, for convenience commands)

## 🚀 Quick Start

### Using Docker Compose (Recommended)

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd pustaka
   ```

2. **Configure environment**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start the services**
   ```bash
   make dev
   ```
   This will:
   - Build and start all services (PostgreSQL, Neo4j, Redis, API, Frontend)
   - Run database migrations
   - Display service URLs and credentials

4. **Create admin user**
   ```bash
   make create-admin
   ```

5. **Access the application**
   - Frontend: http://localhost:3000
   - API: http://localhost:8080
   - Neo4j Browser: http://localhost:7474

### Manual Installation

1. **Install dependencies**
   ```bash
   make deps
   ```

2. **Start databases**
   ```bash
   docker-compose up -d postgres neo4j redis
   ```

3. **Run migrations**
   ```bash
   make migrate
   ```

4. **Start the API**
   ```bash
   make run
   ```

5. **Start the frontend** (in another terminal)
   ```bash
   cd web
   npm install
   npm run dev
   ```

## 📚 Available Commands

### Development Commands
```bash
make help          # Show all available commands
make dev           # Start full development environment
make run           # Run API server only
make build         # Build API binary
make test          # Run tests
make clean         # Clean build artifacts
```

### Docker Commands
```bash
make docker-build  # Build Docker images
make docker-up     # Start services with Docker
make docker-down   # Stop services
```

### Database Commands
```bash
make migrate       # Run database migrations
```

### Code Quality
```bash
make fmt           # Format Go code
make lint          # Run linter
make security      # Run security scan
```

### Production
```bash
make prod          # Start production environment
```

## 🔧 Configuration

### Environment Variables

Key environment variables (see `.env.example`):

```bash
# Database
DATABASE_URL=postgres://pustaka:password@localhost:5432/pustaka?sslmode=disable

# Neo4j
NEO4J_URL=bolt://localhost:7687
NEO4J_USERNAME=neo4j
NEO4J_PASSWORD=password

# Redis
REDIS_URL=redis://:redispassword@localhost:6379

# JWT
JWT_SECRET=your-super-secret-jwt-key
JWT_ACCESS_TOKEN_TTL=24h
JWT_REFRESH_TOKEN_TTL=168h

# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
```

### Database Setup

The application uses three databases:

1. **PostgreSQL**: Primary data storage
   - Host: localhost:5432
   - Database: pustaka
   - User: pustaka
   - Password: password (change in production)

2. **Neo4j**: Graph database for relationships
   - HTTP: localhost:7474
   - Bolt: localhost:7687
   - User: neo4j
   - Password: password (change in production)

3. **Redis**: Caching and sessions
   - Host: localhost:6379
   - Password: redispassword (change in production)

## 🔐 Authentication & Authorization

The platform implements role-based access control (RBAC):

### User Roles
- **admin**: Full access to all resources
- **user**: Standard user access
- **viewer**: Read-only access

### API Endpoints

#### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/refresh` - Refresh token
- `POST /api/v1/auth/logout` - User logout

#### CI Management
- `GET /api/v1/ci/pipelines` - List pipelines
- `POST /api/v1/ci/pipelines` - Create pipeline
- `GET /api/v1/ci/pipelines/{id}` - Get pipeline
- `PUT /api/v1/ci/pipelines/{id}` - Update pipeline
- `DELETE /api/v1/ci/pipelines/{id}` - Delete pipeline

#### Graph Visualization
- `GET /api/v1/graph` - Get graph data
- `GET /api/v1/graph/explore` - Explore graph relationships

#### Audit Logs
- `GET /api/v1/audit/logs` - List audit logs
- `GET /api/v1/audit/logs/{id}` - Get audit log
- `GET /api/v1/audit/stats` - Get audit statistics
- `GET /api/v1/audit/export` - Export audit logs
- `DELETE /api/v1/audit/logs/{id}` - Delete audit log
- `DELETE /api/v1/audit/cleanup` - Cleanup old logs

## 📊 Monitoring & Logging

### Health Checks
- API Health: http://localhost:8080/health
- Database health checks included in API health endpoint

### Logging
- Structured JSON logging
- Configurable log levels (debug, info, warn, error)
- Request/Response logging middleware
- Comprehensive audit logging

## 🧪 Testing

### Running Tests
```bash
# Run all tests
make test

# Run specific test package
go test ./internal/ci/...

# Run with coverage
go test -cover ./...
```

### Test Structure
- Unit tests for business logic
- Integration tests for database operations
- API endpoint tests
- Frontend component tests

## 🔒 Security

### Security Features
- JWT-based authentication
- Role-based authorization
- Input validation and sanitization
- SQL injection prevention
- CORS configuration
- Rate limiting
- Security headers
- Audit logging for compliance

### Security Scans
```bash
make security  # Run gosec security scanner
```

## 🚀 Production Deployment

### Production Checklist

1. **Environment Configuration**
   ```bash
   cp .env.example .env
   # Update all production values
   ```

2. **Security**
   - Change all default passwords
   - Use strong JWT secrets (32+ characters)
   - Update CORS origins to your domain
   - Enable HTTPS

3. **Database Security**
   - Use strong database passwords
   - Enable SSL/TLS for database connections
   - Configure database backups

4. **Start Production**
   ```bash
   make prod
   ```

### Docker Production

The Docker Compose configuration includes:
- Health checks for all services
- Proper restart policies
- Non-root users for security
- Resource limits and monitoring

## 📁 Project Structure

```
pustaka/
├── cmd/                    # Application entry points
│   ├── api/               # API server
│   ├── migrations/        # Database migrations
│   └── setup/            # Setup utilities
├── internal/              # Internal application code
│   ├── api/              # HTTP handlers and middleware
│   ├── auth/             # Authentication logic
│   ├── ci/               # CI pipeline management
│   ├── config/           # Configuration
│   ├── database/         # Database connections
│   └── models/           # Data models
├── pkg/                  # Public packages
├── web/                  # Frontend Vue.js application
├── docker/               # Docker configurations
├── docs/                 # Documentation
├── scripts/              # Utility scripts
├── docker-compose.yml    # Docker Compose configuration
├── Makefile             # Development commands
├── .env.example         # Environment template
└── README.md            # This file
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run tests and linting
6. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

If you encounter issues:

1. Check the logs: `docker-compose logs -f [service-name]`
2. Verify environment configuration
3. Check database connectivity
4. Review the troubleshooting section in the documentation

## 🔄 Version History

- **v1.0.0** - Initial release with core CI/CD functionality
  - CI pipeline management
  - User authentication and authorization
  - Graph visualization
  - Audit logging
  - Docker deployment