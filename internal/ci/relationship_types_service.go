package ci

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

// RelationshipTypeService handles business logic for relationship types
type RelationshipTypeService struct {
	repo         *RelationshipTypeRepository
	ciRepo       *Repository
	neo4jService *Neo4jService
	redis        *redis.Client
	auditService *AuditService
	logger       *pustakaLogger.Logger
	db           *pgxpool.Pool
}

func NewRelationshipTypeService(
	repo *RelationshipTypeRepository,
	ciRepo *Repository,
	neo4jService *Neo4jService,
	redis *redis.Client,
	auditService *AuditService,
	logger *pustakaLogger.Logger,
	db *pgxpool.Pool,
) *RelationshipTypeService {
	return &RelationshipTypeService{
		repo:         repo,
		ciRepo:       ciRepo,
		neo4jService: neo4jService,
		redis:        redis,
		auditService: auditService,
		logger:       logger,
		db:           db,
	}
}

// CreateRelationshipType creates a new relationship type
func (s *RelationshipTypeService) CreateRelationshipType(ctx context.Context, req *CreateRelationshipTypeRequest, userID uuid.UUID) (*RelationshipType, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check for duplicate name
	existing, err := s.repo.GetByName(ctx, req.Name)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("relationship type '%s' already exists", req.Name)
	}

	// Validate CI types if specified
	if err := s.validateCITypes(ctx, req.AllowedSourceTypes, req.AllowedTargetTypes); err != nil {
		return nil, fmt.Errorf("CI type validation failed: %w", err)
	}

	// Create relationship type
	relationshipType := NewRelationshipTypeFromRequest(req, userID)

	result, err := s.repo.Create(ctx, relationshipType)
	if err != nil {
		return nil, fmt.Errorf("failed to create relationship type: %w", err)
	}

	// Invalidate cache
	s.invalidateRelationshipTypeCache(ctx)

	// Log audit event
	s.logAuditEvent(ctx, "relationship_type", result.ID.String(), "create", userID.String(), map[string]interface{}{
		"name":         result.Name,
		"display_name": result.DisplayName,
		"category":     result.Category,
	})

	s.logger.InfoService("relationship_type", "create_relationship_type", map[string]interface{}{
		"relationship_type_id": result.ID,
		"name":                 result.Name,
		"user_id":              userID,
	})

	return result, nil
}

// GetRelationshipType retrieves a relationship type by ID
func (s *RelationshipTypeService) GetRelationshipType(ctx context.Context, id uuid.UUID) (*RelationshipType, error) {
	// Try cache first
	if rt, err := s.getRelationshipTypeFromCache(ctx, id); err == nil {
		return rt, nil
	}

	// Get from database
	rt, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get relationship type: %w", err)
	}

	// Cache the result
	s.cacheRelationshipType(ctx, rt)

	return rt, nil
}

// GetRelationshipTypeByName retrieves a relationship type by name
func (s *RelationshipTypeService) GetRelationshipTypeByName(ctx context.Context, name string) (*RelationshipType, error) {
	return s.repo.GetByName(ctx, name)
}

// ListRelationshipTypes retrieves paginated relationship types with filters
func (s *RelationshipTypeService) ListRelationshipTypes(ctx context.Context, filters ListRelationshipTypeFilters, page, limit int) (*RelationshipTypeListResponse, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	return s.repo.List(ctx, filters, page, limit)
}

// GetActiveRelationshipTypes returns all active relationship types
func (s *RelationshipTypeService) GetActiveRelationshipTypes(ctx context.Context) ([]RelationshipType, error) {
	// Try cache first
	if rts, err := s.getActiveRelationshipTypesFromCache(ctx); err == nil {
		return rts, nil
	}

	// Get from database
	rts, err := s.repo.GetActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active relationship types: %w", err)
	}

	// Cache the result
	s.cacheActiveRelationshipTypes(ctx, rts)

	return rts, nil
}

