package ci

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ConfigurationItem struct {
	ID        uuid.UUID            `json:"id" db:"id"`
	Name      string               `json:"name" db:"name"`
	CIType    string               `json:"ci_type" db:"ci_type"`
	Attributes map[string]interface{} `json:"attributes" db:"attributes"`
	Tags      []string             `json:"tags" db:"tags"`
	CreatedAt time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt time.Time            `json:"updated_at" db:"updated_at"`
	CreatedBy uuid.UUID            `json:"created_by" db:"created_by"`
	UpdatedBy *uuid.UUID           `json:"updated_by,omitempty" db:"updated_by"`
}

type CITypeDefinition struct {
	ID                uuid.UUID              `json:"id" db:"id"`
	Name              string                 `json:"name" db:"name"`
	Description       *string                `json:"description,omitempty" db:"description"`
	RequiredAttributes []AttributeDefinition `json:"required_attributes" db:"required_attributes"`
	OptionalAttributes []AttributeDefinition `json:"optional_attributes" db:"optional_attributes"`
	CreatedBy         uuid.UUID              `json:"created_by" db:"created_by"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}

type AttributeDefinition struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Validation  *AttributeValidation   `json:"validation,omitempty"`
}

type AttributeValidation struct {
	Pattern    string      `json:"pattern,omitempty"`
	MinLength  *int        `json:"min_length,omitempty"`
	MaxLength  *int        `json:"max_length,omitempty"`
	Min        *int        `json:"min,omitempty"`
	Max        *int        `json:"max,omitempty"`
	Enum       []string    `json:"enum,omitempty"`
	Format     string      `json:"format,omitempty"`
}

type Relationship struct {
	ID              uuid.UUID            `json:"id" db:"id"`
	SourceID        uuid.UUID            `json:"source_id" db:"source_id"`
	TargetID        uuid.UUID            `json:"target_id" db:"target_id"`
	RelationshipType string              `json:"relationship_type" db:"relationship_type"`
	Attributes      map[string]interface{} `json:"attributes,omitempty" db:"attributes"`
	CreatedAt       time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt       *time.Time           `json:"updated_at,omitempty" db:"updated_at"`
	CreatedBy       uuid.UUID            `json:"created_by" db:"created_by"`
	UpdatedBy       *uuid.UUID           `json:"updated_by,omitempty" db:"updated_by"`
}

type CreateCIRequest struct {
	Name      string                 `json:"name" validate:"required"`
	CIType    string                 `json:"ci_type" validate:"required"`
	Attributes map[string]interface{} `json:"attributes" validate:"required"`
	Tags      []string               `json:"tags,omitempty"`
}

type UpdateCIRequest struct {
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Tags      []string               `json:"tags,omitempty"`
}

type CreateCITypeRequest struct {
	Name              string                 `json:"name" validate:"required"`
	Description       *string                `json:"description,omitempty"`
	RequiredAttributes []AttributeDefinition `json:"required_attributes" validate:"required,dive"`
	OptionalAttributes []AttributeDefinition `json:"optional_attributes,omitempty,dive"`
}

type UpdateCITypeRequest struct {
	Description       *string                `json:"description,omitempty"`
	RequiredAttributes []AttributeDefinition `json:"required_attributes,omitempty,dive"`
	OptionalAttributes []AttributeDefinition `json:"optional_attributes,omitempty,dive"`
}

type CreateRelationshipRequest struct {
	SourceID        uuid.UUID            `json:"source_id" validate:"required"`
	TargetID        uuid.UUID            `json:"target_id" validate:"required"`
	RelationshipType string              `json:"relationship_type" validate:"required"`
	Attributes      map[string]interface{} `json:"attributes,omitempty"`
}

type UpdateRelationshipRequest struct {
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

type CIListResponse struct {
	CIs        []ConfigurationItem    `json:"cis"`
	Pagination PaginationResponse    `json:"pagination"`
}

type CITypeListResponse struct {
	CITypes    []CITypeDefinition    `json:"ci_types"`
	Pagination PaginationResponse    `json:"pagination"`
}

type RelationshipListResponse struct {
	Relationships []Relationship       `json:"relationships"`
	Pagination   PaginationResponse    `json:"pagination"`
}

type PaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type ListCIFilters struct {
	CIType   string   `json:"ci_type,omitempty"`
	Search   string   `json:"search,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	CreatedBy string  `json:"created_by,omitempty"`
	Sort     string   `json:"sort,omitempty"`
	Order    string   `json:"order,omitempty"`
}

