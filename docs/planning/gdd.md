# Pustaka CMDB - Game Design Document (Technical Architecture)

**Version**: 1.0
**Date**: October 14, 2025
**Status**: Final
**Project Level**: Level 3 - Complex System

## Technical Architecture Overview

### System Philosophy
Pustaka CMDB follows a "Data-First, Relationship-Centric" architecture where:
- Data integrity is paramount (PostgreSQL as source of truth)
- Relationships are first-class citizens (Neo4j for graph operations)
- Performance is achieved through intelligent caching (Redis)
- User experience drives technical decisions

### Architecture Principles
1. **Separation of Concerns**: Clear boundaries between data, relationships, and cache
2. **API-First Design**: All functionality accessible via REST APIs
3. **Event-Driven Architecture**: Audit trails and notifications through events
4. **Scalability by Design**: Horizontal scaling and performance optimization
5. **Security by Default**: Built-in authentication, authorization, and audit

## Core Systems Design

### 1. Multi-Database Architecture

#### PostgreSQL (Primary Data Store)
**Purpose**: Structured data, ACID transactions, relational integrity
**Current Implementation**: ✅ Fully implemented with comprehensive schema
**Enhancement Strategy**:
```go
// Enhanced database connection pool
type DatabaseConfig struct {
    MaxOpenConns    int           `yaml:"max_open_conns"`
    MaxIdleConns    int           `yaml:"max_idle_conns"`
    ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
    ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
}

// Connection pooling with health checks
func NewPostgresDBWithHealthCheck(config DatabaseConfig) (*pgxpool.Pool, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    pool, err := pgxpool.New(ctx, config.URL)
    if err != nil {
        return nil, err
    }

    // Configure pool parameters
    pool.Config().MaxConns = int32(config.MaxOpenConns)
    pool.Config().MinConns = int32(config.MaxIdleConns)
    pool.Config().MaxConnLifetime = config.ConnMaxLifetime
    pool.Config().MaxConnIdleTime = config.ConnMaxIdleTime

    // Health check
    if err := pool.Ping(ctx); err != nil {
        return nil, fmt.Errorf("database health check failed: %w", err)
    }

    return pool, nil
}
```

**Schema Enhancements**:
- **Partitioning**: Time-based partitioning for audit logs
- **Indexing Strategy**: Advanced indexing for performance
- **JSONB Optimization**: Specialized indexes for attribute searches
- **Full-Text Search**: Enhanced search capabilities

#### Neo4j (Graph Database)
**Purpose**: Complex relationships, graph traversals, network analysis
**Current Implementation**: ✅ Implemented with basic relationship management
**Enhancement Strategy**:
```go
// Enhanced Neo4j service with connection pooling
type Neo4jService struct {
    driver      neo4j.DriverWithContext
    logger      *pustakaLogger.Logger
    queryCache  *sync.Map // Query result caching
    metrics     *Neo4jMetrics
}

// Advanced relationship queries
func (s *Neo4jService) GetImpactAnalysis(ctx context.Context, ciID uuid.UUID, depth int) (*ImpactAnalysis, error) {
    query := `
        MATCH path = (start:ConfigurationItem {id: $ciID})-[*1..%d]->(affected)
        RETURN start, collect(DISTINCT affected) as downstream,
               collect(DISTINCT relationships(path)) as relationships
    `

    params := map[string]interface{}{
        "ciID": ciID.String(),
    }

    result, err := s.executeQuery(ctx, fmt.Sprintf(query, depth), params)
    if err != nil {
        return nil, err
    }

    return s.parseImpactAnalysis(result)
}
```

**Graph Enhancements**:
- **Query Optimization**: Cypher query optimization and caching
- **Relationship Indexing**: Automatic index management
- **Path Finding**: Advanced graph algorithms
- **Performance Monitoring**: Query performance tracking