// UpdateRelationshipType updates an existing relationship type
func (s *RelationshipTypeService) UpdateRelationshipType(ctx context.Context, id uuid.UUID, req *UpdateRelationshipTypeRequest, userID uuid.UUID) (*RelationshipType, error) {
	// Get current relationship type
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get relationship type: %w", err)
	}

	// Check if it's a system type
	if !current.CanBeModified() {
		return nil, fmt.Errorf("system relationship types cannot be modified")
	}

	// Validate update request
	if err := s.validateUpdateRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Validate CI types if being updated
	if req.AllowedSourceTypes != nil || req.AllowedTargetTypes != nil {
		sourceTypes := current.AllowedSourceTypes
		targetTypes := current.AllowedTargetTypes

		if req.AllowedSourceTypes != nil {
			sourceTypes = req.AllowedSourceTypes
		}
		if req.AllowedTargetTypes != nil {
			targetTypes = req.AllowedTargetTypes
		}

		if err := s.validateCITypes(ctx, sourceTypes, targetTypes); err != nil {
			return nil, fmt.Errorf("CI type validation failed: %w", err)
		}
	}

	// Update relationship type
	current.UpdateFromRequest(req, userID)

	result, err := s.repo.Update(ctx, current)
	if err != nil {
		return nil, fmt.Errorf("failed to update relationship type: %w", err)
	}

	// Invalidate cache
	s.invalidateRelationshipTypeCache(ctx)

	// Log audit event
	s.logAuditEvent(ctx, "relationship_type", id.String(), "update", userID.String(), map[string]interface{}{
		"name":    result.Name,
		"changes": req,
	})

	s.logger.InfoService("relationship_type", "update_relationship_type", map[string]interface{}{
		"relationship_type_id": id,
		"name":                 result.Name,
		"user_id":              userID,
	})

	return result, nil
}

// DeleteRelationshipType deletes a relationship type
func (s *RelationshipTypeService) DeleteRelationshipType(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// Get relationship type for audit
	rt, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get relationship type: %w", err)
	}

	// Check if it's a system type
	if !rt.CanBeDeleted() {
		return fmt.Errorf("system relationship types cannot be deleted")
	}

	// Check if relationship type is in use
	usageCount, err := s.repo.GetUsageCount(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check usage: %w", err)
	}
	if usageCount > 0 {
		return fmt.Errorf("cannot delete relationship type that is in use (%d relationships)", usageCount)
	}

	// Delete from database
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete relationship type: %w", err)
	}

	// Invalidate cache
	s.invalidateRelationshipTypeCache(ctx)

	// Log audit event
	s.logAuditEvent(ctx, "relationship_type", id.String(), "delete", userID.String(), map[string]interface{}{
		"name": rt.Name,
	})

	s.logger.InfoService("relationship_type", "delete_relationship_type", map[string]interface{}{
		"relationship_type_id": id,
		"name":                 rt.Name,
		"user_id":              userID,
	})

	return nil
}

// ValidateRelationship validates if a relationship can be created between two CIs
func (s *RelationshipTypeService) ValidateRelationship(ctx context.Context, req *RelationshipCompatibilityRequest) (*RelationshipCompatibilityResponse, error) {
	// Get relationship type
	rt, err := s.repo.GetByName(ctx, req.RelationshipType)
	if err != nil {
		return &RelationshipCompatibilityResponse{
			IsValid: false,
			Message: "Relationship type not found",
			Reason:  fmt.Sprintf("Relationship type '%s' does not exist", req.RelationshipType),
		}, nil
	}

	// Parse CI IDs
	sourceID, err := uuid.Parse(req.SourceType)
	if err != nil {
		return &RelationshipCompatibilityResponse{
			IsValid: false,
			Message: "Invalid source CI ID",
			Reason:  "Source CI ID is not a valid UUID",
		}, nil
	}

	targetID, err := uuid.Parse(req.TargetType)
	if err != nil {
		return &RelationshipCompatibilityResponse{
			IsValid: false,
			Message: "Invalid target CI ID",
			Reason:  "Target CI ID is not a valid UUID",
		}, nil
	}

	// Get CIs
	sourceCI, err := s.ciRepo.GetCI(ctx, sourceID)
	if err != nil {
		return &RelationshipCompatibilityResponse{
			IsValid: false,
			Message: "Source CI not found",
			Reason:  fmt.Sprintf("Source CI with ID %s does not exist", sourceID),
		}, nil
	}

	targetCI, err := s.ciRepo.GetCI(ctx, targetID)
	if err != nil {
		return &RelationshipCompatibilityResponse{
			IsValid: false,
			Message: "Target CI not found",
			Reason:  fmt.Sprintf("Target CI with ID %s does not exist", targetID),
		}, nil
	}

	// Validate relationship using the model's validation method
	validation := rt.ValidateRelationship(sourceCI, targetCI)

	if !validation.IsValid {
		return &RelationshipCompatibilityResponse{
			IsValid: false,
			Message: "Relationship validation failed",
			Reason:  strings.Join(validation.Errors, "; "),
		}, nil
	}

	// Additional validation checks for specific relationship types
	if rt.Name == "depends_on" || rt.Name == "contains" {
		// Check for potential circular dependencies using Neo4j
		if err := s.neo4jService.CheckCircularDependency(ctx, sourceID, targetID, rt.Name); err != nil {
			return &RelationshipCompatibilityResponse{
				IsValid: true, // Still allow but warn
				Message: "Warning: Potential circular dependency",
				Reason:  err.Error(),
			}, nil
		}
	}

	return &RelationshipCompatibilityResponse{
		IsValid: true,
		Message: "Relationship is compatible",
	}, nil
}

