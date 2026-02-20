package ea

import (
	"context"

	"github.com/google/uuid"
	"github.com/pustaka/pustaka/internal/ci"
)

// GovernanceService handles Governance domain EA operations
type GovernanceService struct {
	*Service
}

// NewGovernanceService creates a new Governance domain service
func NewGovernanceService(baseService *Service) *GovernanceService {
	return &GovernanceService{Service: baseService}
}

// CreateGovernancePolicy creates a Governance Policy
func (s *GovernanceService) CreateGovernancePolicy(ctx context.Context, req *CreateEACIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	req.CIType = "EA.Governance-Policy"

	if req.Attributes == nil {
		req.Attributes = make(map[string]interface{})
	}

	return s.CreateEACI(ctx, req, userID)
}

// CreateComplianceRequirement creates a Compliance Requirement
func (s *GovernanceService) CreateComplianceRequirement(ctx context.Context, req *CreateEACIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	req.CIType = "EA.Governance-Compliance"

	if req.Attributes == nil {
		req.Attributes = make(map[string]interface{})
	}

	return s.CreateEACI(ctx, req, userID)
}

// ListGovernancePolicies retrieves all Governance Policies
func (s *GovernanceService) ListGovernancePolicies(ctx context.Context) ([]*ci.ConfigurationItem, error) {
	response, err := s.ciService.ListCIs(ctx, ci.ListCIFilters{CIType: "EA.Governance-Policy"}, 1, 1000)
	if err != nil {
		return nil, err
	}
	return response.CIs, nil
}

// ListComplianceRequirements retrieves all Compliance Requirements
func (s *GovernanceService) ListComplianceRequirements(ctx context.Context) ([]*ci.ConfigurationItem, error) {
	response, err := s.ciService.ListCIs(ctx, ci.ListCIFilters{CIType: "EA.Governance-Compliance"}, 1, 1000)
	if err != nil {
		return nil, err
	}
	return response.CIs, nil
}
