package ci

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

// LifecycleStatusService handles business logic for lifecycle statuses
type LifecycleStatusService struct {
	repo         *LifecycleStatusRepository
	ciRepo       *Repository
	redis        *redis.Client
	auditService *AuditService
	logger       *pustakaLogger.Logger
}

func NewLifecycleStatusService(
	repo *LifecycleStatusRepository,
	ciRepo *Repository,
	redis *redis.Client,
	auditService *AuditService,
	logger *pustakaLogger.Logger,
) *LifecycleStatusService {
	return &LifecycleStatusService{
		repo:         repo,
		ciRepo:       ciRepo,
		redis:        redis,
		auditService: auditService,
		logger:       logger,
	}
}

// CreateLifecycleStatus creates a new lifecycle status
func (s *LifecycleStatusService) CreateLifecycleStatus(ctx context.Context, req *CreateLifecycleStatusRequest, userID uuid.UUID) (*LifecycleStatus, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check for duplicate name
	existing, err := s.repo.GetByName(ctx, req.Name)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("lifecycle status '%s' already exists", req.Name)
	}

	// Create lifecycle status
	lifecycleStatus := NewLifecycleStatusFromRequest(req, userID)

	result, err := s.repo.Create(ctx, lifecycleStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to create lifecycle status: %w", err)
	}

	// Invalidate cache
	s.invalidateLifecycleStatusCache(ctx)

	// Log audit event
	s.logAuditEvent(ctx, "lifecycle_status", result.ID.String(), "create", userID.String(), map[string]interface{}{
		"name":          result.Name,
		"display_name":  result.DisplayName,
		"color":         result.Color,
		"sort_order":    result.SortOrder,
	})

	s.logger.InfoService("lifecycle_status", "create_lifecycle_status", map[string]interface{}{
		"lifecycle_status_id": result.ID,
		"name":                result.Name,
		"user_id":             userID,
	})

	return result, nil
}

// GetLifecycleStatus retrieves a lifecycle status by ID
func (s *LifecycleStatusService) GetLifecycleStatus(ctx context.Context, id uuid.UUID) (*LifecycleStatus, error) {
	// Try cache first
	if ls, err := s.getLifecycleStatusFromCache(ctx, id); err == nil {
		return ls, nil
	}

	// Get from database
	ls, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result
	s.cacheLifecycleStatus(ctx, ls)

	return ls, nil
}

// ListLifecycleStatuses retrieves lifecycle statuses with pagination and filtering
func (s *LifecycleStatusService) ListLifecycleStatuses(ctx context.Context, filters *ListLifecycleStatusFilters, page, limit int) (*LifecycleStatusListResponse, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	return s.repo.List(ctx, filters, page, limit)
}

// UpdateLifecycleStatus updates an existing lifecycle status
func (s *LifecycleStatusService) UpdateLifecycleStatus(ctx context.Context, id uuid.UUID, req *UpdateLifecycleStatusRequest, userID uuid.UUID) (*LifecycleStatus, error) {
	// Get existing lifecycle status
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if system status is being modified
	if existing.IsSystem {
		// For system statuses, only allow updating display_name, description, color, icon, sort_order, and is_active
		// Cannot change name for system statuses
		return nil, fmt.Errorf("system lifecycle statuses cannot be modified")
	}

	// Validate request
	if err := s.validateUpdateRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Update lifecycle status
	existing.UpdateFromRequest(req, userID)

	result, err := s.repo.Update(ctx, existing)
	if err != nil {
		return nil, fmt.Errorf("failed to update lifecycle status: %w", err)
	}

	// Invalidate cache
	s.invalidateLifecycleStatusCache(ctx)

	// Log audit event
	s.logAuditEvent(ctx, "lifecycle_status", result.ID.String(), "update", userID.String(), map[string]interface{}{
		"name":          result.Name,
		"display_name":  result.DisplayName,
		"color":         result.Color,
		"sort_order":    result.SortOrder,
		"is_active":     result.IsActive,
	})

	s.logger.InfoService("lifecycle_status", "update_lifecycle_status", map[string]interface{}{
		"lifecycle_status_id": result.ID,
		"name":                result.Name,
		"user_id":             userID,
	})

	return result, nil
}

// DeleteLifecycleStatus deletes a lifecycle status
func (s *LifecycleStatusService) DeleteLifecycleStatus(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// Get existing lifecycle status
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if system status
	if existing.IsSystem {
		return fmt.Errorf("cannot delete system lifecycle status")
	}

	// Check if status is in use
	usageCount, err := s.repo.CountCIsWithStatus(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check lifecycle status usage: %w", err)
	}

	if usageCount > 0 {
		return fmt.Errorf("cannot delete lifecycle status '%s' because it is used by %d CIs", existing.DisplayName, usageCount)
	}

	// Delete lifecycle status
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete lifecycle status: %w", err)
	}

	// Invalidate cache
	s.invalidateLifecycleStatusCache(ctx)

	// Log audit event
	s.logAuditEvent(ctx, "lifecycle_status", existing.ID.String(), "delete", userID.String(), map[string]interface{}{
		"name":          existing.Name,
		"display_name":  existing.DisplayName,
	})

	s.logger.InfoService("lifecycle_status", "delete_lifecycle_status", map[string]interface{}{
		"lifecycle_status_id": existing.ID,
		"name":                existing.Name,
		"user_id":             userID,
	})

	return nil
}

