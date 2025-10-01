package ci

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

type Service struct {
	repo   *Repository
	neo4j  *Neo4jService
	redis  *redis.Client
	logger *pustakaLogger.Logger
}

func NewService(db *Repository, neo4j *Neo4jService, redis *redis.Client, logger *pustakaLogger.Logger) *Service {
	return &Service{
		repo:   db,
		neo4j:  neo4j,
		redis:  redis,
		logger: logger,
	}
}

// CI Operations

func (s *Service) CreateCI(ctx context.Context, req *CreateCIRequest, userID uuid.UUID) (*ConfigurationItem, error) {
	// Validate CI type exists
	ciType, err := s.repo.GetCITypeByName(ctx, req.CIType)
	if err != nil {
		s.logger.ErrorService("ci", "create_ci", err, map[string]interface{}{
			"ci_type": req.CIType,
			"user_id": userID,
		})
		return nil, fmt.Errorf("CI type '%s' does not exist", req.CIType)
	}

	// Validate attributes against schema
	validationErrors := ciType.ValidateAttributes(req.Attributes)
	if len(validationErrors) > 0 {
		return nil, ServiceValidationError{
			Message: "Attribute validation failed",
			Errors:  validationErrors,
		}
	}

	// Check for duplicate name within type
	existing, err := s.repo.GetCIByNameAndType(ctx, req.Name, req.CIType)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("CI with name '%s' already exists for type '%s'", req.Name, req.CIType)
	}

	// Create CI
	ci := &ConfigurationItem{
		Name:      req.Name,
		CIType:    req.CIType,
		Attributes: req.Attributes,
		Tags:      req.Tags,
		CreatedBy: userID,
	}

	result, err := s.repo.CreateCI(ctx, ci)
	if err != nil {
		return nil, err
	}

	// Sync to Neo4j
	if err := s.neo4j.SyncCI(ctx, result); err != nil {
		s.logger.ErrorService("neo4j", "sync_ci", err, map[string]interface{}{
			"ci_id": result.ID,
		})
		// Log error but don't fail the operation
	}

	// Invalidate cache
	s.invalidateCICache(ctx, result.ID)

	// Log audit event
	s.logAuditEvent(ctx, "ci", result.ID.String(), "create", userID.String(), map[string]interface{}{
		"ci_name": result.Name,
		"ci_type": result.CIType,
	})

	s.logger.InfoService("ci", "create_ci", map[string]interface{}{
		"ci_id":   result.ID,
		"ci_name": result.Name,
		"user_id": userID,
	})

	return result, nil
}

