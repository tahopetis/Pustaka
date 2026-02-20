package ea

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pustaka/pustaka/internal/ci"
)

// BusinessService handles Business domain EA operations
type BusinessService struct {
	*Service
}

// NewBusinessService creates a new Business domain service
func NewBusinessService(baseService *Service) *BusinessService {
	return &BusinessService{Service: baseService}
}

// CreateBusinessCapabilityL1 creates a Level 1 Business Capability
func (s *BusinessService) CreateBusinessCapabilityL1(ctx context.Context, req *CreateEACIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	req.CIType = "EA.Business-CapabilityL1"

	if req.Attributes == nil {
		req.Attributes = make(map[string]interface{})
	}

	return s.CreateEACI(ctx, req, userID)
}

// CreateBusinessCapabilityL2 creates a Level 2 Business Capability
func (s *BusinessService) CreateBusinessCapabilityL2(ctx context.Context, req *CreateEACIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	req.CIType = "EA.Business-CapabilityL2"

	if req.Attributes == nil {
		req.Attributes = make(map[string]interface{})
	}

	// Validate parent_capability_id if provided
	if parentID, exists := req.Attributes["parent_capability_id"]; exists {
		// Verify parent exists and is L1
		parentUUID, ok := parentID.(uuid.UUID)
		if !ok {
			// Try converting from string
			if parentStr, ok := parentID.(string); ok {
				var err error
				parentUUID, err = uuid.Parse(parentStr)
				if err != nil {
					return nil, fmt.Errorf("invalid parent_capability_id format")
				}
			} else {
				return nil, fmt.Errorf("invalid parent_capability_id type")
			}
		}

		parent, err := s.ciService.GetCI(ctx, parentUUID)
		if err != nil {
			return nil, fmt.Errorf("parent capability not found: %w", err)
		}
		if parent.CIType != "EA.Business-CapabilityL1" {
			return nil, fmt.Errorf("parent must be a Business Capability L1")
		}
	}

	return s.CreateEACI(ctx, req, userID)
}

// CreateBusinessProcess creates a Business Process
func (s *BusinessService) CreateBusinessProcess(ctx context.Context, req *CreateEACIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	req.CIType = "EA.Business-Process"
	return s.CreateEACI(ctx, req, userID)
}

// CreateBusinessService creates a Business Service
func (s *BusinessService) CreateBusinessService(ctx context.Context, req *CreateEACIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	req.CIType = "EA.Business-Service"

	if req.Attributes == nil {
		req.Attributes = make(map[string]interface{})
	}

	return s.CreateEACI(ctx, req, userID)
}

// ListBusinessCapabilities retrieves all Business Capabilities (L1 and L2)
func (s *BusinessService) ListBusinessCapabilities(ctx context.Context) ([]*ci.ConfigurationItem, error) {
	l1Response, err := s.ciService.ListCIs(ctx, ci.ListCIFilters{CIType: "EA.Business-CapabilityL1"}, 1, 1000)
	if err != nil {
		return nil, err
	}

	l2Response, err := s.ciService.ListCIs(ctx, ci.ListCIFilters{CIType: "EA.Business-CapabilityL2"}, 1, 1000)
	if err != nil {
		return nil, err
	}

	// Convert value types to pointers and merge results
	result := make([]*ci.ConfigurationItem, 0, len(l1Response.CIs)+len(l2Response.CIs))
	for i := range l1Response.CIs {
		result = append(result, &l1Response.CIs[i])
	}
	for i := range l2Response.CIs {
		result = append(result, &l2Response.CIs[i])
	}
	return result, nil
}

// ListBusinessProcesses retrieves all Business Processes
func (s *BusinessService) ListBusinessProcesses(ctx context.Context) ([]*ci.ConfigurationItem, error) {
	response, err := s.ciService.ListCIs(ctx, ci.ListCIFilters{CIType: "EA.Business-Process"}, 1, 1000)
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
