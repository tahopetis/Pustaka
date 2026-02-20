package ea

import (
	"context"

	"github.com/google/uuid"
	"github.com/pustaka/pustaka/internal/ci"
)

// ApplicationService handles Application domain EA operations
type ApplicationService struct {
	*Service
}

// NewApplicationService creates a new Application domain service
func NewApplicationService(baseService *Service) *ApplicationService {
	return &ApplicationService{Service: baseService}
}

// CreateBusinessApplication creates a Business Application
func (s *ApplicationService) CreateBusinessApplication(ctx context.Context, req *CreateEACIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	req.CIType = "EA.Application-BusinessApp"
	return s.CreateEACI(ctx, req, userID)
}

// CreateApplicationComponent creates an Application Component
func (s *ApplicationService) CreateApplicationComponent(ctx context.Context, req *CreateEACIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	req.CIType = "EA.Application-Component"
	return s.CreateEACI(ctx, req, userID)
}

// CreateApplicationInterface creates an Application Interface
func (s *ApplicationService) CreateApplicationInterface(ctx context.Context, req *CreateEACIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	req.CIType = "EA.Application-Interface"

	if req.Attributes == nil {
		req.Attributes = make(map[string]interface{})
	}

	return s.CreateEACI(ctx, req, userID)
}

// ListBusinessApplications retrieves all Business Applications
func (s *ApplicationService) ListBusinessApplications(ctx context.Context) ([]ci.ConfigurationItem, error) {
	response, err := s.ciService.ListCIs(ctx, ci.ListCIFilters{CIType: "EA.Application-BusinessApp"}, 1, 1000)
	if err != nil {
		return nil, err
	}
	return response.CIs, nil
}

// ListApplicationComponents retrieves all Application Components
func (s *ApplicationService) ListApplicationComponents(ctx context.Context) ([]ci.ConfigurationItem, error) {
	response, err := s.ciService.ListCIs(ctx, ci.ListCIFilters{CIType: "EA.Application-Component"}, 1, 1000)
	if err != nil {
		return nil, err
	}
	return response.CIs, nil
}
