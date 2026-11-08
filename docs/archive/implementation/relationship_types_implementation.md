# Relationship Types Management Implementation

## Overview

This document describes the comprehensive backend implementation for relationship types management in the Pustaka CMDB. The implementation provides a robust system for defining, managing, and validating relationship types between Configuration Items (CIs).

## Architecture

### Components

1. **Database Schema** (`cmd/migrations/002_relationship_types.sql`)
   - PostgreSQL table for storing relationship type definitions
   - Support for cardinality, allowed CI types, and attribute schemas
   - System vs custom relationship types
   - Comprehensive indexing for performance

2. **Data Models** (`internal/ci/relationship_types.go`)
   - Go structs representing relationship types
   - Request/response DTOs for API operations
   - Validation methods and business logic
   - Usage statistics and summary models

3. **Repository Layer** (`internal/ci/relationship_types_repository.go`)
   - Data access methods for CRUD operations
   - Advanced filtering and pagination
   - Usage statistics and analytics
   - Cache-aware operations

4. **Service Layer** (`internal/ci/relationship_types_service.go`)
   - Business logic for relationship type management
   - Validation against CI type constraints
   - Circular dependency detection
   - Audit logging integration
   - Redis caching for performance

5. **Neo4j Integration** (`internal/ci/relationship_types_neo4j_integration.go`)
   - Synchronization with graph database
   - Relationship validation in Neo4j
   - Impact analysis and cycle detection
   - Graph statistics and analytics

6. **API Handlers** (`internal/api/relationship_type_handlers.go`)
   - RESTful endpoints for relationship type management
   - Request validation and error handling
   - Comprehensive test coverage
   - OpenAPI documentation ready

7. **Integration Layer** (`internal/ci/relationship_service_integration.go`)
   - Integration with existing relationship service
   - Enhanced relationship creation with validation
   - Compatibility checking between CIs and relationship types

## Default Relationship Types

The system includes 20 default relationship types:

| Name | Display Name | Category | Forward Label | Backward Label | Cardinality |
|------|--------------|----------|---------------|----------------|-------------|
| scope | Scope | structural | scopes | scoped by | one-to-many |
| hosts | Hosts | infrastructure | hosts | hosted by | one-to-many |
| runs_on | Runs On | runtime | runs on | runs | many-to-many |
| part_of | Part Of | structural | part of | contains | many-to-one |
| supports | Supports | functional | supports | supported by | many-to-many |
| used_by | Used By | functional | used by | uses | many-to-many |
| deployed_on | Deployed On | deployment | deployed on | deploys | many-to-one |
| depends_on | Depends On | dependency | depends on | required by | many-to-many |
| monitored_by | Monitored By | operational | monitored by | monitors | many-to-many |
| connected_to | Connected To | infrastructure | connected to | connected from | many-to-many |
| mounted_on | Mounted On | infrastructure | mounted on | mounts | many-to-one |
| powered_by | Powered By | infrastructure | powered by | powers | many-to-one |
| stored_in | Stored In | data | stored in | stores | many-to-many |
| assigned_to | Assigned To | organizational | assigned to | assigns | many-to-one |
| covers | Covers | functional | covers | covered by | many-to-many |
| delivers | Delivers | service | delivers | delivered by | many-to-many |
| manages | Manages | organizational | manages | managed by | many-to-many |
| owned_by | Owned By | organizational | owned by | owns | many-to-one |
| backed_up_by | Backed Up By | operational | backed up by | backs up | many-to-many |
| linked_to | Linked To | general | linked to | linked from | many-to-many |

## API Endpoints

### Relationship Type Management

- `POST /api/v1/relationship-types` - Create a new relationship type
- `GET /api/v1/relationship-types` - List relationship types with filtering
- `GET /api/v1/relationship-types/{id}` - Get a specific relationship type
- `PUT /api/v1/relationship-types/{id}` - Update a relationship type
- `DELETE /api/v1/relationship-types/{id}` - Delete a relationship type

### Utility Endpoints

- `GET /api/v1/relationship-types/active` - Get all active relationship types
- `GET /api/v1/relationship-types/usage` - Get usage statistics
- `GET /api/v1/relationship-types/statistics` - Get comprehensive statistics
- `GET /api/v1/relationship-types/validate` - Validate relationship compatibility

## Features

### Validation and Constraints

1. **CI Type Constraints**
   - Relationship types can specify allowed source and target CI types
   - Automatic validation during relationship creation
   - Prevents incompatible relationships

2. **Cardinality Enforcement**
   - Support for one-to-one, one-to-many, many-to-one, many-to-many
   - Validation during relationship creation
   - Prevents constraint violations

3. **Circular Dependency Detection**
   - Automatic detection for dependency relationships
   - Neo4j-based cycle detection
   - Warnings for potentially problematic relationships

4. **Attribute Schema Validation**
   - JSON schema for relationship attributes
   - Type validation, length constraints, and enum validation
   - Required and optional attribute support

### Performance Features

