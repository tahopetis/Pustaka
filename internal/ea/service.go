package ea

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
	"github.com/pustaka/pustaka/internal/ci"
)

// Service is the base EA service that extends CI service
type Service struct {
	ciService    *ci.Service
	ciRepo       *ci.Repository
	repo         *Repository
	neo4j        *ci.Neo4jService
	redis        *redis.Client
	auditService *ci.AuditService
	logger       *pustakaLogger.Logger
}

// NewService creates a new EA service
func NewService(
	ciService *ci.Service,
	ciRepo *ci.Repository,
	repo *Repository,
	neo4j *ci.Neo4jService,
	redis *redis.Client,
	auditService *ci.AuditService,
	logger *pustakaLogger.Logger,
) *Service {
	return &Service{
		ciService:    ciService,
		ciRepo:       ciRepo,
		repo:         repo,
		neo4j:        neo4j,
		redis:        redis,
		auditService: auditService,
		logger:       logger,
	}
}

// ============================================================================
// EA Entity CRUD (extends CI service)
// ============================================================================

// CreateEACI creates an EA entity with domain validation
func (s *Service) CreateEACI(ctx context.Context, req *CreateEACIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	// 1. Validate EA domain from CI type
	domain, err := ExtractEADomain(req.CIType)
	if err != nil {
		s.logger.Error().Err(err).Str("ci_type", req.CIType).Msg("Invalid EA CI type")
		return nil, fmt.Errorf("invalid EA CI type: %w", err)
	}

	// 2. Verify EA team exists
	team, err := s.repo.GetTeamByName(ctx, req.Owner)
	if err != nil {
		s.logger.Error().Err(err).Str("team_name", req.Owner).Msg("EA team not found")
		return nil, fmt.Errorf("EA team not found: %w", err)
	}

	// 3. Get CI type definition for validation
	ciType, err := s.ciRepo.GetCITypeByName(ctx, req.CIType)
	if err != nil {
		s.logger.Error().Err(err).Str("ci_type", req.CIType).Msg("CI type not found")
		return nil, fmt.Errorf("CI type not found: %w", err)
	}

	// 4. EA-specific validation (with override support)
	var validationErrors []string
	if !req.OverrideValidation {
		if err := s.validateEACIAttributes(ctx, domain, req.CIType, req.Attributes); err != nil {
			validationErrors = append(validationErrors, err.Error())
		}
	} else {
		// Admin override - log justification
		s.logger.Warn().
			Str("ea_domain", string(domain)).
			Str("ci_type", req.CIType).
			Str("override_justification", req.OverrideJustification).
			Str("user_id", userID.String()).
			Msg("EA validation overridden by admin")
	}

	// 5. Calculate data quality score
	validAttrCount := len(req.Attributes)
	totalAttrCount := len(ciType.RequiredAttributes) + len(ciType.OptionalAttributes)
	dataQualityScore := CalculateDataQualityScore(validAttrCount, totalAttrCount, validationErrors)

	// 6. Add EA metadata to attributes
	eaMetadata := EAMetadata{
		Source:           "manual",
		LastUpdatedBy:    userID.String(),
		DataQualityScore: dataQualityScore,
		ValidationErrors: validationErrors,
	}

	// Merge metadata into attributes (convert to map[string]interface{})
	if req.Attributes == nil {
		req.Attributes = make(map[string]interface{})
	}
	req.Attributes["metadata"] = eaMetadata
	req.Attributes["ea_domain"] = string(domain)
	req.Attributes["ea_team_id"] = team.ID.String()

	// 7. Add EA domain tag for filtering
	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}
	tags = append(tags, string(domain))

	// 8. Convert to CI request and create via CI service
	// Note: CreateCIRequest doesn't have Description field, store in attributes
	if req.Attributes == nil {
		req.Attributes = make(map[string]interface{})
	}
	if req.Description != "" {
		req.Attributes["description"] = req.Description
	}

	ciReq := &ci.CreateCIRequest{
		Name:              req.Name,
		CIType:            req.CIType,
		Attributes:        req.Attributes,
		Tags:              tags,
		LifecycleStatusID: &req.LifecycleStatusID,
	}

	ciEntity, err := s.ciService.CreateCI(ctx, ciReq, userID)
	if err != nil {
		s.logger.Error().Err(err).Str("ea_domain", string(domain)).Msg("Failed to create EA entity via CI service")
		return nil, fmt.Errorf("failed to create EA entity: %w", err)
	}

	// 9. Log EA-specific audit entry
	s.auditService.CreateAuditLog(ctx, "ea_entity", &ciEntity.ID, "create", userID, map[string]interface{}{
		"ea_domain": string(domain),
		"ci_name":   req.Name,
		"ci_type":   req.CIType,
		"team":      team.Name,
	}, "", "")

	s.logger.Info().
		Str("ea_domain", string(domain)).
		Str("ci_name", req.Name).
		Str("ci_id", ciEntity.ID.String()).
		Int("data_quality_score", dataQualityScore).
		Msg("EA entity created successfully")

	return ciEntity, nil
}