// GetRelationshipTypeUsage returns usage statistics for relationship types
func (s *RelationshipTypeService) GetRelationshipTypeUsage(ctx context.Context) ([]RelationshipTypeUsage, error) {
	return s.repo.GetUsage(ctx)
}

// GetRelationshipTypeStatistics returns comprehensive statistics
func (s *RelationshipTypeService) GetRelationshipTypeStatistics(ctx context.Context) (*RelationshipTypeStatistics, error) {
	return s.repo.GetStatistics(ctx)
}

// GetRelationshipTypeCategories returns all unique category names
func (s *RelationshipTypeService) GetRelationshipTypeCategories(ctx context.Context) ([]string, error) {
	stats, err := s.repo.GetStatistics(ctx)
	if err != nil {
		return nil, err
	}

	categories := make([]string, 0, len(stats.Categories))
	for category := range stats.Categories {
		categories = append(categories, category)
	}

	return categories, nil
}

// Helper methods

func (s *RelationshipTypeService) validateCreateRequest(req *CreateRelationshipTypeRequest) error {
	// Name validation
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if len(req.Name) < 2 || len(req.Name) > 100 {
		return fmt.Errorf("name must be between 2 and 100 characters")
	}
	if !isValidRelationshipTypeName(req.Name) {
		return fmt.Errorf("name must contain only lowercase letters, numbers, and underscores")
	}

	// Forward label validation
	if strings.TrimSpace(req.ForwardLabel) == "" {
		return fmt.Errorf("forward label is required")
	}
	if len(req.ForwardLabel) < 1 || len(req.ForwardLabel) > 100 {
		return fmt.Errorf("forward label must be between 1 and 100 characters")
	}

	// Reverse label validation
	if strings.TrimSpace(req.ReverseLabel) == "" {
		return fmt.Errorf("reverse label is required")
	}
	if len(req.ReverseLabel) < 1 || len(req.ReverseLabel) > 100 {
		return fmt.Errorf("reverse label must be between 1 and 100 characters")
	}

	// Cardinality validation
	validCardinalities := map[string]bool{
		"one": true, "many": true,
	}
	if !validCardinalities[req.CardinalitySource] {
		return fmt.Errorf("invalid source cardinality: %s", req.CardinalitySource)
	}
	if !validCardinalities[req.CardinalityTarget] {
		return fmt.Errorf("invalid target cardinality: %s", req.CardinalityTarget)
	}

	return nil
}