1. **Caching**
   - Redis caching for active relationship types
   - Cache invalidation on updates
   - Improved response times for frequently accessed data

2. **Indexing**
   - Database indexes on name, category, and activity status
   - Full-text search on display names and descriptions
   - GIN indexes for array fields

3. **Pagination**
   - Efficient pagination for large datasets
   - Configurable page sizes
   - Total count and page metadata

### Security and Audit

1. **Role-Based Access Control**
   - Granular permissions for relationship type operations
   - System relationship types protected from modification
   - Audit trail for all changes

2. **Audit Logging**
   - Comprehensive audit trail for all CRUD operations
   - User context and timestamps
   - Change details captured

## Migration Strategy

### Database Migration

1. **Run the Migration**
   ```bash
   # Install golang-migrate if not already installed
   # https://github.com/golang-migrate/migrate

   # Run the migration
   migrate -path cmd/migrations -database "YOUR_DATABASE_URL" up
   ```

2. **Verify Schema**
   ```sql
   -- Check that the table was created
   \d relationship_types

   -- Verify default data
   SELECT * FROM relationship_types WHERE is_system = true;
   ```

### Application Integration

1. **Update Service Initialization**
   ```go
   // In your main application setup
   relationshipTypeRepo := ci.NewRelationshipTypeRepository(db, logger)
   relationshipTypeService := ci.NewRelationshipTypeService(
       relationshipTypeRepo,
       ciRepo,
       redis,
       auditService,
       logger,
       db,
   )
   neo4jIntegration := ci.NewRelationshipTypesNeo4jIntegration(neo4jDriver, logger)

   // Update router initialization
   router := api.NewRouter(ciService, relationshipTypeService, logger)
   ```

2. **Update Relationship Creation**
   ```go
   // Enhanced relationship creation with validation
   integration := ci.NewRelationshipServiceIntegration(
       relationshipTypeService,
       neo4jIntegration,
   )

   relationship, err := integration.CreateRelationshipWithValidation(ctx, req, userID)
   ```

3. **Frontend Integration**
   - Update relationship creation forms to include relationship type selection
   - Add validation UI for relationship compatibility
   - Implement relationship type management interface

### Backward Compatibility

1. **Existing Relationships**
   - Existing relationships will continue to work
   - Relationship type validation is enforced for new relationships
   - Migration can be performed gradually

2. **API Compatibility**
   - Existing relationship endpoints remain unchanged
   - New relationship type endpoints are additive
   - No breaking changes to existing functionality

## Testing

### Unit Tests
```bash
# Run relationship type tests
go test ./internal/ci/relationship_types_test.go

# Run API handler tests
go test ./internal/api/relationship_type_handlers_test.go

# Run integration tests
go test ./internal/ci/relationship_types_integration_test.go
```

### Load Testing
```bash
# Test relationship type performance
go test -bench=BenchmarkListRelationshipTypes ./internal/api/

# Test validation performance
go test -bench=BenchmarkValidateRelationship ./internal/ci/
```

## Monitoring and Observability

### Metrics
- Relationship type creation/update/delete rates
- Validation success/failure rates
- Cache hit/miss ratios
- Database query performance

### Logging
- Structured logging for all operations
- Error tracking with context
- Performance metrics
- Audit trail logging

### Health Checks
- Database connectivity
- Redis connectivity
- Neo4j connectivity
- Service health endpoints

## Future Enhancements

1. **Advanced Validation**
   - Custom validation rules
   - Script-based validation
   - External validation service integration

2. **Relationship Type Templates**
   - Predefined templates for common patterns
   - Template inheritance
   - Dynamic template generation

3. **Visualization**
   - Relationship type diagram
   - Impact analysis visualization
   - Interactive relationship explorer

4. **Integration Features**
   - CMDB federation support
   - External relationship type import
   - REST API for external systems

## Troubleshooting

### Common Issues

1. **Migration Failures**
   - Check database connectivity
   - Verify user permissions
   - Review migration logs

2. **Validation Errors**
   - Check CI type definitions
   - Verify relationship type constraints
   - Review circular dependency warnings

3. **Performance Issues**
   - Check cache configuration
   - Review database indexes
   - Monitor query performance

### Debug Commands

```bash
# Check database schema
psql -d pustaka -c "\d relationship_types"

# Verify relationship types
psql -d pustaka -c "SELECT * FROM relationship_types ORDER BY name;"

# Check relationships without types
psql -d pustaka -c "SELECT * FROM relationships WHERE relationship_type NOT IN (SELECT name FROM relationship_types);"

# Cache status (Redis)
redis-cli GET "relationship_types:active"
```

## Conclusion

The relationship types management implementation provides a robust, scalable, and maintainable system for managing relationships in the Pustaka CMDB. It offers comprehensive validation, performance optimization, and integration capabilities while maintaining backward compatibility with existing functionality.

The modular design allows for easy extension and customization, while the comprehensive testing and documentation ensure reliability and maintainability.