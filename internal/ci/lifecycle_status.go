package ci

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// LifecycleStatus represents a lifecycle status for CIs
type LifecycleStatus struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	DisplayName string     `json:"display_name" db:"display_name"`
	Description *string    `json:"description,omitempty" db:"description"`
	Color       *string    `json:"color,omitempty" db:"color"`
	Icon        *string    `json:"icon,omitempty" db:"icon"`
	SortOrder   int        `json:"sort_order" db:"sort_order"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	IsSystem    bool       `json:"is_system" db:"is_system"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty" db:"updated_at"`
	CreatedBy   uuid.UUID  `json:"created_by" db:"created_by"`
	UpdatedBy   *uuid.UUID `json:"updated_by,omitempty" db:"updated_by"`
}

// Request/Response DTOs for Lifecycle Statuses

// CreateLifecycleStatusRequest represents a request to create a new lifecycle status
type CreateLifecycleStatusRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=100"`
	DisplayName string  `json:"display_name" binding:"required,min=2,max=100"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=500"`
	Color       *string `json:"color,omitempty" binding:"omitempty,len=7"`
	Icon        *string `json:"icon,omitempty" binding:"omitempty,max=50"`
	SortOrder   *int    `json:"sort_order,omitempty" binding:"omitempty,min=0"`
}

// UpdateLifecycleStatusRequest represents a request to update an existing lifecycle status
type UpdateLifecycleStatusRequest struct {
	DisplayName *string `json:"display_name,omitempty" binding:"omitempty,min=2,max=100"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=500"`
	Color       *string `json:"color,omitempty" binding:"omitempty,len=7"`
	Icon        *string `json:"icon,omitempty" binding:"omitempty,max=50"`
	SortOrder   *int    `json:"sort_order,omitempty" binding:"omitempty,min=0"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// LifecycleStatusListResponse represents a paginated list of lifecycle statuses
type LifecycleStatusListResponse struct {
	LifecycleStatuses []LifecycleStatus `json:"lifecycle_statuses"`
	Page              int               `json:"page"`
	Limit             int               `json:"limit"`
	Total             int64             `json:"total"`
	TotalPages        int               `json:"total_pages"`
}

// ListLifecycleStatusFilters represents filtering options for listing lifecycle statuses
type ListLifecycleStatusFilters struct {
	Search   string  `json:"search,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
	IsSystem *bool   `json:"is_system,omitempty"`
	Sort     string  `json:"sort,omitempty"`
	Order    string  `json:"order,omitempty"`
}

// LifecycleStatusUsage represents usage statistics for a single lifecycle status
type LifecycleStatusUsage struct {
	LifecycleStatus LifecycleStatus `json:"lifecycle_status"`
	UsageCount     int64            `json:"usage_count"`
}

// LifecycleStatusUsageResponse represents usage statistics for all lifecycle statuses
type LifecycleStatusUsageResponse struct {
	TotalCIs          int64                    `json:"total_cis"`
	CIsWithStatus     int64                    `json:"cis_with_status"`
	CIsWithoutStatus  int64                    `json:"cis_without_status"`
	StatusUsage       []LifecycleStatusUsage   `json:"status_usage"`
	StatusDistribution map[string]int64        `json:"status_distribution"`
}

// CIStatusDistribution represents CI status distribution for dashboard
type CIStatusDistribution struct {
	StatusName     string `json:"status_name"`
	DisplayName    string `json:"display_name"`
	Color          string `json:"color"`
	Icon           string `json:"icon"`
	Count          int64  `json:"count"`
	Percentage     float64 `json:"percentage"`
}

// CIStatusDistributionByType represents CI status distribution grouped by CI type
type CIStatusDistributionByType struct {
	CITypeName string                `json:"ci_type_name"`
	Statuses   []CIStatusDistribution `json:"statuses"`
	TotalCIs   int64                 `json:"total_cis"`
}

// Helper methods

// NewLifecycleStatusFromRequest creates a new LifecycleStatus from a CreateLifecycleStatusRequest
func NewLifecycleStatusFromRequest(req *CreateLifecycleStatusRequest, createdBy uuid.UUID) *LifecycleStatus {
	now := time.Now()
	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	return &LifecycleStatus{
		ID:          uuid.New(),
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
		SortOrder:   sortOrder,
		IsActive:    true,
		IsSystem:    false,
		CreatedBy:   createdBy,
		CreatedAt:   now,
	}
}

// UpdateFromRequest updates a LifecycleStatus from an UpdateLifecycleStatusRequest
func (ls *LifecycleStatus) UpdateFromRequest(req *UpdateLifecycleStatusRequest, updatedBy uuid.UUID) {
	if req.DisplayName != nil {
		ls.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		ls.Description = req.Description
	}
	if req.Color != nil {
		ls.Color = req.Color
	}
	if req.Icon != nil {
		ls.Icon = req.Icon
	}
	if req.SortOrder != nil {
		ls.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		ls.IsActive = *req.IsActive
	}

	ls.UpdatedBy = &updatedBy
	now := time.Now()
	ls.UpdatedAt = &now
}

// CanBeModified checks if the lifecycle status can be modified
func (ls *LifecycleStatus) CanBeModified() bool {
	return !ls.IsSystem
}

// CanBeDeleted checks if the lifecycle status can be deleted
func (ls *LifecycleStatus) CanBeDeleted() bool {
	return !ls.IsSystem
}

// GetDisplayColor returns the display color or a default
func (ls *LifecycleStatus) GetDisplayColor() string {
	if ls.Color != nil {
		return *ls.Color
	}
	return "#6b7280" // Default gray color
}

// GetDisplayIcon returns the display icon or a default
func (ls *LifecycleStatus) GetDisplayIcon() string {
	if ls.Icon != nil {
		return *ls.Icon
	}
	return "circle" // Default icon
}

// Validate checks if the lifecycle status is valid
func (ls *LifecycleStatus) Validate() error {
	if ls.Name == "" {
		return fmt.Errorf("name is required")
	}
	if ls.DisplayName == "" {
		return fmt.Errorf("display name is required")
	}
	if ls.Color != nil && *ls.Color != "" {
		if len(*ls.Color) != 7 || (*ls.Color)[0] != '#' {
			return fmt.Errorf("color must be a valid hex color code (#RRGGBB)")
		}
	}
	return nil
}