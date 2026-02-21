package ea

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EADomain represents the 8 EA domains
type EADomain string

const (
	EADomainStrategy       EADomain = "Strategy"
	EADomainBusiness       EADomain = "Business"
	EADomainApplication    EADomain = "Application"
	EADomainData           EADomain = "Data"
	EADomainTechnology     EADomain = "Technology"
	EADomainInfrastructure EADomain = "Infrastructure"
	EADomainSecurity       EADomain = "Security"
	EADomainGovernance     EADomain = "Governance"
)

// AllEADomains is a slice of all valid EA domains
var AllEADomains = []EADomain{
	EADomainStrategy,
	EADomainBusiness,
	EADomainApplication,
	EADomainData,
	EADomainTechnology,
	EADomainInfrastructure,
	EADomainSecurity,
	EADomainGovernance,
}

// String returns the string representation of EADomain
func (d EADomain) String() string {
	return string(d)
}

// IsValid checks if the domain is valid
func (d EADomain) IsValid() bool {
	for _, validDomain := range AllEADomains {
		if d == validDomain {
			return true
		}
	}
	return false
}

// EATeam represents an EA team for ownership
type EATeam struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   uuid.UUID `json:"created_by"`
}

// CreateEATeamRequest represents a request to create an EA team
type CreateEATeamRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"max=500"`
}

// UpdateEATeamRequest represents a request to update an EA team
type UpdateEATeamRequest struct {
	Name        string `json:"name" validate:"omitempty,min=3,max=100"`
	Description string `json:"description" validate:"omitempty,max=500"`
}

// CreateEACIRequest represents a request to create an EA entity
type CreateEACIRequest struct {
	Name                string                 `json:"name" validate:"required,min=3,max=100"`
	CIType              string                 `json:"ci_type" validate:"required"` // e.g., "EA.Application-BusinessApp"
	Description         string                 `json:"description" validate:"max=1000"`
	Owner               string                 `json:"owner" validate:"required"` // EA team name
	Attributes          map[string]interface{} `json:"attributes"`
	Tags                []string               `json:"tags"`
	LifecycleStatusID   uuid.UUID              `json:"lifecycle_status_id" validate:"required"`
	// Validation override (admin only)
	OverrideValidation   bool   `json:"override_validation"`
	OverrideJustification string `json:"override_justification" validate:"required_if=OverrideValidation true"`
}

// UpdateEACIRequest represents a request to update an EA entity
type UpdateEACIRequest struct {
	Name                string                 `json:"name" validate:"omitempty,min=3,max=100"`
	Description         string                 `json:"description" validate:"omitempty,max=1000"`
	Owner               string                 `json:"owner" validate:"omitempty"`
	Attributes          map[string]interface{} `json:"attributes"`
	Tags                []string               `json:"tags"`
	LifecycleStatusID   *uuid.UUID             `json:"lifecycle_status_id"`
	// Validation override (admin only)
	OverrideValidation   bool   `json:"override_validation"`
	OverrideJustification string `json:"override_justification" validate:"required_if=OverrideValidation true"`
}

// CreateEARelationshipRequest represents a request to create an EA relationship
type CreateEARelationshipRequest struct {
	SourceID         uuid.UUID              `json:"source_id" validate:"required"`
	TargetID         uuid.UUID              `json:"target_id" validate:"required"`
	RelationshipType string                 `json:"relationship_type" validate:"required"`
	Attributes       map[string]interface{} `json:"attributes"`
}

// DataQualityScore represents the data quality of an EA entity
type DataQualityScore struct {
	Score            int      `json:"score"`             // 0-100
	ValidAttributes  int      `json:"valid_attributes"`
	TotalAttributes  int      `json:"total_attributes"`
	ValidationErrors []string `json:"validation_errors"`
	LastAssessed     time.Time `json:"last_assessed"`
}

// EAMetadata represents standard EA metadata fields
type EAMetadata struct {
	Source            string    `json:"source"`              // manual, import, discovery
	LastUpdatedBy     string    `json:"last_updated_by"`     // user reference
	DataQualityScore  int       `json:"data_quality_score"`  // 0-100
	ValidationErrors  []string  `json:"validation_errors"`
	DocumentationURL  string    `json:"documentation_url,omitempty"`
	StaleSince        *time.Time `json:"stale_since,omitempty"`
	ReviewDate        *time.Time `json:"review_date,omitempty"`
}

// ExtractEADomain extracts the EA domain from a CI type name
// e.g., "EA.Application-BusinessApp" -> "Application"
func ExtractEADomain(ciType string) (EADomain, error) {
	if !startsWithCITypePrefix(ciType) {
		return "", ErrInvalidEACIType
	}

	// Remove "EA." prefix
	domainPart := trimPrefix(ciType, "EA.")

	// Extract domain (first part before "-")
	domain := extractDomainFromString(domainPart)

	domainEnum := EADomain(domain)
	if !domainEnum.IsValid() {
		return "", ErrInvalidEADomain
	}

	return domainEnum, nil
}

// Helper functions
func startsWithCITypePrefix(ciType string) bool {
	return len(ciType) > 3 && ciType[0:3] == "EA."
}

func trimPrefix(s, prefix string) string {
	if len(s) > len(prefix) && s[0:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}

func extractDomainFromString(s string) string {
	for i, r := range s {
		if r == '-' {
			return s[:i]
		}
	}
	return s
}

// ErrRelationshipsExist is returned when attempting to delete an entity with relationships
type ErrRelationshipsExist struct {
	Count int
}

func (e *ErrRelationshipsExist) Error() string {
	return fmt.Sprintf("cannot delete EA entity with existing relationships (%d relationships)", e.Count)
}

// ErrInvalidLifecycleTransition is returned when attempting an invalid lifecycle status transition
type ErrInvalidLifecycleTransition struct {
	Current string
	New     string
}

func (e *ErrInvalidLifecycleTransition) Error() string {
	return fmt.Sprintf("invalid lifecycle transition: %s → %s", e.Current, e.New)
}

// Errors
var (
	ErrInvalidEACIType  = fmt.Errorf("invalid EA CI type format")
	ErrInvalidEADomain  = fmt.Errorf("invalid EA domain")
	ErrTeamNotFound     = fmt.Errorf("EA team not found")
	ErrValidationFailed = fmt.Errorf("EA validation failed")
)