#### Redis (Cache & Session Store)
**Purpose**: Caching, session storage, rate limiting, job queues
**Current Implementation**: ✅ Basic caching implemented
**Enhancement Strategy**:
```go
// Enhanced Redis client with multi-level caching
type RedisCache struct {
    client      redis.Client
    l1Cache     *sync.Map // In-memory L1 cache
    l2TTL       time.Duration
    l3TTL       time.Duration
    metrics     *CacheMetrics
}

// Multi-level cache strategy
func (c *RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
    // L1 Cache (in-memory)
    if value, ok := c.l1Cache.Load(key); ok {
        c.metrics.L1Hits.Inc()
        return value, nil
    }

    // L2 Cache (Redis)
    result, err := c.client.Get(ctx, key).Result()
    if err == nil {
        c.metrics.L2Hits.Inc()
        var value interface{}
        json.Unmarshal([]byte(result), &value)

        // Promote to L1
        c.l1Cache.Store(key, value)
        return value, nil
    }

    c.metrics.Misses.Inc()
    return nil, redis.Nil
}
```

**Cache Enhancements**:
- **Multi-Level Caching**: L1 (memory) + L2 (Redis) + L3 (database)
- **Intelligent Invalidation**: Cache invalidation based on data changes
- **Compression**: Data compression for cache storage
- **Metrics**: Comprehensive cache performance metrics

### 2. API Architecture

#### RESTful API Design
**Current Implementation**: ✅ Comprehensive REST API with Chi v5
**Enhancement Strategy**:
```go
// Enhanced API router with middleware chaining
func setupRouter(config *Config, services *Services) *chi.Mux {
    r := chi.NewRouter()

    // Global middleware stack
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Recoverer)
    r.Use(middleware.Timeout(60 * time.Second))
    r.Use(middleware.Logger(config.Logging))
    r.Use(middleware.CORS(config.CORS))
    r.Use(middleware.RateLimiting(config.RateLimit))
    r.Use(middleware.Metrics(config.Metrics))

    // API versioning
    r.Route("/api/v1", func(r chi.Router) {
        setupAPIRoutes(r, services)
    })

    // Health checks
    r.Route("/health", func(r chi.Router) {
        r.Get("/", handlers.HealthCheck(services))
        r.Get("/detailed", handlers.DetailedHealthCheck(services))
    })

    return r
}

// Advanced middleware for rate limiting
func RateLimiting(config RateLimitConfig) func(http.Handler) http.Handler {
    limiter := rate.NewLimiter(rate.Limit(config.RequestsPerSecond), config.Burst)

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

**API Enhancements**:
- **Rate Limiting**: Advanced rate limiting with user-based limits
- **API Versioning**: Clean API versioning strategy
- **Documentation**: Interactive API documentation
- **Validation**: Enhanced request/response validation
- **Error Handling**: Standardized error responses

#### GraphQL Consideration
**Future Enhancement**: GraphQL endpoint for complex queries
```go
// GraphQL schema design (future)
type GraphQLResolver struct {
    services *Services
}

func (r *GraphQLResolver) ConfigurationItems(ctx context.Context, filter CIFilter) (*CIConnection, error) {
    return r.services.CIService.ListCIsWithFilter(ctx, filter)
}

func (r *GraphQLResolver) Relationships(ctx context.Context, ciID uuid.UUID) ([]*Relationship, error) {
    return r.services.CIService.GetCIRelationships(ctx, ciID)
}
```

### 3. Authentication & Authorization

#### JWT Authentication (Enhanced)
**Current Implementation**: ✅ JWT with access/refresh tokens
**Enhancement Strategy**:
```go
// Enhanced JWT service with advanced features
type JWTService struct {
    secret          []byte
    accessTokenTTL  time.Duration
    refreshTokenTTL time.Duration
    issuer          string
    tokenBlacklist  *redis.Client
    keyRotation     *KeyRotationManager
}