func (s *RelationshipTypeService) validateUpdateRequest(req *UpdateRelationshipTypeRequest) error {
	// Display name validation
	if req.DisplayName != nil {
		if strings.TrimSpace(*req.DisplayName) == "" {
			return fmt.Errorf("display name cannot be empty")
		}
		if len(*req.DisplayName) < 2 || len(*req.DisplayName) > 100 {
			return fmt.Errorf("display name must be between 2 and 100 characters")
		}
	}

	// Labels validation
	if req.ForwardLabel != nil {
		if strings.TrimSpace(*req.ForwardLabel) == "" {
			return fmt.Errorf("forward label cannot be empty")
		}
		if len(*req.ForwardLabel) < 1 || len(*req.ForwardLabel) > 100 {
			return fmt.Errorf("forward label must be between 1 and 100 characters")
		}
	}

	if req.ReverseLabel != nil {
		if strings.TrimSpace(*req.ReverseLabel) == "" {
			return fmt.Errorf("reverse label cannot be empty")
		}
		if len(*req.ReverseLabel) < 1 || len(*req.ReverseLabel) > 100 {
			return fmt.Errorf("reverse label must be between 1 and 100 characters")
		}
	}

	// Cardinality validation
	if req.CardinalitySource != nil {
		validCardinalities := map[string]bool{
			"one": true, "many": true,
		}
		if !validCardinalities[*req.CardinalitySource] {
			return fmt.Errorf("invalid source cardinality: %s", *req.CardinalitySource)
		}
	}

	if req.CardinalityTarget != nil {
		validCardinalities := map[string]bool{
			"one": true, "many": true,
		}
		if !validCardinalities[*req.CardinalityTarget] {
			return fmt.Errorf("invalid target cardinality: %s", *req.CardinalityTarget)
		}
	}

	return nil
}

func (s *RelationshipTypeService) validateCITypes(ctx context.Context, sourceTypes, targetTypes []string) error {
	// Validate source types
	for _, ciType := range sourceTypes {
		_, err := s.ciRepo.GetCITypeByName(ctx, ciType)
		if err != nil {
			return fmt.Errorf("invalid source CI type: %s", ciType)
		}
	}

	// Validate target types
	for _, ciType := range targetTypes {
		_, err := s.ciRepo.GetCITypeByName(ctx, ciType)
		if err != nil {
			return fmt.Errorf("invalid target CI type: %s", ciType)
		}
	}

	return nil
}

// Cache methods

func (s *RelationshipTypeService) getRelationshipTypeFromCache(ctx context.Context, id uuid.UUID) (*RelationshipType, error) {
	key := fmt.Sprintf("relationship_type:%s", id)
	data, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var rt RelationshipType
	if err := json.Unmarshal([]byte(data), &rt); err != nil {
		return nil, err
	}

	return &rt, nil
}

func (s *RelationshipTypeService) cacheRelationshipType(ctx context.Context, rt *RelationshipType) error {
	key := fmt.Sprintf("relationship_type:%s", rt.ID)
	data, err := json.Marshal(rt)
	if err != nil {
		return err
	}

	return s.redis.Set(ctx, key, data, 10*time.Minute).Err()
}

func (s *RelationshipTypeService) getActiveRelationshipTypesFromCache(ctx context.Context) ([]RelationshipType, error) {
	key := "relationship_types:active"
	data, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var rts []RelationshipType
	if err := json.Unmarshal([]byte(data), &rts); err != nil {
		return nil, err
	}

	return rts, nil
}

func (s *RelationshipTypeService) cacheActiveRelationshipTypes(ctx context.Context, rts []RelationshipType) error {
	key := "relationship_types:active"
	data, err := json.Marshal(rts)
	if err != nil {
		return err
	}

	return s.redis.Set(ctx, key, data, 5*time.Minute).Err()
}

func (s *RelationshipTypeService) invalidateRelationshipTypeCache(ctx context.Context) {
	// Invalidate specific relationship type cache keys
	keys := []string{
		"relationship_types:active",
	}

	for _, key := range keys {
		s.redis.Del(ctx, key)
	}
}

func (s *RelationshipTypeService) logAuditEvent(ctx context.Context, entityType, entityID, action, performedBy string, details map[string]interface{}) {
	// Parse UUIDs
	var entityUUID *uuid.UUID
	if entityID != "" {
		if parsedID, err := uuid.Parse(entityID); err == nil {
			entityUUID = &parsedID
		}
	}

	performedByUUID, err := uuid.Parse(performedBy)
	if err != nil {
		s.logger.Error().Err(err).Str("performed_by", performedBy).Msg("Invalid user ID for audit logging")
		return
	}

	// Save to database via audit service
	if err := s.auditService.CreateAuditLog(ctx, entityType, entityUUID, action, performedByUUID, details, "", ""); err != nil {
		s.logger.Error().Err(err).Str("entity_type", entityType).Str("action", action).Msg("Failed to create audit log")
	}

	// Also log to structured logs for debugging
	s.logger.InfoAudit(entityType, entityID, action, performedBy, details)
}

// Validation helper function
func isValidRelationshipTypeName(name string) bool {
	if len(name) == 0 {
		return false
	}

	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}

	return true
}