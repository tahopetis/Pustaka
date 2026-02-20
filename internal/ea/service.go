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
