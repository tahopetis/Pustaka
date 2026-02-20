package ea

import (
	"context"

	"github.com/google/uuid"
	"github.com/pustaka/pustaka/internal/ci"
)

// DataService handles Data domain EA operations
type DataService struct {
	*Service
}

// NewDataService creates a new Data domain service
func NewDataService(baseService *Service) *DataService {
	return &DataService{Service: baseService}
}

// CreateDataObject creates a Data Object
func (s *DataService) CreateDataObject(ctx context.Context, req *CreateEACIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	req.CIType = "EA.Data-DataObject"
	return s.CreateEACI(ctx, req, userID)
}

// CreateDataSet creates a Data Set
func (s *DataService) CreateDataSet(ctx context.Context, req *CreateEACIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	req.CIType = "EA.Data-DataSet"

	if req.Attributes == nil {
		req.Attributes = make(map[string]interface{})
	}

	return s.CreateEACI(ctx, req, userID)
}

// CreateDataEntity creates a Data Entity
func (s *DataService) CreateDataEntity(ctx context.Context, req *CreateEACIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	req.CIType = "EA.Data-Entity"

	if req.Attributes == nil {
		req.Attributes = make(map[string]interface{})
	}

	return s.CreateEACI(ctx, req, userID)
}

// ListDataObjects retrieves all Data Objects
func (s *DataService) ListDataObjects(ctx context.Context) ([]*ci.ConfigurationItem, error) {
	response, err := s.ciService.ListCIs(ctx, ci.ListCIFilters{CIType: "EA.Data-DataObject"}, 1, 1000)
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

// ListDataSets retrieves all Data Sets
func (s *DataService) ListDataSets(ctx context.Context) ([]*ci.ConfigurationItem, error) {
	response, err := s.ciService.ListCIs(ctx, ci.ListCIFilters{CIType: "EA.Data-DataSet"}, 1, 1000)
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
