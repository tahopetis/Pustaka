package ea

import (
	"context"

	"github.com/google/uuid"
	"github.com/pustaka/pustaka/internal/ci"
)

// SecurityService handles Security domain EA operations
type SecurityService struct {
	*Service
}

// NewSecurityService creates a new Security domain service
func NewSecurityService(baseService *Service) *SecurityService {
	return &SecurityService{Service: baseService}
}

// CreateSecurityControl creates a Security Control
func (s *SecurityService) CreateSecurityControl(ctx context.Context, req *CreateEACIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	req.CIType = "EA.Security-Control"

	if req.Attributes == nil {
		req.Attributes = make(map[string]interface{})
	}

	return s.CreateEACI(ctx, req, userID)
}

// CreateSecurityPolicy creates a Security Policy
func (s *SecurityService) CreateSecurityPolicy(ctx context.Context, req *CreateEACIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	req.CIType = "EA.Security-Policy"

	if req.Attributes == nil {
		req.Attributes = make(map[string]interface{})
	}

	return s.CreateEACI(ctx, req, userID)
}

// ListSecurityControls retrieves all Security Controls
func (s *SecurityService) ListSecurityControls(ctx context.Context) ([]*ci.ConfigurationItem, error) {
	response, err := s.ciService.ListCIs(ctx, ci.ListCIFilters{CIType: "EA.Security-Control"}, 1, 1000)
	if err != nil {
		return nil, err
	}
	return response.CIs, nil
}

// ListSecurityPolicies retrieves all Security Policies
func (s *SecurityService) ListSecurityPolicies(ctx context.Context) ([]*ci.ConfigurationItem, error) {
	response, err := s.ciService.ListCIs(ctx, ci.ListCIFilters{CIType: "EA.Security-Policy"}, 1, 1000)
	if err != nil {
		return nil, err
	}
	return response.CIs, nil
}
