package ea

import (
	"context"

	"github.com/google/uuid"
	"github.com/pustaka/pustaka/internal/ci"
)

// InfrastructureService handles Infrastructure domain EA operations
type InfrastructureService struct {
	*Service
}

// NewInfrastructureService creates a new Infrastructure domain service
func NewInfrastructureService(baseService *Service) *InfrastructureService {
	return &InfrastructureService{Service: baseService}
}

// CreateInfrastructureNode creates an Infrastructure Node
func (s *InfrastructureService) CreateInfrastructureNode(ctx context.Context, req *CreateEACIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	req.CIType = "EA.Infrastructure-Node"

	if req.Attributes == nil {
		req.Attributes = make(map[string]interface{})
	}

	return s.CreateEACI(ctx, req, userID)
}

// CreateInfrastructureNetwork creates an Infrastructure Network
func (s *InfrastructureService) CreateInfrastructureNetwork(ctx context.Context, req *CreateEACIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	req.CIType = "EA.Infrastructure-Network"

	if req.Attributes == nil {
		req.Attributes = make(map[string]interface{})
	}

	return s.CreateEACI(ctx, req, userID)
}

// ListInfrastructureNodes retrieves all Infrastructure Nodes
func (s *InfrastructureService) ListInfrastructureNodes(ctx context.Context) ([]*ci.ConfigurationItem, error) {
	response, err := s.ciService.ListCIs(ctx, ci.ListCIFilters{CIType: "EA.Infrastructure-Node"}, 1, 1000)
	if err != nil {
		return nil, err
	}
	return response.CIs, nil
}

// ListInfrastructureNetworks retrieves all Infrastructure Networks
func (s *InfrastructureService) ListInfrastructureNetworks(ctx context.Context) ([]*ci.ConfigurationItem, error) {
	response, err := s.ciService.ListCIs(ctx, ci.ListCIFilters{CIType: "EA.Infrastructure-Network"}, 1, 1000)
	if err != nil {
		return nil, err
	}
	return response.CIs, nil
}