// Advanced token management
func (s *JWTService) GenerateTokenPair(userID uuid.UUID, roles []string) (*TokenPair, error) {
    now := time.Now()

    // Access token with enhanced claims
    accessClaims := jwt.MapClaims{
        "user_id":    userID.String(),
        "roles":      roles,
        "token_type": "access",
        "iat":        now.Unix(),
        "exp":        now.Add(s.accessTokenTTL).Unix(),
        "iss":        s.issuer,
        "jti":        uuid.New().String(), // JWT ID for revocation
    }

    // Refresh token with longer TTL
    refreshClaims := jwt.MapClaims{
        "user_id":    userID.String(),
        "token_type": "refresh",
        "iat":        now.Unix(),
        "exp":        now.Add(s.refreshTokenTTL).Unix(),
        "iss":        s.issuer,
        "jti":        uuid.New().String(),
    }

    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
    refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

    return &TokenPair{
        AccessToken:  accessToken.SignedString(s.secret),
        RefreshToken: refreshToken.SignedString(s.secret),
        ExpiresIn:    int(s.accessTokenTTL.Seconds()),
    }, nil
}
```

#### Enhanced RBAC System
**Current Implementation**: ✅ Basic RBAC with resource:action permissions
**Enhancement Strategy**:
```go
// Enhanced RBAC with attribute-based access control
type EnhancedRBACService struct {
    db          *pgxpool.Pool
    cache       *RedisCache
    logger      *pustakaLogger.Logger
    policies    *PolicyEngine
}

// Attribute-based access control
func (s *EnhancedRBACService) CheckPermission(ctx context.Context, userID uuid.UUID, permission string, resource map[string]interface{}) (bool, error) {
    // Check basic role permissions
    if hasPermission, err := s.checkBasicPermission(userID, permission); err != nil {
        return false, err
    } else if hasPermission {
        return true, nil
    }

    // Check attribute-based policies
    return s.policies.Evaluate(ctx, PolicyRequest{
        Subject: userID,
        Action:  permission,
        Resource: resource,
    })
}

// Policy engine for complex access control
type PolicyEngine struct {
    policies []Policy
    cache    *sync.Map
}

type Policy struct {
    Name        string                 `json:"name"`
    Effect      string                 `json:"effect"` // allow/deny
    Actions     []string               `json:"actions"`
    Resources   []string               `json:"resources"`
    Conditions  map[string]interface{} `json:"conditions"`
}
```

### 4. Frontend Architecture

#### Vue 3 Composition API (Enhanced)
**Current Implementation**: ✅ Vue 3 with Composition API and Pinia
**Enhancement Strategy**:
```typescript
// Enhanced composables for complex operations
export function useCIManagement() {
  const store = useCIStore()
  const { isLoading, error } = storeToRefs(store)

  // Advanced CI operations
  const bulkCreate = async (ciData: CreateCIData[]) => {
    try {
      isLoading.value = true
      const results = await Promise.allSettled(
        ciData.map(data => store.createCI(data))
      )

      const successful = results.filter(r => r.status === 'fulfilled')
      const failed = results.filter(r => r.status === 'rejected')

      return {
        successful: successful.length,
        failed: failed.length,
        results
      }
    } finally {
      isLoading.value = false
    }
  }

  // Advanced search with debouncing
  const searchCIs = useDebounceFn(async (query: string, filters: CIFilters) => {
    return await store.searchCIs(query, filters)
  }, 300)

  return {
    bulkCreate,
    searchCIs,
    isLoading,
    error
  }
}

// Enhanced state management with persistence
export const useEnhancedCIStore = defineStore('enhanced-ci', () => {
  const state = reactive({
    cis: [] as ConfigurationItem[],
    relationships: [] as Relationship[],
    filters: {} as CIFilters,
    pagination: {
      page: 1,
      limit: 20,
      total: 0
    }
  })

  // Persistence with local storage
  const persistToLocalStorage = () => {
    localStorage.setItem('ci-store-state', JSON.stringify({
      filters: state.filters,
      pagination: state.pagination
    }))
  }

  // Watch for changes and persist
  watch([state.filters, state.pagination], persistToLocalStorage, { deep: true })

  return {
    ...toRefs(state)
  }
})
```

#### Performance Optimizations
```typescript
// Virtual scrolling for large datasets
export function useVirtualScrolling<T>(items: Ref<T[]>, itemHeight: number) {
  const containerHeight = ref(0)
  const scrollTop = ref(0)

  const visibleItems = computed(() => {
    const start = Math.floor(scrollTop.value / itemHeight)
    const end = Math.min(
      start + Math.ceil(containerHeight.value / itemHeight),
      items.value.length
    )

    return items.value.slice(start, end).map((item, index) => ({
      item,
      index: start + index,
      top: (start + index) * itemHeight
    }))
  })

  return {
    visibleItems,
    containerHeight,
    scrollTop
  }
}