type ListRelationshipFilters struct {
	SourceID         *uuid.UUID `json:"source_id,omitempty"`
	TargetID         *uuid.UUID `json:"target_id,omitempty"`
	RelationshipType string     `json:"relationship_type,omitempty"`
	Search           string     `json:"search,omitempty"`
	Sort             string     `json:"sort,omitempty"`
	Order            string     `json:"order,omitempty"`
}

// ValidateAttributes validates CI attributes against a CI type definition
func (ciType *CITypeDefinition) ValidateAttributes(attributes map[string]interface{}) []ValidationError {
	var errors []ValidationError

	// Check required attributes
	for _, reqAttr := range ciType.RequiredAttributes {
		value, exists := attributes[reqAttr.Name]
		if !exists || value == nil {
			errors = append(errors, ValidationError{
				Field:   reqAttr.Name,
				Message: "required field is missing",
			})
			continue
		}

		// Validate field type and constraints
		if fieldErrors := validateField(reqAttr, value); len(fieldErrors) > 0 {
			errors = append(errors, fieldErrors...)
		}
	}

	// Check optional attributes (if provided)
	for _, optAttr := range ciType.OptionalAttributes {
		value, exists := attributes[optAttr.Name]
		if !exists || value == nil {
			continue
		}

		// Validate field type and constraints
		if fieldErrors := validateField(optAttr, value); len(fieldErrors) > 0 {
			errors = append(errors, fieldErrors...)
		}
	}

	// Check for unknown attributes
	knownAttrs := make(map[string]bool)
	for _, attr := range ciType.RequiredAttributes {
		knownAttrs[attr.Name] = true
	}
	for _, attr := range ciType.OptionalAttributes {
		knownAttrs[attr.Name] = true
	}

	for attrName := range attributes {
		if !knownAttrs[attrName] {
			errors = append(errors, ValidationError{
				Field:   attrName,
				Message: "unknown attribute for this CI type",
			})
		}
	}

	return errors
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func validateArrayField(attrDef AttributeDefinition, value []interface{}) []ValidationError {
	var errors []ValidationError
	validation := attrDef.Validation

	if validation == nil {
		return errors
	}

	// Min length validation for arrays
	if validation.MinLength != nil && len(value) < *validation.MinLength {
		errors = append(errors, ValidationError{
			Field:   attrDef.Name,
			Message: fmt.Sprintf("array must have at least %d items", *validation.MinLength),
		})
	}

	// Max length validation for arrays
	if validation.MaxLength != nil && len(value) > *validation.MaxLength {
		errors = append(errors, ValidationError{
			Field:   attrDef.Name,
			Message: fmt.Sprintf("array must have at most %d items", *validation.MaxLength),
		})
	}

	// Enum validation for array items
	if len(validation.Enum) > 0 {
		for i, item := range value {
			if strItem, ok := item.(string); ok {
				found := false
				for _, enumValue := range validation.Enum {
					if strItem == enumValue {
						found = true
						break
					}
				}
				if !found {
					errors = append(errors, ValidationError{
						Field:   fmt.Sprintf("%s[%d]", attrDef.Name, i),
						Message: fmt.Sprintf("value must be one of: %v", validation.Enum),
					})
				}
			}
		}
	}

	return errors
}

func validateObjectField(attrDef AttributeDefinition, value map[string]interface{}) []ValidationError {
	var errors []ValidationError
	validation := attrDef.Validation

	if validation == nil {
		return errors
	}

	// Min length validation for objects (number of properties)
	if validation.MinLength != nil && len(value) < *validation.MinLength {
		errors = append(errors, ValidationError{
			Field:   attrDef.Name,
			Message: fmt.Sprintf("object must have at least %d properties", *validation.MinLength),
		})
	}

	// Max length validation for objects (number of properties)
	if validation.MaxLength != nil && len(value) > *validation.MaxLength {
		errors = append(errors, ValidationError{
			Field:   attrDef.Name,
			Message: fmt.Sprintf("object must have at most %d properties", *validation.MaxLength),
		})
	}

	return errors
}

func validateField(attrDef AttributeDefinition, value interface{}) []ValidationError {
	var errors []ValidationError

	// Type validation
	switch attrDef.Type {
	case "string":
		if _, ok := value.(string); !ok {
			errors = append(errors, ValidationError{
				Field:   attrDef.Name,
				Message: "must be a string",
			})
			return errors
		}
	case "integer":
		if _, ok := value.(float64); !ok {
			errors = append(errors, ValidationError{
				Field:   attrDef.Name,
				Message: "must be an integer",
			})
			return errors
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			errors = append(errors, ValidationError{
				Field:   attrDef.Name,
				Message: "must be a boolean",
			})
			return errors
		}
	case "array":
		if arrayValue, ok := value.([]interface{}); !ok {
			errors = append(errors, ValidationError{
				Field:   attrDef.Name,
				Message: "must be an array",
			})
			return errors
		} else {
			// Validate array specific rules
			if fieldErrors := validateArrayField(attrDef, arrayValue); len(fieldErrors) > 0 {
				errors = append(errors, fieldErrors...)
			}
		}
	case "object":
		if objectValue, ok := value.(map[string]interface{}); !ok {
			errors = append(errors, ValidationError{
				Field:   attrDef.Name,
				Message: "must be an object",
			})
			return errors
		} else {
			// Validate object specific rules
			if fieldErrors := validateObjectField(attrDef, objectValue); len(fieldErrors) > 0 {
				errors = append(errors, fieldErrors...)
			}
		}
	}

	// Type-specific validation rules
	if attrDef.Validation != nil {
		if strValue, ok := value.(string); ok {
			validationErrors := validateStringField(attrDef, strValue)
			errors = append(errors, validationErrors...)
		} else if intValue, ok := value.(float64); ok {
			validationErrors := validateIntegerField(attrDef, int(intValue))
			errors = append(errors, validationErrors...)
		}
	}

	return errors
}

func validateStringField(attrDef AttributeDefinition, value string) []ValidationError {
	var errors []ValidationError
	validation := attrDef.Validation

	if validation == nil {
		return errors
	}

	// Min length validation
	if validation.MinLength != nil && len(value) < *validation.MinLength {
		errors = append(errors, ValidationError{
			Field:   attrDef.Name,
			Message: fmt.Sprintf("minimum length is %d", *validation.MinLength),
		})
	}

	// Max length validation
	if validation.MaxLength != nil && len(value) > *validation.MaxLength {
		errors = append(errors, ValidationError{
			Field:   attrDef.Name,
			Message: fmt.Sprintf("maximum length is %d", *validation.MaxLength),
		})
	}

	// Pattern validation (regex)
	if validation.Pattern != "" {
		// Simple pattern validation - in production, use regexp package
		// For now, just check if pattern exists (basic implementation)
		errors = append(errors, ValidationError{
			Field:   attrDef.Name,
			Message: fmt.Sprintf("must match pattern: %s", validation.Pattern),
		})
	}

	// Format validation
	if validation.Format != "" {
		if formatErrors := validateFormat(attrDef.Name, value, validation.Format); len(formatErrors) > 0 {
			errors = append(errors, formatErrors...)
		}
	}

	// Enum validation
	if len(validation.Enum) > 0 {
		found := false
		for _, enumValue := range validation.Enum {
			if value == enumValue {
				found = true
				break
			}
		}
		if !found {
			errors = append(errors, ValidationError{
				Field:   attrDef.Name,
				Message: fmt.Sprintf("value must be one of: %v", validation.Enum),
			})
		}
	}

	return errors
}

func validateIntegerField(attrDef AttributeDefinition, value int) []ValidationError {
	var errors []ValidationError
	validation := attrDef.Validation

	if validation == nil {
		return errors
	}

	// Min validation
	if validation.Min != nil && value < *validation.Min {
		errors = append(errors, ValidationError{
			Field:   attrDef.Name,
			Message: fmt.Sprintf("minimum value is %d", *validation.Min),
		})
	}

	// Max validation
	if validation.Max != nil && value > *validation.Max {
		errors = append(errors, ValidationError{
			Field:   attrDef.Name,
			Message: fmt.Sprintf("maximum value is %d", *validation.Max),
		})
	}

	return errors
}

func validateFormat(fieldName, value, format string) []ValidationError {
	var errors []ValidationError

	switch format {
	case "email":
		if !isValidEmail(value) {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: "must be a valid email address",
			})
		}
	case "url":
		if !isValidURL(value) {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: "must be a valid URL",
			})
		}
	case "ipv4":
		if !isValidIPv4(value) {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: "must be a valid IPv4 address",
			})
		}
	case "date":
		if !isValidDate(value) {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: "must be a valid date (YYYY-MM-DD)",
			})
		}
	case "datetime":
		if !isValidDateTime(value) {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: "must be a valid datetime (ISO 8601)",
			})
		}
	}

	return errors
}

