package ci

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Relationship Type Models

// RelationshipType represents a type of relationship that can exist between CIs
type RelationshipType struct {
	ID                 uuid.UUID            `json:"id" db:"id"`
	Name               string               `json:"name" db:"name"`
	DisplayName        *string              `json:"display_name,omitempty" db:"display_name"`
	Description        *string              `json:"description,omitempty" db:"description"`
	ForwardLabel       string               `json:"forward_label" db:"forward_label"`
	ReverseLabel       string               `json:"reverse_label" db:"reverse_label"`
	Color              *string              `json:"color,omitempty" db:"color"`
	Icon               *string              `json:"icon,omitempty" db:"icon"`
	Category           *string              `json:"category,omitempty" db:"category"`
	IsActive           bool                 `json:"is_active" db:"is_active"`
	IsSystem           bool                 `json:"is_system" db:"is_system"`
	AllowedSourceTypes []string             `json:"allowed_source_types,omitempty" db:"allowed_source_types"`
	AllowedTargetTypes []string             `json:"allowed_target_types,omitempty" db:"allowed_target_types"`
	CardinalitySource  string               `json:"cardinality_source" db:"cardinality_source"`
	CardinalityTarget  string               `json:"cardinality_target" db:"cardinality_target"`
	Bidirectional      bool                 `json:"bidirectional" db:"bidirectional"`
	Attributes         map[string]interface{} `json:"attributes,omitempty" db:"attributes"`
	CreatedAt          time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt          *time.Time           `json:"updated_at,omitempty" db:"updated_at"`
	CreatedBy          uuid.UUID            `json:"created_by" db:"created_by"`
	UpdatedBy          *uuid.UUID           `json:"updated_by,omitempty" db:"updated_by"`
}

// RelationshipTypeAttributeSchema defines the schema for relationship type attributes
type RelationshipTypeAttributeSchema struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Description  string                 `json:"description"`
	Required     bool                   `json:"required"`
	DefaultValue interface{}            `json:"default_value,omitempty"`
	Validation   *AttributeValidation   `json:"validation,omitempty"`
}

// Request/Response DTOs for Relationship Types

// CreateRelationshipTypeRequest represents a request to create a new relationship type
type CreateRelationshipTypeRequest struct {
	Name               string                 `json:"name" binding:"required,min=2,max=100"`
	DisplayName        *string                `json:"display_name,omitempty"`
	Description        *string                `json:"description,omitempty"`
	ForwardLabel       string                 `json:"forward_label" binding:"required,min=1,max=100"`
	ReverseLabel       string                 `json:"reverse_label" binding:"required,min=1,max=100"`
	Color              *string                `json:"color,omitempty" binding:"omitempty,len=7"`
	Icon               *string                `json:"icon,omitempty" binding:"omitempty,max=50"`
	Category           *string                `json:"category,omitempty" binding:"omitempty,max=50"`
	AllowedSourceTypes []string               `json:"allowed_source_types,omitempty"`
	AllowedTargetTypes []string               `json:"allowed_target_types,omitempty"`
	CardinalitySource  string                 `json:"cardinality_source" binding:"required,oneof=one many"`
	CardinalityTarget  string                 `json:"cardinality_target" binding:"required,oneof=one many"`
	Bidirectional      bool                   `json:"bidirectional"`
	Attributes         map[string]interface{} `json:"attributes,omitempty"`
}

// UpdateRelationshipTypeRequest represents a request to update an existing relationship type
type UpdateRelationshipTypeRequest struct {
	DisplayName        *string                `json:"display_name,omitempty" binding:"omitempty,min=2,max=100"`
	Description        *string                `json:"description,omitempty"`
	ForwardLabel       *string                `json:"forward_label,omitempty" binding:"omitempty,min=1,max=100"`
	ReverseLabel       *string                `json:"reverse_label,omitempty" binding:"omitempty,min=1,max=100"`
	Color              *string                `json:"color,omitempty" binding:"omitempty,len=7"`
	Icon               *string                `json:"icon,omitempty" binding:"omitempty,max=50"`
	Category           *string                `json:"category,omitempty" binding:"omitempty,max=50"`
	IsActive           *bool                  `json:"is_active,omitempty"`
	AllowedSourceTypes []string               `json:"allowed_source_types,omitempty"`
	AllowedTargetTypes []string               `json:"allowed_target_types,omitempty"`
	CardinalitySource  *string                `json:"cardinality_source,omitempty" binding:"omitempty,oneof=one many"`
	CardinalityTarget  *string                `json:"cardinality_target,omitempty" binding:"omitempty,oneof=one many"`
	Bidirectional      *bool                  `json:"bidirectional,omitempty"`
	Attributes         map[string]interface{} `json:"attributes,omitempty"`
}

