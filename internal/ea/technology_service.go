package ea

import (
	"context"

	"github.com/google/uuid"
	"github.com/pustaka/pustaka/internal/ci"
)

// TechnologyService handles Technology domain EA operations
type TechnologyService struct {
	*Service
}

// NewTechnologyService creates a new Technology domain service
func NewTechnologyService(baseService *Service) *TechnologyService {
	return &TechnologyService{Service: baseService}
}

// CreateTechnologyComponent creates a Technology Component
func (s *TechnologyService) CreateTechnologyComponent(ctx context.Context, req *CreateEACIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	req.CIType = "EA.Technology-Component"

	if req.Attributes == nil {
		req.Attributes = make(map[string]interface{})
	}

	return s.CreateEACI(ctx, req, userID)
}

// CreateTechnologyPlatform creates a Technology Platform
func (s *TechnologyService) CreateTechnologyPlatform(ctx context.Context, req *CreateEACIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	req.CIType = "EA.Technology-Platform"

	if req.Attributes == nil {
		req.Attributes = make(map[string]interface{})
	}

	return s.CreateEACI(ctx, req, userID)
}

// ListTechnologyComponents retrieves all Technology Components
func (s *TechnologyService) ListTechnologyComponents(ctx context.Context) ([]*ci.ConfigurationItem, error) {
	response, err := s.ciService.ListCIs(ctx, ci.ListCIFilters{CIType: "EA.Technology-Component"}, 1, 1000)
	if err != nil {
		return nil, err
	}
	return response.CIs, nil
}

// ListTechnologyPlatforms retrieves all Technology Platforms
func (s *TechnologyService) ListTechnologyPlatforms(ctx context.Context) ([]*ci.ConfigurationItem, error) {
	response, err := s.ciService.ListCIs(ctx, ci.ListCIFilters{CIType: "EA.Technology-Platform"}, 1, 1000)
	if err != nil {
		return nil, err
	}
	return response.CIs, nil
}