// Advanced graph visualization
export function useGraphVisualization() {
  const network = ref<any>(null)
  const container = ref<HTMLElement>()

  const initializeGraph = (data: GraphData) => {
    if (!container.value) return

    const options = {
      nodes: {
        shape: 'dot',
        size: 16,
        font: {
          size: 12,
          color: '#333'
        },
        borderWidth: 2
      },
      edges: {
        width: 2,
        color: {
          inherit: 'from'
        },
        smooth: {
          type: 'continuous'
        }
      },
      physics: {
        forceAtlas2Based: {
          gravitationalConstant: -26,
          centralGravity: 0.005,
          springLength: 230,
          springConstant: 0.18
        },
        maxVelocity: 146,
        solver: 'forceAtlas2Based',
        timestep: 0.35,
        stabilization: {
          enabled: true,
          iterations: 80,
          updateInterval: 25
        }
      }
    }

    network.value = new Network(container.value, data, options)
  }

  return {
    network,
    container,
    initializeGraph
  }
}
```

### 5. Data Model Enhancements

#### Flexible Schema System (Enhanced)
**Current Implementation**: ✅ JSONB-based flexible attributes with validation
**Enhancement Strategy**:
```go
// Enhanced CI type definitions with inheritance
type CITypeDefinition struct {
    ID                   uuid.UUID              `json:"id" db:"id"`
    Name                 string                 `json:"name" db:"name"`
    Description          *string                `json:"description,omitempty" db:"description"`
    ParentType           *string                `json:"parent_type,omitempty" db:"parent_type"`
    RequiredAttributes   []AttributeDefinition  `json:"required_attributes" db:"required_attributes"`
    OptionalAttributes   []AttributeDefinition  `json:"optional_attributes" db:"optional_attributes"`
    ValidationRules      []ValidationRule       `json:"validation_rules" db:"validation_rules"`
    LifecycleStates      []LifecycleState       `json:"lifecycle_states" db:"lifecycle_states"`
    DefaultTemplates     []CITemplate           `json:"default_templates" db:"default_templates"`
    IntegrationPoints    []IntegrationPoint     `json:"integration_points" db:"integration_points"`
    CreatedBy            uuid.UUID              `json:"created_by" db:"created_by"`
    CreatedAt            time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt            time.Time              `json:"updated_at" db:"updated_at"`
}

// Advanced validation with custom rules
type ValidationRule struct {
    Name        string                 `json:"name"`
    Type        string                 `json:"type"` // custom, script, regex
    Definition  map[string]interface{} `json:"definition"`
    ErrorMessage string                 `json:"error_message"`
}

// CI templates for rapid creation
type CITemplate struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Attributes  map[string]interface{} `json:"attributes"`
    Tags        []string               `json:"tags"`
}