// validateEACIAttributes performs EA-specific validation
func (s *Service) validateEACIAttributes(ctx context.Context, domain EADomain, ciType string, attributes map[string]interface{}) error {
	// Domain-specific validation
	switch domain {
	case EADomainBusiness:
		return ValidateBusinessAttributes(ciType, attributes)
	case EADomainApplication:
		return ValidateApplicationAttributes(ciType, attributes)
	case EADomainData:
		return ValidateDataAttributes(ciType, attributes)
	case EADomainTechnology:
		return ValidateTechnologyAttributes(ciType, attributes)
	case EADomainInfrastructure:
		return ValidateInfrastructureAttributes(ciType, attributes)
	case EADomainSecurity:
		return ValidateSecurityAttributes(ciType, attributes)
	case EADomainGovernance:
		return ValidateGovernanceAttributes(ciType, attributes)
	case EADomainStrategy:
		return ValidateStrategyAttributes(ciType, attributes)
	default:
		return fmt.Errorf("unknown EA domain: %s", domain)
	}
}

// GetEACI retrieves an EA entity (wraps CI service)
func (s *Service) GetEACI(ctx context.Context, id uuid.UUID) (*ci.ConfigurationItem, error) {
	ciEntity, err := s.ciService.GetCI(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify it's an EA entity
	if _, err := ExtractEADomain(ciEntity.CIType); err != nil {
		return nil, fmt.Errorf("not an EA entity: %w", err)
	}

	return ciEntity, nil
}

// UpdateEACI updates an EA entity with EA validation
func (s *Service) UpdateEACI(ctx context.Context, id uuid.UUID, req *UpdateEACIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	// 1. Get existing EA entity
	existing, err := s.GetEACI(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Extract domain for validation
	domain, err := ExtractEADomain(existing.CIType)
	if err != nil {
		return nil, err
	}

	// 3. EA-specific validation (with override support)
	var validationErrors []string
	if req.Attributes != nil && !req.OverrideValidation {
		if err := s.validateEACIAttributes(ctx, domain, existing.CIType, req.Attributes); err != nil {
			validationErrors = append(validationErrors, err.Error())
		}
	}

	// 4. Update metadata and handle name/description in attributes
	if req.Attributes == nil {
		req.Attributes = make(map[string]interface{})
	}
	eaMetadata := EAMetadata{
		Source:           "manual",
		LastUpdatedBy:    userID.String(),
		DataQualityScore: 0, // Will recalculate
		ValidationErrors: validationErrors,
	}
	req.Attributes["metadata"] = eaMetadata

	// Note: UpdateCIRequest doesn't have Name or Description fields
	// Store them in attributes if provided
	if req.Name != "" {
		req.Attributes["name"] = req.Name
	}
	if req.Description != "" {
		req.Attributes["description"] = req.Description
	}

	// 5. Convert to CI request and update via CI service
	ciReq := &ci.UpdateCIRequest{
		Attributes:        req.Attributes,
		Tags:              req.Tags,
		LifecycleStatusID: req.LifecycleStatusID,
	}

	ciEntity, err := s.ciService.UpdateCI(ctx, id, ciReq, userID)
	if err != nil {
		return nil, err
	}

	// 6. Log EA-specific audit entry
	s.auditService.CreateAuditLog(ctx, "ea_entity", &id, "update", userID, map[string]interface{}{
		"ea_domain": string(domain),
		"ci_name":   ciEntity.Name,
	}, "", "")

	s.logger.Info().
		Str("ea_domain", string(domain)).
		Str("ci_id", id.String()).
		Msg("EA entity updated successfully")

	return ciEntity, nil
}

// DeleteEACI deletes an EA entity (wraps CI service with EA audit log)
func (s *Service) DeleteEACI(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// 1. Get existing EA entity for audit
	existing, err := s.GetEACI(ctx, id)
	if err != nil {
		return err
	}

	domain, _ := ExtractEADomain(existing.CIType)

	// 2. Delete via CI service
	if err := s.ciService.DeleteCI(ctx, id, userID); err != nil {
		return err
	}

	// 3. Log EA-specific audit entry
	s.auditService.CreateAuditLog(ctx, "ea_entity", &id, "delete", userID, map[string]interface{}{
		"ea_domain":  string(domain),
		"ci_name":    existing.Name,
		"ci_type":    existing.CIType,
	}, "", "")

	s.logger.Info().
		Str("ea_domain", string(domain)).
		Str("ci_id", id.String()).
		Msg("EA entity deleted successfully")

	return nil
}

// ============================================================================
// EA Relationship Management (with cross-domain validation)
// ============================================================================

// CreateEARelationship creates an EA relationship with cross-domain validation
func (s *Service) CreateEARelationship(ctx context.Context, req *CreateEARelationshipRequest, userID uuid.UUID) (*ci.Relationship, error) {
	// 1. Get source and target CIs
	source, err := s.ciService.GetCI(ctx, req.SourceID)
	if err != nil {
		return nil, fmt.Errorf("source CI not found: %w", err)
	}

	target, err := s.ciService.GetCI(ctx, req.TargetID)
	if err != nil {
		return nil, fmt.Errorf("target CI not found: %w", err)
	}

	// 2. Cross-domain validation
	if err := ValidateCrossDomainRelationship(source.CIType, target.CIType, req.RelationshipType); err != nil {
		s.logger.Warn().
			Str("source_type", source.CIType).
			Str("target_type", target.CIType).
			Str("relationship_type", req.RelationshipType).
			Err(err).
			Msg("Cross-domain relationship validation failed")
		return nil, fmt.Errorf("cross-domain validation failed: %w", err)
	}

	// 3. Convert to CI relationship request
	ciReq := &ci.CreateRelationshipRequest{
		SourceID:         req.SourceID,
		TargetID:         req.TargetID,
		RelationshipType: req.RelationshipType,
		Attributes:       req.Attributes,
	}

	// 4. Create relationship via CI service
	rel, err := s.ciService.CreateRelationship(ctx, ciReq, userID)
	if err != nil {
		return nil, err
	}

	// 5. Log EA-specific audit entry
	s.auditService.CreateAuditLog(ctx, "ea_relationship", &rel.ID, "create", userID, map[string]interface{}{
		"source_name":       source.Name,
		"target_name":       target.Name,
		"relationship_type": req.RelationshipType,
	}, "", "")

	s.logger.Info().
		Str("source_type", source.CIType).
		Str("target_type", target.CIType).
		Str("relationship_type", req.RelationshipType).
		Msg("EA relationship created successfully")

	return rel, nil
}

// ============================================================================
// EA Teams Management
// ============================================================================

// CreateTeam creates a new EA team
func (s *Service) CreateTeam(ctx context.Context, req *CreateEATeamRequest, userID uuid.UUID) (*EATeam, error) {
	team, err := s.repo.CreateTeam(ctx, req, userID)
	if err != nil {
		return nil, err
	}

	s.logger.Info().
		Str("team_name", team.Name).
		Str("team_id", team.ID.String()).
		Msg("EA team created successfully")

	return team, nil
}

// GetTeam retrieves an EA team by name
func (s *Service) GetTeam(ctx context.Context, name string) (*EATeam, error) {
	return s.repo.GetTeamByName(ctx, name)
}

// ListTeams retrieves all EA teams
func (s *Service) ListTeams(ctx context.Context) ([]*EATeam, error) {
	return s.repo.ListTeams(ctx)
}

// UpdateTeam updates an EA team
func (s *Service) UpdateTeam(ctx context.Context, id uuid.UUID, req *UpdateEATeamRequest, userID uuid.UUID) (*EATeam, error) {
	team, err := s.repo.UpdateTeam(ctx, id, req)
	if err != nil {
		return nil, err
	}

	s.logger.Info().
		Str("team_id", id.String()).
		Msg("EA team updated successfully")

	return team, nil
}

// DeleteTeam deletes an EA team
func (s *Service) DeleteTeam(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	if err := s.repo.DeleteTeam(ctx, id); err != nil {
		return err
	}

	s.logger.Info().
		Str("team_id", id.String()).
		Msg("EA team deleted successfully")

	return nil
}

// ============================================================================
// EA Entity CRUD with New Repository (direct database access)
// ============================================================================

// CreateEntity creates a new EA entity using the EA repository
func (s *Service) CreateEntity(ctx context.Context, req *CreateEACIRequest, userID uuid.UUID) (*EAEntity, error) {
	// 1. Validate EA domain from CI type
	domain, err := ExtractEADomain(req.CIType)
	if err != nil {
		s.logger.Error().Err(err).Str("ci_type", req.CIType).Msg("Invalid EA CI type")
		return nil, fmt.Errorf("invalid EA CI type: %w", err)
	}

	// 2. Verify EA team exists
	team, err := s.repo.GetTeamByName(ctx, req.Owner)
	if err != nil {
		s.logger.Error().Err(err).Str("team_name", req.Owner).Msg("EA team not found")
		return nil, fmt.Errorf("EA team not found: %w", err)
	}

	// 3. Get CI type definition for validation
	ciType, err := s.repo.GetCITypeByName(ctx, req.CIType)
	if err != nil {
		s.logger.Error().Err(err).Str("ci_type", req.CIType).Msg("CI type not found")
		return nil, fmt.Errorf("CI type not found: %w", err)
	}

	// 4. Prepare attributes with EA metadata
	attributes := req.Attributes
	if attributes == nil {
		attributes = make(map[string]interface{})
	}

	// Add EA-specific metadata
	attributes["ea_domain"] = string(domain)
	attributes["ea_owner"] = req.Owner
	attributes["ea_team_id"] = team.ID.String()

	if req.Description != "" {
		attributes["description"] = req.Description
	}

	// 5. Validate entity attributes
	validationResult, err := ValidateEntityAttributes(&EAEntity{
		Name:       req.Name,
		CIType:     req.CIType,
		Attributes: attributes,
		Tags:       req.Tags,
	}, ciType)
	if err != nil {
		s.logger.Error().Err(err).Msg("Validation error")
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// 6. Validate cross-field rules
	crossFieldErrors := ValidateCrossFieldRules(&EAEntity{
		Name:       req.Name,
		CIType:     req.CIType,
		Attributes: attributes,
		Tags:       req.Tags,
	}, ciType)

	// Merge validation errors
	for _, e := range crossFieldErrors {
		validationResult.Errors = append(validationResult.Errors, e)
	}
	if len(crossFieldErrors) > 0 {
		validationResult.IsValid = false
	}

	// 7. Handle admin override
	if !req.OverrideValidation && !validationResult.IsValid {
		// Return validation errors
		s.logger.Warn().
			Str("ea_domain", string(domain)).
			Int("error_count", len(validationResult.Errors)).
			Msg("EA entity validation failed")
		return nil, ErrValidationFailed
	}

	// 8. Add EA domain tag
	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}
	tags = append(tags, string(domain))

	// 9. Create entity
	entity := &EAEntity{
		Name:              req.Name,
		CIType:            req.CIType,
		Attributes:        attributes,
		Tags:              tags,
		LifecycleStatusID: &req.LifecycleStatusID,
		DataQualityScore:  validationResult.DataQualityScore,
		CreatedBy:         userID,
	}

	result, err := s.repo.Create(ctx, entity)
	if err != nil {
		s.logger.Error().Err(err).Str("ea_domain", string(domain)).Msg("Failed to create EA entity")
		return nil, fmt.Errorf("failed to create EA entity: %w", err)
	}

	// 10. Log audit event
	s.auditService.CreateAuditLog(ctx, "ea", &result.ID, "create", userID, map[string]interface{}{
		"ea_domain":         string(domain),
		"ci_name":           req.Name,
		"ci_type":           req.CIType,
		"team":              team.Name,
		"data_quality_score": validationResult.DataQualityScore,
		"validation_errors": len(validationResult.Errors),
	}, "", "")

	s.logger.Info().
		Str("ea_domain", string(domain)).
		Str("ci_name", req.Name).
		Str("ci_id", result.ID.String()).
		Float64("data_quality_score", validationResult.DataQualityScore).
		Msg("EA entity created successfully")

	return result, nil
}

// GetEntity retrieves an EA entity by ID
func (s *Service) GetEntity(ctx context.Context, id string) (*EAEntity, error) {
	entity, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

// UpdateEntity updates an EA entity
func (s *Service) UpdateEntity(ctx context.Context, id string, req *UpdateEACIRequest, userID uuid.UUID) (*EAEntity, error) {
	// 1. Get existing entity
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Get CI type for validation
	ciType, err := s.repo.GetCITypeByName(ctx, existing.CIType)
	if err != nil {
		return nil, fmt.Errorf("CI type not found: %w", err)
	}

	// 3. Validate lifecycle status transition if status is being changed
	if req.LifecycleStatusID != nil && existing.LifecycleStatusID != nil {
		// Get current and new lifecycle status names
		currentStatus, err := s.repo.GetLifecycleStatusByID(ctx, *existing.LifecycleStatusID)
		if err != nil {
			return nil, fmt.Errorf("failed to get current lifecycle status: %w", err)
		}

		newStatus, err := s.repo.GetLifecycleStatusByID(ctx, *req.LifecycleStatusID)
		if err != nil {
			return nil, fmt.Errorf("failed to get new lifecycle status: %w", err)
		}

		// Validate transition if status is actually changing
		if currentStatus.ID != newStatus.ID {
			if err := ValidateLifecycleTransition(currentStatus.Name, newStatus.Name); err != nil {
				return nil, err
			}
		}
	}

	// 4. Prepare updated attributes
	attributes := existing.Attributes
	if req.Attributes != nil {
		attributes = req.Attributes
	}

	// Update EA metadata
	attributes["ea_last_updated_by"] = userID.String()

	if req.Description != "" {
		attributes["description"] = req.Description
	}

	if req.Owner != "" {
		// Verify team exists
		team, err := s.repo.GetTeamByName(ctx, req.Owner)
		if err != nil {
			return nil, fmt.Errorf("EA team not found: %w", err)
		}
		attributes["ea_owner"] = req.Owner
		attributes["ea_team_id"] = team.ID.String()
	}

	// 5. Validate updated attributes
	updatedEntity := &EAEntity{
		ID:                existing.ID,
		Name:              req.Name,
		CIType:            existing.CIType,
		Attributes:        attributes,
		Tags:              req.Tags,
		LifecycleStatusID: req.LifecycleStatusID,
	}

	validationResult, err := ValidateEntityAttributes(updatedEntity, ciType)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	crossFieldErrors := ValidateCrossFieldRules(updatedEntity, ciType)
	for _, e := range crossFieldErrors {
		validationResult.Errors = append(validationResult.Errors, e)
	}

	// 6. Handle admin override
	if !req.OverrideValidation && !validationResult.IsValid {
		return nil, ErrValidationFailed
	}

	// 7. Update entity
	entity := &EAEntity{
		ID:                existing.ID,
		Name:              req.Name,
		CIType:            existing.CIType,
		Attributes:        attributes,
		Tags:              req.Tags,
		LifecycleStatusID: req.LifecycleStatusID,
		DataQualityScore:  validationResult.DataQualityScore,
		UpdatedBy:         &userID,
	}

	if err := s.repo.Update(ctx, entity); err != nil {
		return nil, fmt.Errorf("failed to update EA entity: %w", err)
	}

	// 8. Get updated entity
	result, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 9. Log audit event
	domain, _ := ExtractEADomain(existing.CIType)

	// Include lifecycle transition details in audit log
	details := map[string]interface{}{
		"ea_domain":         string(domain),
		"ci_name":           result.Name,
		"data_quality_score": validationResult.DataQualityScore,
	}

	if req.LifecycleStatusID != nil && existing.LifecycleStatusID != nil {
		currentStatus, _ := s.repo.GetLifecycleStatusByID(ctx, *existing.LifecycleStatusID)
		newStatus, _ := s.repo.GetLifecycleStatusByID(ctx, *req.LifecycleStatusID)
		if currentStatus.ID != newStatus.ID {
			details["lifecycle_transition"] = fmt.Sprintf("%s → %s", currentStatus.Name, newStatus.Name)
		}
	}

	s.auditService.CreateAuditLog(ctx, "ea", &result.ID, "update", userID, details, "", "")

	s.logger.Info().
		Str("ea_domain", string(domain)).
		Str("ci_id", id).
		Msg("EA entity updated successfully")

	return result, nil
}

// DeleteEntity deletes an EA entity
func (s *Service) DeleteEntity(ctx context.Context, id string, userID uuid.UUID, forceDelete bool) error {
	// 1. Get existing entity for audit
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	domain, _ := ExtractEADomain(existing.CIType)

	// 2. Check relationships (done in repo.Delete)
	if err := s.repo.Delete(ctx, id, forceDelete); err != nil {
		return err
	}

	// 3. Log audit event
	s.auditService.CreateAuditLog(ctx, "ea", &existing.ID, "delete", userID, map[string]interface{}{
		"ea_domain": string(domain),
		"ci_name":   existing.Name,
		"ci_type":   existing.CIType,
	}, "", "")

	s.logger.Info().
		Str("ea_domain", string(domain)).
		Str("ci_id", id).
		Msg("EA entity deleted successfully")

	return nil
}

// ListEntities retrieves EA entities with filtering and pagination
func (s *Service) ListEntities(ctx context.Context, filter EAFilter) ([]EAEntity, int, error) {
	entities, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

// ValidateEntity validates an EA entity and returns validation result
func (s *Service) ValidateEntity(ctx context.Context, id string) (*ValidationResult, error) {
	// 1. Get entity
	entity, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Get CI type
	ciType, err := s.repo.GetCITypeByName(ctx, entity.CIType)
	if err != nil {
		return nil, fmt.Errorf("CI type not found: %w", err)
	}

	// 3. Validate attributes
	result, err := ValidateEntityAttributes(entity, ciType)
	if err != nil {
		return nil, err
	}

	// 4. Validate cross-field rules
	crossFieldErrors := ValidateCrossFieldRules(entity, ciType)
	for _, e := range crossFieldErrors {
		result.Errors = append(result.Errors, e)
	}
	if len(crossFieldErrors) > 0 {
		result.IsValid = false
	}

	return result, nil
}

// GetEntityAuditLogs retrieves audit logs for an EA entity with user information
func (s *Service) GetEntityAuditLogs(ctx context.Context, entityID uuid.UUID, page, pageSize int) ([]map[string]interface{}, int64, error) {
	// Get audit logs from audit service
	filters := ci.AuditLogFilters{
		EntityType: "ea",
		EntityID:   &entityID,
		Page:       page,
		Limit:      pageSize,
		Sort:       "timestamp",
		Order:      "DESC",
	}

	response, err := s.auditService.ListAuditLogs(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	// Enrich audit logs with user information
	auditLogsWithUsers := make([]map[string]interface{}, len(response.AuditLogs))
	for i, log := range response.AuditLogs {
		auditLogsWithUsers[i] = map[string]interface{}{
			"id":          log.ID,
			"timestamp":   log.Timestamp,
			"action":      log.Action,
			"user_id":     log.PerformedBy,
			"details":     log.Details,
			"ip_address":  log.IPAddress,
			"user_agent":  log.UserAgent,
		}

		// Try to get user name (this will be filled by the handler if user service is available)
		// For now, we'll just return the user_id
	}

	return auditLogsWithUsers, response.Pagination.Total, nil
}