// GetActiveLifecycleStatuses retrieves all active lifecycle statuses
func (s *LifecycleStatusService) GetActiveLifecycleStatuses(ctx context.Context) ([]LifecycleStatus, error) {
	// Try cache first
	if statuses, err := s.getActiveLifecycleStatusesFromCache(ctx); err == nil {
		return statuses, nil
	}

	// Get from database
	statuses, err := s.repo.GetActive(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the result
	s.cacheActiveLifecycleStatuses(ctx, statuses)

	return statuses, nil
}

// GetLifecycleStatusUsage retrieves usage statistics for lifecycle statuses
func (s *LifecycleStatusService) GetLifecycleStatusUsage(ctx context.Context) (*LifecycleStatusUsageResponse, error) {
	return s.repo.GetUsageStats(ctx)
}

// GetCIStatusDistribution retrieves CI status distribution for dashboard
func (s *LifecycleStatusService) GetCIStatusDistribution(ctx context.Context) ([]CIStatusDistribution, error) {
	usage, err := s.repo.GetUsageStats(ctx)
	if err != nil {
		return nil, err
	}

	var distribution []CIStatusDistribution
	for _, statusUsage := range usage.StatusUsage {
		if statusUsage.UsageCount > 0 {
			percentage := float64(statusUsage.UsageCount) / float64(usage.CIsWithStatus) * 100
			distribution = append(distribution, CIStatusDistribution{
				StatusName:  statusUsage.LifecycleStatus.Name,
				DisplayName: statusUsage.LifecycleStatus.DisplayName,
				Color:       statusUsage.LifecycleStatus.GetDisplayColor(),
				Icon:        statusUsage.LifecycleStatus.GetDisplayIcon(),
				Count:       statusUsage.UsageCount,
				Percentage:  percentage,
			})
		}
	}

	return distribution, nil
}

// GetCIStatusDistributionByType retrieves CI status distribution grouped by CI type
func (s *LifecycleStatusService) GetCIStatusDistributionByType(ctx context.Context) ([]CIStatusDistributionByType, error) {
	// This is a simplified implementation that aggregates the data from the usage stats
	// In a full implementation, you would want to create a more efficient database query

	// For now, return an empty slice as this feature is not fully implemented
	// This can be extended later with proper database queries
	return []CIStatusDistributionByType{}, nil
}

// Validation methods

func (s *LifecycleStatusService) validateCreateRequest(req *CreateLifecycleStatusRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}

	// Validate name format (lowercase, numbers, underscores)
	if !isValidLifecycleStatusName(req.Name) {
		return fmt.Errorf("name must contain only lowercase letters, numbers, and underscores")
	}

	if req.DisplayName == "" {
		return fmt.Errorf("display name is required")
	}

	if req.Color != nil && *req.Color != "" {
		if len(*req.Color) != 7 || (*req.Color)[0] != '#' {
			return fmt.Errorf("color must be a valid hex color code (#RRGGBB)")
		}
	}

	return nil
}

func (s *LifecycleStatusService) validateUpdateRequest(req *UpdateLifecycleStatusRequest) error {
	if req.DisplayName != nil && *req.DisplayName == "" {
		return fmt.Errorf("display name cannot be empty")
	}

	if req.Color != nil && *req.Color != "" {
		if len(*req.Color) != 7 || (*req.Color)[0] != '#' {
			return fmt.Errorf("color must be a valid hex color code (#RRGGBB)")
		}
	}

	if req.SortOrder != nil && *req.SortOrder < 0 {
		return fmt.Errorf("sort order cannot be negative")
	}

	return nil
}

func isValidLifecycleStatusName(name string) bool {
	if len(name) < 2 || len(name) > 100 {
		return false
	}

	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}

	return true
}

// Cache methods

func (s *LifecycleStatusService) getLifecycleStatusFromCache(ctx context.Context, id uuid.UUID) (*LifecycleStatus, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("cache not available")
	}

	key := fmt.Sprintf("lifecycle_status:%s", id.String())
	data, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var ls LifecycleStatus
	if err := json.Unmarshal([]byte(data), &ls); err != nil {
		return nil, err
	}

	return &ls, nil
}

func (s *LifecycleStatusService) cacheLifecycleStatus(ctx context.Context, ls *LifecycleStatus) {
	if s.redis == nil {
		return
	}

	key := fmt.Sprintf("lifecycle_status:%s", ls.ID.String())
	data, _ := json.Marshal(ls)
	s.redis.Set(ctx, key, data, 5*time.Minute)
}

func (s *LifecycleStatusService) getActiveLifecycleStatusesFromCache(ctx context.Context) ([]LifecycleStatus, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("cache not available")
	}

	key := "lifecycle_statuses:active"
	data, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var statuses []LifecycleStatus
	if err := json.Unmarshal([]byte(data), &statuses); err != nil {
		return nil, err
	}

	return statuses, nil
}

func (s *LifecycleStatusService) cacheActiveLifecycleStatuses(ctx context.Context, statuses []LifecycleStatus) {
	if s.redis == nil {
		return
	}

	key := "lifecycle_statuses:active"
	data, _ := json.Marshal(statuses)
	s.redis.Set(ctx, key, data, 5*time.Minute)
}

func (s *LifecycleStatusService) invalidateLifecycleStatusCache(ctx context.Context) {
	if s.redis == nil {
		return
	}

	// Delete individual status cache keys
	keys, _ := s.redis.Keys(ctx, "lifecycle_status:*").Result()
	if len(keys) > 0 {
		s.redis.Del(ctx, keys...)
	}

	// Delete active statuses cache
	s.redis.Del(ctx, "lifecycle_statuses:active")
}

// Audit logging

func (s *LifecycleStatusService) logAuditEvent(ctx context.Context, resourceType, resourceID, action, userID string, details map[string]interface{}) {
	if s.auditService == nil {
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return
	}

	entityUUID, err := uuid.Parse(resourceID)
	if err != nil {
		return
	}

	s.auditService.CreateAuditLog(ctx, resourceType, &entityUUID, action, userUUID, details, "", "")
}