// Enhanced attribute validation
func (ciType *CITypeDefinition) ValidateAttributesAdvanced(attributes map[string]interface{}) []ValidationError {
    var errors []ValidationError

    // Inherit validation from parent type
    if ciType.ParentType != nil {
        parentType, err := ciType.getParentType()
        if err == nil {
            parentErrors := parentType.ValidateAttributesAdvanced(attributes)
            errors = append(errors, parentErrors...)
        }
    }

    // Custom validation rules
    for _, rule := range ciType.ValidationRules {
        if validationErrors := rule.Validate(attributes); len(validationErrors) > 0 {
            errors = append(errors, validationErrors...)
        }
    }

    return errors
}
```

#### Relationship Management (Enhanced)
```go
// Enhanced relationship types with advanced features
type RelationshipType struct {
    ID                 uuid.UUID            `json:"id" db:"id"`
    Name               string               `json:"name" db:"name"`
    DisplayName        *string              `json:"display_name,omitempty" db:"display_name"`
    Description        *string              `json:"description,omitempty" db:"description"`
    ForwardLabel       string               `json:"forward_label" db:"forward_label"`
    ReverseLabel       string               `json:"reverse_label" db:"reverse_label"`
    Category           *string              `json:"category,omitempty" db:"category"`
    Color              *string              `json:"color,omitempty" db:"color"`
    Icon               *string              `json:"icon,omitempty" db:"icon"`

    // Cardinality and constraints
    CardinalitySource  string               `json:"cardinality_source" db:"cardinality_source"`  // one, many
    CardinalityTarget  string               `json:"cardinality_target" db:"cardinality_target"`  // one, many
    Bidirectional      bool                 `json:"bidirectional" db:"bidirectional"`
    AllowSelfReference bool                 `json:"allow_self_reference" db:"allow_self_reference"`

    // Type constraints
    AllowedSourceTypes []string             `json:"allowed_source_types" db:"allowed_source_types"`
    AllowedTargetTypes []string             `json:"allowed_target_types" db:"allowed_target_types"`

    // Advanced features
    Attributes         map[string]interface{} `json:"attributes" db:"attributes"`
    ValidationRules    []RelationshipRule   `json:"validation_rules" db:"validation_rules"`
    ImpactRules        []ImpactRule         `json:"impact_rules" db:"impact_rules"`

    // Metadata
    IsActive           bool                 `json:"is_active" db:"is_active"`
    IsSystem           bool                 `json:"is_system" db:"is_system"`
    CreatedAt          time.Time            `json:"created_at" db:"created_at"`
    UpdatedAt          *time.Time           `json:"updated_at,omitempty" db:"updated_at"`
    CreatedBy          uuid.UUID            `json:"created_by" db:"created_by"`
    UpdatedBy          *uuid.UUID           `json:"updated_by,omitempty" db:"updated_by"`
}

// Relationship validation rules
type RelationshipRule struct {
    Name        string                 `json:"name"`
    Condition   string                 `json:"condition"`
    Action      string                 `json:"action"` // prevent, warn, require
    Message     string                 `json:"message"`
}

// Impact analysis rules
type ImpactRule struct {
    Trigger     string                 `json:"trigger"`
    Analysis    string                 `json:"analysis"`
    Action      string                 `json:"action"`
}
```

### 6. Performance & Scalability

#### Database Optimization
```sql
-- Advanced indexing strategy
CREATE INDEX CONCURRENTLY idx_cis_composite
ON configuration_items(ci_type, created_at DESC)
WHERE created_at > NOW() - INTERVAL '1 year';

-- Partial indexes for active data
CREATE INDEX CONCURRENTLY idx_cis_active
ON configuration_items(name, ci_type)
WHERE updated_at > NOW() - INTERVAL '6 months';

-- JSONB expression indexes
CREATE INDEX CONCURRENTLY idx_cis_server_hostname
ON configuration_items USING GIN((attributes->'hostname'))
WHERE ci_type = 'Server';

-- Full-text search with trigram support
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX CONCURRENTLY idx_cis_name_trgm
ON configuration_items USING GIN(name gin_trgm_ops);
```

#### Caching Strategy
```go
// Multi-level cache implementation
type CacheManager struct {
    l1Cache *sync.Map        // In-memory
    l2Cache *RedisCache      // Redis
    l3Cache *PostgresCache   // Database query cache
}

