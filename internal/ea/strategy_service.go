package ea

import (
	"context"

	"github.com/google/uuid"
	"github.com/pustaka/pustaka/internal/ci"
)

// StrategyService handles Strategy domain EA operations
type StrategyService struct {
	*Service
}

// NewStrategyService creates a new Strategy domain service
func NewStrategyService(baseService *Service) *StrategyService {
	return &StrategyService{Service: baseService}
}

// CreateStrategyObjective creates a Strategy Objective with domain-specific validation
func (s *StrategyService) CreateStrategyObjective(ctx context.Context, req *CreateEACIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	// Set CI type
	req.CIType = "EA.Strategy-Objective"

	// Domain-specific pre-validation
	if req.Attributes == nil {
		req.Attributes = make(map[string]interface{})
	}

	// Create via base EA service
	return s.CreateEACI(ctx, req, userID)
}

// CreateStrategyInitiative creates a Strategy Initiative
func (s *StrategyService) CreateStrategyInitiative(ctx context.Context, req *CreateEACIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	req.CIType = "EA.Strategy-Initiative"

	if req.Attributes == nil {
		req.Attributes = make(map[string]interface{})
	}

	return s.CreateEACI(ctx, req, userID)
}

// ListStrategyObjectives retrieves all Strategy Objectives
func (s *StrategyService) ListStrategyObjectives(ctx context.Context) ([]*ci.ConfigurationItem, error) {
	response, err := s.ciService.ListCIs(ctx, ci.ListCIFilters{CIType: "EA.Strategy-Objective"}, 1, 1000)
	if err != nil {
		return nil, err
	}
	// Convert value types to pointers
	result := make([]*ci.ConfigurationItem, len(response.CIs))
	for i := range response.CIs {
		result[i] = &response.CIs[i]
	}
	return result, nil
}

// ListStrategyInitiatives retrieves all Strategy Initiatives
func (s *StrategyService) ListStrategyInitiatives(ctx context.Context) ([]*ci.ConfigurationItem, error) {
	response, err := s.ciService.ListCIs(ctx, ci.ListCIFilters{CIType: "EA.Strategy-Initiative"}, 1, 1000)
	if err != nil {
		return nil, err
	}
	// Convert value types to pointers
	result := make([]*ci.ConfigurationItem, len(response.CIs))
	for i := range response.CIs {
		result[i] = &response.CIs[i]
	}
	return result, nil
}