// RelationshipTypeListResponse represents a paginated list of relationship types
type RelationshipTypeListResponse struct {
	RelationshipTypes []RelationshipType `json:"relationship_types"`
	Page              int                `json:"page"`
	Limit             int                `json:"limit"`
	Total             int64              `json:"total"`
	TotalPages        int                `json:"total_pages"`
}

// ListRelationshipTypeFilters represents filtering options for listing relationship types
type ListRelationshipTypeFilters struct {
	Search   string  `json:"search,omitempty"`
	Category string  `json:"category,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
	IsSystem *bool   `json:"is_system,omitempty"`
	Sort     string  `json:"sort,omitempty"`
	Order    string  `json:"order,omitempty"`
}

// RelationshipTypeStatistics provides usage statistics for relationship types
type RelationshipTypeStatistics struct {
	TotalTypes          int64                    `json:"total_types"`
	ActiveTypes         int64                    `json:"active_types"`
	SystemTypes         int64                    `json:"system_types"`
	CustomTypes         int64                    `json:"custom_types"`
	TotalRelationships  int64                    `json:"total_relationships"`
	Categories          map[string]int64         `json:"categories"`
	CardinalityDistribution map[string]int64    `json:"cardinality_distribution"`
}

// RelationshipTypeUsage represents usage statistics for a single relationship type
type RelationshipTypeUsage struct {
	RelationshipType RelationshipType `json:"relationship_type"`
	UsageCount        int64            `json:"usage_count"`
}

// RelationshipCompatibilityRequest represents a request to validate relationship compatibility
type RelationshipCompatibilityRequest struct {
	RelationshipType string `json:"relationship_type" binding:"required"`
	SourceType       string `json:"source_type" binding:"required"`
	TargetType       string `json:"target_type" binding:"required"`
}

// RelationshipCompatibilityResponse represents the result of relationship compatibility validation
type RelationshipCompatibilityResponse struct {
	IsValid bool   `json:"is_valid"`
	Message string `json:"message,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

// RelationshipTypeValidation represents validation results for relationship creation
type RelationshipTypeValidation struct {
	IsValid     bool     `json:"is_valid"`
	Errors      []string `json:"errors,omitempty"`
	Warnings    []string `json:"warnings,omitempty"`
	Constraints []string `json:"constraints,omitempty"`
}

// Helper methods

func (rt *RelationshipType) validateCardinality(sourceCI, targetCI *ConfigurationItem) error {
	// This is a simplified cardinality validation
	// In a full implementation, you would check existing relationships
	// to enforce cardinality constraints

	// Combine source and target cardinality to determine relationship type
	cardinality := rt.CardinalitySource + "-to-" + rt.CardinalityTarget

	switch cardinality {
	case "one-to-one":
		// Check if either CI already has a relationship of this type
		// This would require additional database queries
	case "one-to-many":
		// Check if target CI already has a relationship of this type where it's the "one" side
	case "many-to-one":
		// Check if source CI already has a relationship of this type where it's the "one" side
	case "many-to-many":
		// No cardinality restrictions
	}

	return nil
}

func (rt *RelationshipType) GetLabel(isForward bool, sourceCI, targetCI *ConfigurationItem) string {
	if isForward {
		return rt.ForwardLabel
	}
	return rt.ReverseLabel
}

func (rt *RelationshipType) CanBeModified() bool {
	return !rt.IsSystem
}

func (rt *RelationshipType) CanBeDeleted() bool {
	return !rt.IsSystem
}

func (rt *RelationshipType) ValidateRelationship(sourceCI, targetCI *ConfigurationItem) *RelationshipTypeValidation {
	validation := &RelationshipTypeValidation{
		IsValid: true,
	}

	// Check if relationship type is active
	if !rt.IsActive {
		validation.IsValid = false
		validation.Errors = append(validation.Errors, "Relationship type is not active")
	}

	// Check source type constraints
	if len(rt.AllowedSourceTypes) > 0 {
		allowed := false
		for _, allowedType := range rt.AllowedSourceTypes {
			if allowedType == sourceCI.CIType {
				allowed = true
				break
			}
		}
		if !allowed {
			validation.IsValid = false
			validation.Errors = append(validation.Errors, fmt.Sprintf("Source CI type '%s' is not allowed", sourceCI.CIType))
		}
	}

	// Check target type constraints
	if len(rt.AllowedTargetTypes) > 0 {
		allowed := false
		for _, allowedType := range rt.AllowedTargetTypes {
			if allowedType == targetCI.CIType {
				allowed = true
				break
			}
		}
		if !allowed {
			validation.IsValid = false
			validation.Errors = append(validation.Errors, fmt.Sprintf("Target CI type '%s' is not allowed", targetCI.CIType))
		}
	}

	// Validate cardinality
	if err := rt.validateCardinality(sourceCI, targetCI); err != nil {
		validation.IsValid = false
		validation.Errors = append(validation.Errors, err.Error())
	}

	return validation
}

// Helper methods for relationship type management

func NewRelationshipTypeFromRequest(req *CreateRelationshipTypeRequest, createdBy uuid.UUID) *RelationshipType {
	now := time.Now()

	return &RelationshipType{
		ID:                 uuid.New(),
		Name:               req.Name,
		DisplayName:        req.DisplayName,
		Description:        req.Description,
		ForwardLabel:       req.ForwardLabel,
		ReverseLabel:       req.ReverseLabel,
		Color:              req.Color,
		Icon:               req.Icon,
		Category:           req.Category,
		AllowedSourceTypes: req.AllowedSourceTypes,
		AllowedTargetTypes: req.AllowedTargetTypes,
		CardinalitySource:  req.CardinalitySource,
		CardinalityTarget:  req.CardinalityTarget,
		Bidirectional:      req.Bidirectional,
		Attributes:         req.Attributes,
		IsActive:           true,
		IsSystem:           false,
		CreatedBy:          createdBy,
		CreatedAt:          now,
	}
}

func (rt *RelationshipType) UpdateFromRequest(req *UpdateRelationshipTypeRequest, updatedBy uuid.UUID) {

	if req.DisplayName != nil {
		rt.DisplayName = req.DisplayName
	}
	if req.Description != nil {
		rt.Description = req.Description
	}
	if req.ForwardLabel != nil {
		rt.ForwardLabel = *req.ForwardLabel
	}
	if req.ReverseLabel != nil {
		rt.ReverseLabel = *req.ReverseLabel
	}
	if req.Color != nil {
		rt.Color = req.Color
	}
	if req.Icon != nil {
		rt.Icon = req.Icon
	}
	if req.Category != nil {
		rt.Category = req.Category
	}
	if req.IsActive != nil {
		rt.IsActive = *req.IsActive
	}
	if req.AllowedSourceTypes != nil {
		rt.AllowedSourceTypes = req.AllowedSourceTypes
	}
	if req.AllowedTargetTypes != nil {
		rt.AllowedTargetTypes = req.AllowedTargetTypes
	}
	if req.CardinalitySource != nil {
		rt.CardinalitySource = *req.CardinalitySource
	}
	if req.CardinalityTarget != nil {
		rt.CardinalityTarget = *req.CardinalityTarget
	}
	if req.Bidirectional != nil {
		rt.Bidirectional = *req.Bidirectional
	}
	if req.Attributes != nil {
		rt.Attributes = req.Attributes
	}

	rt.UpdatedBy = &updatedBy
	now := time.Now()
	rt.UpdatedAt = &now
}