func (c *CacheManager) Get(ctx context.Context, key string, query func() (interface{}, error)) (interface{}, error) {
    // Try L1 cache
    if value, ok := c.l1Cache.Load(key); ok {
        return value, nil
    }

    // Try L2 cache
    if value, err := c.l2Cache.Get(ctx, key); err == nil {
        c.l1Cache.Store(key, value)
        return value, nil
    }

    // Try L3 cache (database query cache)
    if value, err := c.l3Cache.Get(ctx, key); err == nil {
        c.l1Cache.Store(key, value)
        c.l2Cache.Set(ctx, key, value, 5*time.Minute)
        return value, nil
    }

    // Execute query and cache result
    value, err := query()
    if err != nil {
        return nil, err
    }

    c.l1Cache.Store(key, value)
    c.l2Cache.Set(ctx, key, value, 5*time.Minute)
    c.l3Cache.Set(ctx, key, value, 30*time.Minute)

    return value, nil
}
```

### 7. Monitoring & Observability

#### Metrics Collection
```go
// Comprehensive metrics collection
type MetricsCollector struct {
    httpRequests   *prometheus.CounterVec
    httpRequestDuration *prometheus.HistogramVec
    databaseQueries *prometheus.CounterVec
    cacheHits      *prometheus.CounterVec
    activeSessions prometheus.Gauge
    systemMetrics  *prometheus.GaugeVec
}

func NewMetricsCollector() *MetricsCollector {
    return &MetricsCollector{
        httpRequests: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "http_requests_total",
                Help: "Total number of HTTP requests",
            },
            []string{"method", "endpoint", "status"},
        ),
        httpRequestDuration: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "http_request_duration_seconds",
                Help: "HTTP request duration in seconds",
                Buckets: prometheus.DefBuckets,
            },
            []string{"method", "endpoint"},
        ),
        databaseQueries: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "database_queries_total",
                Help: "Total number of database queries",
            },
            []string{"database", "operation"},
        ),
        cacheHits: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "cache_hits_total",
                Help: "Total number of cache hits",
            },
            []string{"cache_level"},
        ),
        activeSessions: prometheus.NewGauge(
            prometheus.GaugeOpts{
                Name: "active_sessions",
                Help: "Number of active user sessions",
            },
        ),
        systemMetrics: prometheus.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "system_metrics",
                Help: "System metrics",
            },
            []string{"metric_type"},
        ),
    }
}
```

#### Health Checks
```go
// Comprehensive health check system
type HealthChecker struct {
    checks map[string]HealthCheck
}

type HealthCheck interface {
    Name() string
    Check(ctx context.Context) HealthStatus
}

type DatabaseHealthCheck struct {
    db *pgxpool.Pool
}