// Basic format validation functions (simplified implementations)
func isValidEmail(email string) bool {
	// Basic email validation - in production, use proper regex or email package
	return len(email) > 3 && len(email) < 254 &&
		   len(email) > 0 && email[0] != '@' &&
		   len(email) > 0 && email[len(email)-1] != '@' &&
		   contains(email, "@")
}

func isValidURL(url string) bool {
	// Basic URL validation
	return len(url) > 7 && (startsWith(url, "http://") || startsWith(url, "https://"))
}

func isValidIPv4(ip string) bool {
	// Basic IPv4 validation - simplified
	parts := splitBy(ip, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if len(part) == 0 || len(part) > 3 {
			return false
		}
		// Check if all characters are digits
		for _, c := range part {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}

func isValidDate(date string) bool {
	// Basic date validation (YYYY-MM-DD)
	if len(date) != 10 {
		return false
	}
	return date[4] == '-' && date[7] == '-'
}

func isValidDateTime(datetime string) bool {
	// Basic datetime validation (ISO 8601)
	if len(datetime) < 19 {
		return false
	}
	return datetime[4] == '-' && datetime[7] == '-' &&
		   datetime[10] == 'T' && datetime[13] == ':' && datetime[16] == ':'
}

// Helper functions for string operations (simplified implementations)
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func startsWith(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return s[:len(prefix)] == prefix
}

func splitBy(s, sep string) []string {
	var parts []string
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			parts = append(parts, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

// Graph Models

type RelationshipGraph struct {
	ID               uuid.UUID            `json:"id"`
	RelationshipType string               `json:"relationship_type"`
	Attributes       map[string]interface{} `json:"attributes"`
	CreatedAt        time.Time            `json:"created_at"`
	CreatedBy        uuid.UUID            `json:"created_by"`
	RelatedCI        ConfigurationItem    `json:"related_ci"`
}

type GraphNode struct {
	ID         uuid.UUID            `json:"id"`
	Name       string               `json:"name"`
	Type       string               `json:"type"`
	Attributes map[string]interface{} `json:"attributes"`
	Tags       []string             `json:"tags"`
}

type GraphEdge struct {
	ID               uuid.UUID            `json:"id"`
	Source           string               `json:"source"`
	Target           string               `json:"target"`
	RelationshipType string               `json:"relationship_type"`
	Attributes       map[string]interface{} `json:"attributes"`
}

type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

type GraphFilters struct {
	CITypes []string `json:"ci_types,omitempty"`
	Search  string   `json:"search,omitempty"`
	Limit   int      `json:"limit,omitempty"`
}

type CINetwork struct {
	Center uuid.UUID   `json:"center"`
	Nodes  []GraphNode `json:"nodes"`
	Edges  []GraphEdge `json:"edges"`
}

type CIImpact struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Depth     int       `json:"depth"`
	Direction string    `json:"direction"`
}

type ImpactAnalysis struct {
	CI         uuid.UUID  `json:"ci"`
	Downstream []CIImpact `json:"downstream"`
	Upstream   []CIImpact `json:"upstream"`
}

type CITypeUsage struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

type CIConnectivity struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	ConnectionCount int       `json:"connection_count"`
}