func (s *Service) GetCI(ctx context.Context, id uuid.UUID) (*ConfigurationItem, error) {
	// Try cache first
	if ci, err := s.getCIFromCache(ctx, id); err == nil {
		return ci, nil
	}

	// Get from database
	ci, err := s.repo.GetCI(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result
	s.cacheCI(ctx, ci)

	return ci, nil
}

func (s *Service) ListCIs(ctx context.Context, filters ListCIFilters, page, limit int) (*CIListResponse, error) {
	return s.repo.ListCIs(ctx, filters, page, limit)
}

func (s *Service) UpdateCI(ctx context.Context, id uuid.UUID, req *UpdateCIRequest, userID uuid.UUID) (*ConfigurationItem, error) {
	// Get current CI
	current, err := s.repo.GetCI(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get CI type for validation
	ciType, err := s.repo.GetCITypeByName(ctx, current.CIType)
	if err != nil {
		return nil, fmt.Errorf("CI type '%s' does not exist", current.CIType)
	}

	// Prepare updated attributes
	updatedAttributes := current.Attributes
	if req.Attributes != nil {
		updatedAttributes = req.Attributes
	}

	// Validate attributes against schema
	validationErrors := ciType.ValidateAttributes(updatedAttributes)
	if len(validationErrors) > 0 {
		return nil, ServiceValidationError{
			Message: "Attribute validation failed",
			Errors:  validationErrors,
		}
	}

	// Update CI
	result, err := s.repo.UpdateCI(ctx, id, req, userID)
	if err != nil {
		return nil, err
	}

	// Sync to Neo4j
	if err := s.neo4j.UpdateCI(ctx, result); err != nil {
		s.logger.ErrorService("neo4j", "update_ci", err, map[string]interface{}{
			"ci_id": result.ID,
		})
		// Log error but don't fail the operation
	}

	// Invalidate cache
	s.invalidateCICache(ctx, id)

	// Log audit event
	s.logAuditEvent(ctx, "ci", id.String(), "update", userID.String(), map[string]interface{}{
		"ci_name": result.Name,
		"ci_type": result.CIType,
		"changes": req,
	})

	s.logger.InfoService("ci", "update_ci", map[string]interface{}{
		"ci_id":   id,
		"user_id": userID,
	})

	return result, nil
}

func (s *Service) DeleteCI(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// Get CI for audit
	ci, err := s.repo.GetCI(ctx, id)
	if err != nil {
		return err
	}

	// Check for relationships
	relationships, err := s.neo4j.GetCIRelationships(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check relationships: %w", err)
	}

	if len(relationships) > 0 {
		return fmt.Errorf("cannot delete CI with existing relationships")
	}

	// Delete from database
	if err := s.repo.DeleteCI(ctx, id); err != nil {
		return err
	}

	// Delete from Neo4j
	if err := s.neo4j.DeleteCI(ctx, id); err != nil {
		s.logger.ErrorService("neo4j", "delete_ci", err, map[string]interface{}{
			"ci_id": id,
		})
		// Log error but don't fail the operation
	}

	// Invalidate cache
	s.invalidateCICache(ctx, id)

	// Log audit event
	s.logAuditEvent(ctx, "ci", id.String(), "delete", userID.String(), map[string]interface{}{
		"ci_name": ci.Name,
		"ci_type": ci.CIType,
	})

	s.logger.InfoService("ci", "delete_ci", map[string]interface{}{
		"ci_id":   id,
		"ci_name": ci.Name,
		"user_id": userID,
	})

	return nil
}

// CI Type Operations

func (s *Service) CreateCIType(ctx context.Context, req *CreateCITypeRequest, userID uuid.UUID) (*CITypeDefinition, error) {
	// Validate CI type name doesn't exist
	existing, err := s.repo.GetCITypeByName(ctx, req.Name)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("CI type '%s' already exists", req.Name)
	}

	// Validate schema
	if err := s.validateCITypeSchema(req); err != nil {
		return nil, err
	}

	ciType := &CITypeDefinition{
		Name:                req.Name,
		Description:         req.Description,
		RequiredAttributes: req.RequiredAttributes,
		OptionalAttributes: req.OptionalAttributes,
		CreatedBy:           userID,
	}

	result, err := s.repo.CreateCIType(ctx, ciType)
	if err != nil {
		return nil, err
	}

	// Log audit event
	s.logAuditEvent(ctx, "ci_type", result.ID.String(), "create", userID.String(), map[string]interface{}{
		"ci_type_name": result.Name,
	})

	s.logger.InfoService("ci_type", "create_ci_type", map[string]interface{}{
		"ci_type_id":   result.ID,
		"ci_type_name": result.Name,
		"user_id":      userID,
	})

	return result, nil
}

func (s *Service) GetCIType(ctx context.Context, id uuid.UUID) (*CITypeDefinition, error) {
	return s.repo.GetCIType(ctx, id)
}

func (s *Service) ListCITypes(ctx context.Context, page, limit int, search string) (*CITypeListResponse, error) {
	return s.repo.ListCITypes(ctx, page, limit, search)
}

func (s *Service) UpdateCIType(ctx context.Context, id uuid.UUID, req *UpdateCITypeRequest, userID uuid.UUID) (*CITypeDefinition, error) {
	// Validate schema if provided
	if req.RequiredAttributes != nil || req.OptionalAttributes != nil {
		if err := s.validateCITypeSchemaUpdate(req); err != nil {
			return nil, err
		}
	}

	result, err := s.repo.UpdateCIType(ctx, id, req)
	if err != nil {
		return nil, err
	}

	// Log audit event
	s.logAuditEvent(ctx, "ci_type", id.String(), "update", userID.String(), map[string]interface{}{
		"ci_type_name": result.Name,
		"changes":      req,
	})

	s.logger.InfoService("ci_type", "update_ci_type", map[string]interface{}{
		"ci_type_id": id,
		"user_id":     userID,
	})

	return result, nil
}

func (s *Service) DeleteCIType(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// Get CI type for audit
	ciType, err := s.repo.GetCIType(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteCIType(ctx, id); err != nil {
		return err
	}

	// Log audit event
	s.logAuditEvent(ctx, "ci_type", id.String(), "delete", userID.String(), map[string]interface{}{
		"ci_type_name": ciType.Name,
	})

	s.logger.InfoService("ci_type", "delete_ci_type", map[string]interface{}{
		"ci_type_id":   id,
		"ci_type_name": ciType.Name,
		"user_id":      userID,
	})

	return nil
}

// Relationship Operations

func (s *Service) CreateRelationship(ctx context.Context, req *CreateRelationshipRequest, userID uuid.UUID) (*Relationship, error) {
	// Validate CIs exist
	sourceCI, err := s.repo.GetCI(ctx, req.SourceID)
	if err != nil {
		return nil, fmt.Errorf("source CI not found: %w", err)
	}

	targetCI, err := s.repo.GetCI(ctx, req.TargetID)
	if err != nil {
		return nil, fmt.Errorf("target CI not found: %w", err)
	}

	// Check for circular dependency
	if req.SourceID == req.TargetID {
		return nil, fmt.Errorf("cannot create self-referencing relationship")
	}

	// Create relationship
	relationship := &Relationship{
		SourceID:        req.SourceID,
		TargetID:        req.TargetID,
		RelationshipType: req.RelationshipType,
		Attributes:      req.Attributes,
		CreatedBy:       userID,
	}

	result, err := s.repo.CreateRelationship(ctx, relationship)
	if err != nil {
		return nil, err
	}

	// Sync to Neo4j
	if err := s.neo4j.CreateRelationship(ctx, result, sourceCI, targetCI); err != nil {
		s.logger.ErrorService("neo4j", "create_relationship", err, map[string]interface{}{
			"relationship_id": result.ID,
		})
		// Log error but don't fail the operation
	}

	// Log audit event
	s.logAuditEvent(ctx, "relationship", result.ID.String(), "create", userID.String(), map[string]interface{}{
		"source_id":         req.SourceID,
		"target_id":         req.TargetID,
		"relationship_type": req.RelationshipType,
	})

	s.logger.InfoService("relationship", "create_relationship", map[string]interface{}{
		"relationship_id": result.ID,
		"user_id":         userID,
	})

	return result, nil
}

func (s *Service) GetRelationship(ctx context.Context, id uuid.UUID) (*Relationship, error) {
	return s.repo.GetRelationship(ctx, id)
}

func (s *Service) ListRelationships(ctx context.Context, filters ListRelationshipFilters, page, limit int) (*RelationshipListResponse, error) {
	return s.repo.ListRelationships(ctx, filters, page, limit)
}

func (s *Service) UpdateRelationship(ctx context.Context, id uuid.UUID, req *UpdateRelationshipRequest, userID uuid.UUID) (*Relationship, error) {
	result, err := s.repo.UpdateRelationship(ctx, id, req, userID)
	if err != nil {
		return nil, err
	}

	// Sync to Neo4j
	if err := s.neo4j.UpdateRelationship(ctx, result); err != nil {
		s.logger.ErrorService("neo4j", "update_relationship", err, map[string]interface{}{
			"relationship_id": id,
		})
		// Log error but don't fail the operation
	}

	// Log audit event
	s.logAuditEvent(ctx, "relationship", id.String(), "update", userID.String(), map[string]interface{}{
		"changes": req,
	})

	return result, nil
}

func (s *Service) DeleteRelationship(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// Get relationship for audit
	relationship, err := s.repo.GetRelationship(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteRelationship(ctx, id); err != nil {
		return err
	}

	// Delete from Neo4j
	if err := s.neo4j.DeleteRelationship(ctx, id); err != nil {
		s.logger.ErrorService("neo4j", "delete_relationship", err, map[string]interface{}{
			"relationship_id": id,
		})
		// Log error but don't fail the operation
	}

	// Log audit event
	s.logAuditEvent(ctx, "relationship", id.String(), "delete", userID.String(), map[string]interface{}{
		"source_id":         relationship.SourceID,
		"target_id":         relationship.TargetID,
		"relationship_type": relationship.RelationshipType,
	})

	return nil
}

// Helper methods

func (s *Service) validateCITypeSchema(req *CreateCITypeRequest) error {
	// Check for duplicate attribute names in required
	requiredNames := make(map[string]bool)
	for _, attr := range req.RequiredAttributes {
		if requiredNames[attr.Name] {
			return fmt.Errorf("duplicate required attribute name: %s", attr.Name)
		}
		requiredNames[attr.Name] = true
	}

	// Check for duplicate attribute names in optional
	optionalNames := make(map[string]bool)
	for _, attr := range req.OptionalAttributes {
		if optionalNames[attr.Name] {
			return fmt.Errorf("duplicate optional attribute name: %s", attr.Name)
		}
		if requiredNames[attr.Name] {
			return fmt.Errorf("attribute '%s' exists in both required and optional", attr.Name)
		}
		optionalNames[attr.Name] = true
	}

	return nil
}

func (s *Service) validateCITypeSchemaUpdate(req *UpdateCITypeRequest) error {
	// Similar validation as create, but for partial updates
	if req.RequiredAttributes != nil && req.OptionalAttributes != nil {
		requiredNames := make(map[string]bool)
		for _, attr := range req.RequiredAttributes {
			if requiredNames[attr.Name] {
				return fmt.Errorf("duplicate required attribute name: %s", attr.Name)
			}
			requiredNames[attr.Name] = true
		}

		for _, attr := range req.OptionalAttributes {
			if requiredNames[attr.Name] {
				return fmt.Errorf("attribute '%s' exists in both required and optional", attr.Name)
			}
		}
	}

	return nil
}

func (s *Service) getCIFromCache(ctx context.Context, id uuid.UUID) (*ConfigurationItem, error) {
	key := fmt.Sprintf("ci:%s", id)
	data, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var ci ConfigurationItem
	if err := json.Unmarshal([]byte(data), &ci); err != nil {
		return nil, err
	}

	return &ci, nil
}

func (s *Service) cacheCI(ctx context.Context, ci *ConfigurationItem) error {
	key := fmt.Sprintf("ci:%s", ci.ID)
	data, err := json.Marshal(ci)
	if err != nil {
		return err
	}

	return s.redis.Set(ctx, key, data, 5*time.Minute).Err()
}

func (s *Service) invalidateCICache(ctx context.Context, id uuid.UUID) {
	key := fmt.Sprintf("ci:%s", id)
	s.redis.Del(ctx, key)
}

func (s *Service) logAuditEvent(ctx context.Context, entityType, entityID, action, performedBy string, details map[string]interface{}) {
	// This would integrate with the audit service
	s.logger.InfoAudit(entityType, entityID, action, performedBy, details)
}

// Graph Operations - delegating to Neo4j service

func (s *Service) GetCIRelationships(ctx context.Context, id uuid.UUID) ([]RelationshipGraph, error) {
	return s.neo4j.GetCIRelationships(ctx, id)
}

func (s *Service) GetGraphData(ctx context.Context, filters GraphFilters) (*GraphData, error) {
	return s.neo4j.GetGraphData(ctx, filters)
}

func (s *Service) GetCINetwork(ctx context.Context, id uuid.UUID, depth int) (*CINetwork, error) {
	return s.neo4j.GetCINetwork(ctx, id, depth)
}

func (s *Service) GetImpactAnalysis(ctx context.Context, id uuid.UUID) (*ImpactAnalysis, error) {
	return s.neo4j.GetImpactAnalysis(ctx, id)
}

func (s *Service) GetCITypesByUsage(ctx context.Context) ([]CITypeUsage, error) {
	return s.neo4j.GetCITypesByUsage(ctx)
}

func (s *Service) FindCycles(ctx context.Context) ([][]uuid.UUID, error) {
	return s.neo4j.FindCycles(ctx)
}

func (s *Service) GetMostConnectedCIs(ctx context.Context, limit int) ([]CIConnectivity, error) {
	return s.neo4j.GetMostConnectedCIs(ctx, limit)
}

// Dashboard statistics

type DashboardStats struct {
	TotalCIs         int64 `json:"total_cis"`
	TotalCITypes     int64 `json:"total_ci_types"`
	TotalRelationships int64 `json:"total_relationships"`
	TotalUsers       int64 `json:"total_users"`
}

func (s *Service) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	// Get counts concurrently for better performance
	type result struct {
		name  string
		count int64
		err   error
	}

	results := make(chan result, 4)

	// Count CIs
	go func() {
		count, err := s.repo.CountCIs(ctx)
		results <- result{"cis", count, err}
	}()

	// Count CI types
	go func() {
		count, err := s.repo.CountCITypes(ctx)
		results <- result{"ci_types", count, err}
	}()

	// Count relationships
	go func() {
		count, err := s.repo.CountRelationships(ctx)
		results <- result{"relationships", count, err}
	}()

	// Count users
	go func() {
		count, err := s.repo.CountUsers(ctx)
		results <- result{"users", count, err}
	}()

	stats := &DashboardStats{}
	completed := 0

	for completed < 4 {
		select {
		case res := <-results:
			switch res.name {
			case "cis":
				if res.err != nil {
					return nil, fmt.Errorf("failed to get CI count: %w", res.err)
				}
				stats.TotalCIs = res.count
			case "ci_types":
				if res.err != nil {
					return nil, fmt.Errorf("failed to get CI type count: %w", res.err)
				}
				stats.TotalCITypes = res.count
			case "relationships":
				if res.err != nil {
					return nil, fmt.Errorf("failed to get relationship count: %w", res.err)
				}
				stats.TotalRelationships = res.count
			case "users":
				if res.err != nil {
					return nil, fmt.Errorf("failed to get user count: %w", res.err)
				}
				stats.TotalUsers = res.count
			}
			completed++
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return stats, nil
}

// Custom error types
type ServiceValidationError struct {
	Message string                 `json:"message"`
	Errors  []ValidationError   `json:"errors"`
}

func (e ServiceValidationError) Error() string {
	return e.Message
}