func (h *DatabaseHealthCheck) Check(ctx context.Context) HealthStatus {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    if err := h.db.Ping(ctx); err != nil {
        return HealthStatus{
            Status: "unhealthy",
            Message: fmt.Sprintf("Database ping failed: %v", err),
        }
    }

    // Check connection pool
    stats := h.db.Stat()
    if stats.AcquiredConns() > stats.MaxConns()*0.9 {
        return HealthStatus{
            Status: "degraded",
            Message: "Database connection pool near capacity",
        }
    }

    return HealthStatus{
        Status: "healthy",
        Message: "Database responding normally",
    }
}
```

### 8. Security Architecture

#### Enhanced Security Middleware
```go
// Comprehensive security middleware stack
func SecurityMiddleware(config SecurityConfig) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Security headers
            w.Header().Set("X-Content-Type-Options", "nosniff")
            w.Header().Set("X-Frame-Options", "DENY")
            w.Header().Set("X-XSS-Protection", "1; mode=block")
            w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
            w.Header().Set("Content-Security-Policy", config.CSP)

            // Rate limiting
            if !checkRateLimit(r, config.RateLimit) {
                http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
                return
            }

            // IP whitelisting/blacklisting
            if !checkIPAccess(r.RemoteAddr, config.IPAccess) {
                http.Error(w, "Access denied", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// Input validation and sanitization
func ValidationMiddleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Validate content type
            if r.Method == "POST" || r.Method == "PUT" {
                contentType := r.Header.Get("Content-Type")
                if !strings.Contains(contentType, "application/json") {
                    http.Error(w, "Invalid content type", http.StatusBadRequest)
                    return
                }
            }

            // Validate request size
            if r.ContentLength > config.MaxRequestSize {
                http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

### 9. Deployment Architecture

#### Container Strategy
```dockerfile
# Multi-stage build for production
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/api/main.go

# Production image
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080
CMD ["./main"]
```

#### Docker Compose for Development
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: pustaka
      POSTGRES_USER: pustaka
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./cmd/migrations:/docker-entrypoint-initdb.d
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U pustaka"]
      interval: 30s
      timeout: 10s
      retries: 3

  neo4j:
    image: neo4j:5
    environment:
      NEO4J_AUTH: neo4j/password
      NEO4J_PLUGINS: '["apoc", "graph-data-science"]'
    volumes:
      - neo4j_data:/data
      - neo4j_logs:/logs
    ports:
      - "7474:7474"
      - "7687:7687"
    healthcheck:
      test: ["CMD", "cypher-shell", "-u", "neo4j", "-p", "password", "RETURN 1"]
      interval: 30s
      timeout: 10s
      retries: 3

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

  api:
    build: .
    environment:
      DATABASE_URL: postgres://pustaka:password@postgres:5432/pustaka
      NEO4J_URI: bolt://neo4j:7687
      NEO4J_USERNAME: neo4j
      NEO4J_PASSWORD: password
      REDIS_URL: redis://redis:6379
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      neo4j:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  frontend:
    build: ./web
    ports:
      - "3000:3000"
    environment:
      VITE_API_URL: http://localhost:8080
    depends_on:
      - api

volumes:
  postgres_data:
  neo4j_data:
  neo4j_logs:
  redis_data:
```

### 10. Development Workflow

#### CI/CD Pipeline
```yaml
# .github/workflows/ci.yml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      neo4j:
        image: neo4j:5
        env:
          NEO4J_AUTH: neo4j/password
        options: >-
          --health-cmd cypher-shell -u neo4j -p password "RETURN 1"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: '1.21'

    - name: Set up Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '18'
        cache: 'npm'
        cache-dependency-path: web/package-lock.json

    - name: Install Go dependencies
      run: go mod download

    - name: Run Go tests
      run: go test -v -race -coverprofile=coverage.out ./...
      env:
        DATABASE_URL: postgres://postgres:postgres@localhost:5432/postgres
        NEO4J_URI: bolt://localhost:7687
        NEO4J_USERNAME: neo4j
        NEO4J_PASSWORD: password
        REDIS_URL: redis://localhost:6379

    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out

    - name: Install frontend dependencies
      working-directory: ./web
      run: npm ci

    - name: Run frontend tests
      working-directory: ./web
      run: npm run test

    - name: Build frontend
      working-directory: ./web
      run: npm run build

    - name: Build Go binary
      run: go build -o pustaka cmd/api/main.go

    - name: Run integration tests
      run: go test -v ./tests/integration/...
      env:
        DATABASE_URL: postgres://postgres:postgres@localhost:5432/postgres
        NEO4J_URI: bolt://localhost:7687
        NEO4J_USERNAME: neo4j
        NEO4J_PASSWORD: password
        REDIS_URL: redis://localhost:6379

    - name: Security scan
      run: |
        go install github.com/securecodewarrior/github-action-add-sarif@v1
        gosec -fmt sarif -out gosec.sarif ./...
        github-action-add-sarif gosec.sarif

    - name: Build Docker image
      if: github.ref == 'refs/heads/main'
      run: |
        docker build -t pustaka:${{ github.sha }} .
        docker tag pustaka:${{ github.sha }} pustaka:latest
```

## Implementation Phases

### Phase 1: Foundation Enhancement (Months 1-3)
- Database optimization and indexing
- Enhanced caching strategy
- API performance improvements
- Security enhancements
- Basic monitoring setup

### Phase 2: Feature Enhancement (Months 4-7)
- Bulk operations and advanced search
- Enhanced relationship management
- Improved user interface
- Advanced authentication/authorization
- Reporting and analytics foundation

### Phase 3: Advanced Features (Months 8-12)
- Workflow engine implementation
- Advanced analytics and dashboards
- Integration framework
- Performance optimization
- Production deployment preparation

---

*This GDD provides the technical foundation for implementing the PRD requirements. It serves as a guide for the development team and will be updated as the architecture evolves.*