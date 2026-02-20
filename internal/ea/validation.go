package ea

import (
	"fmt"
	"strings"
	"time"
)

// ============================================================================
// Cross-Domain Relationship Validation
// ============================================================================

// allowedCrossDomainRelationships defines which relationship types are allowed
// between which EA domains (source -> target -> relationship types)
var allowedCrossDomainRelationships = map[EADomain]map[EADomain][]string{
	EADomainApplication: {
		EADomainBusiness:       {"supports", "realizes", "flows_to"},
		EADomainData:          {"accesses", "manipulates", "flows_to"},
		EADomainTechnology:     {"deployed_on", "uses", "depends_on", "runs_on"},
		EADomainInfrastructure: {"deployed_on", "runs_on"},
		EADomainApplication:    {"depends_on", "associated_with"},
	},
	EADomainBusiness: {
		EADomainStrategy:    {"supports", "aligned_with"},
		EADomainApplication: {"uses", "depends_on"},
		EADomainData:        {"accesses", "flows_to"},
		EADomainBusiness:    {"aggregates", "composes", "decomposes"},
	},
	EADomainData: {
		EADomainApplication: {"flows_to", "associated_with"},
		EADomainData:        {"derived_from", "associated_with"},
	},
	EADomainTechnology: {
		EADomainApplication:    {"supports", "implements"},
		EADomainTechnology:     {"depends_on", "associated_with"},
		EADomainInfrastructure: {"runs_on"},
	},
	EADomainInfrastructure: {
		EADomainApplication:    {"supports", "runs_on"},
		EADomainInfrastructure: {"contains", "connected_to"},
	},
	EADomainSecurity: {
		EADomainSecurity:   {"validates", "mitigates", "associated_with"},
		EADomainGovernance: {"enforces", "assesses"},
		EADomainApplication: {"validates", "assesses"},
		EADomainData:       {"protects", "accesses"},
	},
	EADomainGovernance: {
		EADomainBusiness:    {"governs", "conforms_to"},
		EADomainApplication: {"governs", "conforms_to"},
		EADomainData:        {"governs", "conforms_to"},
		EADomainSecurity:    {"assesses", "governs"},
	},
	EADomainStrategy: {
		EADomainBusiness:    {"aligned_with", "supports"},
		EADomainApplication: {"aligned_with"},
	},
}

// ValidateCrossDomainRelationship checks if a relationship is allowed between
// the source and target CI types
func ValidateCrossDomainRelationship(sourceType, targetType, relationshipType string) error {
	// Extract EA domains from CI type names
	sourceDomain, err := ExtractEADomain(sourceType)
	if err != nil {
		return fmt.Errorf("invalid source CI type: %w", err)
	}

	targetDomain, err := ExtractEADomain(targetType)
	if err != nil {
		return fmt.Errorf("invalid target CI type: %w", err)
	}

	// Check if source domain has any allowed targets
	allowedTargets, sourceExists := allowedCrossDomainRelationships[sourceDomain]
	if !sourceExists {
		return fmt.Errorf("source domain %s has no allowed relationship targets defined", sourceDomain)
	}

	// Check if source can connect to target domain
	allowedRelTypes, targetExists := allowedTargets[targetDomain]
	if !targetExists {
		return fmt.Errorf("source domain %s cannot connect to target domain %s", sourceDomain, targetDomain)
	}

	// Check if relationship type is allowed
	for _, allowedType := range allowedRelTypes {
		if allowedType == relationshipType {
			return nil // Valid relationship
		}
	}

	return fmt.Errorf("relationship type %s not allowed between %s and %s domains",
		relationshipType, sourceDomain, targetDomain)
}

// ============================================================================
// Domain-Specific Validation Functions
// ============================================================================

// ValidateBusinessAttributes validates Business domain-specific attributes
func ValidateBusinessAttributes(ciType string, attributes map[string]interface{}) error {
	switch {
	case strings.HasPrefix(ciType, "EA.Business-Capability"):
		return validateBusinessCapabilityAttributes(attributes)
	case ciType == "EA.Business-Process":
		return validateBusinessProcessAttributes(attributes)
	case ciType == "EA.Business-Service":
		return validateBusinessServiceAttributes(attributes)
	}
	return nil
}

// validateBusinessCapabilityAttributes validates business capability attributes
func validateBusinessCapabilityAttributes(attributes map[string]interface{}) error {
	// Check for required owner (should reference ea_teams)
	owner, exists := attributes["owner"]
	if !exists || owner == nil {
		return fmt.Errorf("business capabilities must have an owner (EA team)")
	}

	// Validate strategic_alignment enum if present
	if strategicAlignment, exists := attributes["strategic_alignment"]; exists {
		allowedValues := map[string]bool{"critical": true, "high": true, "medium": true, "low": true}
		str, ok := strategicAlignment.(string)
		if !ok || !allowedValues[str] {
			return fmt.Errorf("strategic_alignment must be one of: critical, high, medium, low")
		}
	}

	return nil
}

// validateBusinessProcessAttributes validates business process attributes
func validateBusinessProcessAttributes(attributes map[string]interface{}) error {
	// Business processes must have at least one input or output
	inputs, hasInputs := attributes["inputs"]
	outputs, hasOutputs := attributes["outputs"]

	inputsArr, ok := inputs.([]interface{})
	hasInputsValue := hasInputs && ok && len(inputsArr) > 0

	outputsArr, ok := outputs.([]interface{})
	hasOutputsValue := hasOutputs && ok && len(outputsArr) > 0

	if !hasInputsValue && !hasOutputsValue {
		return fmt.Errorf("business process must have at least one input or output")
	}

	return nil
}

// validateBusinessServiceAttributes validates business service attributes
func validateBusinessServiceAttributes(attributes map[string]interface{}) error {
	// Business services should have a service level definition
	if _, exists := attributes["service_level"]; !exists {
		// Warning only (warn-but-allow approach)
		// Log this but don't return error
	}
	return nil
}

// ValidateApplicationAttributes validates Application domain-specific attributes
func ValidateApplicationAttributes(ciType string, attributes map[string]interface{}) error {
	switch {
	case strings.HasPrefix(ciType, "EA.Application-BusinessApp"):
		return validateBusinessAppAttributes(attributes)
	case ciType == "EA.Application-Component":
		return validateApplicationComponentAttributes(attributes)
	}
	return nil
}

// validateBusinessAppAttributes validates business application attributes
func validateBusinessAppAttributes(attributes map[string]interface{}) error {
	// Validate lifecycle_status enum if present
	if lifecycleStatus, exists := attributes["lifecycle_status"]; exists {
		allowedValues := map[string]bool{"proposed": true, "active": true, "deprecated": true, "retired": true}
		status, ok := lifecycleStatus.(string)
		if !ok || !allowedValues[status] {
			return fmt.Errorf("lifecycle_status must be one of: proposed, active, deprecated, retired")
		}
	}

	// Validate criticality enum if present
	if criticality, exists := attributes["criticality"]; exists {
		allowedValues := map[string]bool{"critical": true, "high": true, "medium": true, "low": true}
		crit, ok := criticality.(string)
		if !ok || !allowedValues[crit] {
			return fmt.Errorf("criticality must be one of: critical, high, medium, low")
		}
	}

	return nil
}

// validateApplicationComponentAttributes validates application component attributes
func validateApplicationComponentAttributes(attributes map[string]interface{}) error {
	// Components should have a parent application or technology
	if _, hasParent := attributes["parent_application_id"]; !hasParent {
		if _, hasTechnology := attributes["technology"]; !hasTechnology {
			// Warning only (component can exist standalone)
		}
	}
	return nil
}

// ValidateDataAttributes validates Data domain-specific attributes
func ValidateDataAttributes(ciType string, attributes map[string]interface{}) error {
	switch {
	case ciType == "EA.Data-DataObject":
		return validateDataObjectAttributes(attributes)
	case ciType == "EA.Data-DataSet":
		return validateDataSetAttributes(attributes)
	}
	return nil
}

// validateDataObjectAttributes validates data object attributes
func validateDataObjectAttributes(attributes map[string]interface{}) error {
	// Validate data_classification enum if present
	if classification, exists := attributes["data_classification"]; exists {
		allowedValues := map[string]bool{"public": true, "internal": true, "confidential": true, "restricted": true}
		class, ok := classification.(string)
		if !ok || !allowedValues[class] {
			return fmt.Errorf("data_classification must be one of: public, internal, confidential, restricted")
		}
	}

	return nil
}

// validateDataSetAttributes validates data set attributes
func validateDataSetAttributes(attributes map[string]interface{}) error {
	// Data sets should contain data objects
	if objects, exists := attributes["data_objects"]; exists {
		objArr, ok := objects.([]interface{})
		if !ok || len(objArr) == 0 {
			return fmt.Errorf("data set must contain at least one data object")
		}
	}

	return nil
}

// ValidateTechnologyAttributes validates Technology domain-specific attributes
func ValidateTechnologyAttributes(ciType string, attributes map[string]interface{}) error {
	// Technology components should have version information
	if _, exists := attributes["version"]; !exists {
		// Warning only (version tracking recommended but not required)
	}

	// Check for end_of_support date if present
	if eos, exists := attributes["end_of_support"]; exists {
		// Validate date format (should be ISO 8601 date string)
		eosStr, ok := eos.(string)
		if !ok {
			return fmt.Errorf("end_of_support must be a date string")
		}
		if _, err := time.Parse("2006-01-02", eosStr); err != nil {
			return fmt.Errorf("end_of_support must be in YYYY-MM-DD format")
		}
	}

	return nil
}

// ValidateInfrastructureAttributes validates Infrastructure domain-specific attributes
func ValidateInfrastructureAttributes(ciType string, attributes map[string]interface{}) error {
	switch {
	case ciType == "EA.Infrastructure-Node":
		return validateNodeAttributes(attributes)
	case ciType == "EA.Infrastructure-Network":
		return validateNetworkAttributes(attributes)
	}
	return nil
}

// validateNodeAttributes validates infrastructure node attributes
func validateNodeAttributes(attributes map[string]interface{}) error {
	// Validate node_type enum if present
	if nodeType, exists := attributes["node_type"]; exists {
		allowedValues := map[string]bool{"physical": true, "virtual": true, "container": true}
		nt, ok := nodeType.(string)
		if !ok || !allowedValues[nt] {
			return fmt.Errorf("node_type must be one of: physical, virtual, container")
		}
	}

	// Nodes should have capacity information
	if _, hasCPU := attributes["cpu_cores"]; !hasCPU {
		// Warning only
	}

	return nil
}

// validateNetworkAttributes validates infrastructure network attributes
func validateNetworkAttributes(attributes map[string]interface{}) error {
	// Validate network_type enum if present
	if networkType, exists := attributes["network_type"]; exists {
		allowedValues := map[string]bool{"lan": true, "wan": true, "vlan": true, "vpn": true}
		nt, ok := networkType.(string)
		if !ok || !allowedValues[nt] {
			return fmt.Errorf("network_type must be one of: lan, wan, vlan, vpn")
		}
	}

	// Networks should have CIDR block
	if _, hasCIDR := attributes["cidr_block"]; !hasCIDR {
		// Warning only
	}

	return nil
}

// ValidateSecurityAttributes validates Security domain-specific attributes
func ValidateSecurityAttributes(ciType string, attributes map[string]interface{}) error {
	switch {
	case ciType == "EA.Security-Control":
		return validateControlAttributes(attributes)
	case ciType == "EA.Security-Policy":
		return validateSecurityPolicyAttributes(attributes)
	}
	return nil
}

// validateControlAttributes validates security control attributes
func validateControlAttributes(attributes map[string]interface{}) error {
	// Validate control_type enum if present
	if controlType, exists := attributes["control_type"]; exists {
		allowedValues := map[string]bool{"preventive": true, "detective": true, "corrective": true}
		ct, ok := controlType.(string)
		if !ok || !allowedValues[ct] {
			return fmt.Errorf("control_type must be one of: preventive, detective, corrective")
		}
	}

	return nil
}

// validateSecurityPolicyAttributes validates security policy attributes
func validateSecurityPolicyAttributes(attributes map[string]interface{}) error {
	// Policies should have approval and review dates
	if _, hasApproval := attributes["approval_date"]; !hasApproval {
		// Warning only
	}

	return nil
}

// ValidateGovernanceAttributes validates Governance domain-specific attributes
func ValidateGovernanceAttributes(ciType string, attributes map[string]interface{}) error {
	switch {
	case ciType == "EA.Governance-Policy":
		return validateGovernancePolicyAttributes(attributes)
	case ciType == "EA.Governance-Compliance":
		return validateComplianceAttributes(attributes)
	}
	return nil
}

// validateGovernancePolicyAttributes validates governance policy attributes
func validateGovernancePolicyAttributes(attributes map[string]interface{}) error {
	// Validate compliance_level enum if present
	if complianceLevel, exists := attributes["compliance_level"]; exists {
		allowedValues := map[string]bool{"mandatory": true, "recommended": true}
		cl, ok := complianceLevel.(string)
		if !ok || !allowedValues[cl] {
			return fmt.Errorf("compliance_level must be one of: mandatory, recommended")
		}
	}

	return nil
}

// validateComplianceAttributes validates compliance requirement attributes
func validateComplianceAttributes(attributes map[string]interface{}) error {
	// Validate compliance_status enum if present
	if status, exists := attributes["compliance_status"]; exists {
		allowedValues := map[string]bool{"compliant": true, "non_compliant": true, "in_progress": true}
		st, ok := status.(string)
		if !ok || !allowedValues[st] {
			return fmt.Errorf("compliance_status must be one of: compliant, non_compliant, in_progress")
		}
	}

	return nil
}

// ValidateStrategyAttributes validates Strategy domain-specific attributes
func ValidateStrategyAttributes(ciType string, attributes map[string]interface{}) error {
	// Strategy objectives should have target dates
	if _, hasTarget := attributes["target_date"]; !hasTarget {
		// Warning only (target dates recommended)
	}

	// Validate strategic_alignment enum if present
	if alignment, exists := attributes["strategic_alignment"]; exists {
		allowedValues := map[string]bool{"high": true, "medium": true, "low": true}
		al, ok := alignment.(string)
		if !ok || !allowedValues[al] {
			return fmt.Errorf("strategic_alignment must be one of: high, medium, low")
		}
	}

	return nil
}

// CalculateDataQualityScore calculates data quality score for an EA entity
func CalculateDataQualityScore(validAttributes int, totalAttributes int, validationErrors []string) int {
	if totalAttributes == 0 {
		return 100 // No attributes to validate = perfect score
	}

	score := (validAttributes * 100) / totalAttributes

	// Penalize for validation errors
	errorPenalty := len(validationErrors) * 5
	score -= errorPenalty

	if score < 0 {
		score = 0
	}

	return score
}

// ============================================================================
// Entity Attribute Validation (Schema-based)
// ============================================================================

// ValidateEntityAttributes validates entity attributes against CI type schema
func ValidateEntityAttributes(entity *EAEntity, ciType *CITypeDefinition) (*ValidationResult, error) {
	result := &ValidationResult{
		IsValid:          true,
		Errors:           []ValidationError{},
		DataQualityScore: 100,
		ValidAttributes:  0,
		TotalRequired:    len(ciType.RequiredAttributes),
	}

	// Validate required attributes
	for _, attrDef := range ciType.RequiredAttributes {
		value, exists := entity.Attributes[attrDef.Name]

		if !exists || value == nil {
			result.Errors = append(result.Errors, ValidationError{
				Field:    attrDef.Name,
				Message:  "Required attribute is missing",
				Severity: "error",
			})
			result.IsValid = false
			continue
		}

		// Validate data type
		if err := validateAttributeType(attrDef.Name, attrDef.Type, value, attrDef.Validation); err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Field:    attrDef.Name,
				Message:  err.Error(),
				Severity: "error",
			})
			result.IsValid = false
			continue
		}

		result.ValidAttributes++
	}

	// Validate optional attributes if present
	for _, attrDef := range ciType.OptionalAttributes {
		if value, exists := entity.Attributes[attrDef.Name]; exists && value != nil {
			if err := validateAttributeType(attrDef.Name, attrDef.Type, value, attrDef.Validation); err != nil {
				result.Errors = append(result.Errors, ValidationError{
					Field:    attrDef.Name,
					Message:  err.Error(),
					Severity: "warning", // Optional attributes are warnings
				})
			}
		}
	}

	// Calculate data quality score
	if result.TotalRequired > 0 {
		result.DataQualityScore = float64(result.ValidAttributes*100) / float64(result.TotalRequired)
	}

	return result, nil
}

// validateAttributeType validates an attribute value against its type definition
func validateAttributeType(name, attrType string, value interface{}, validation map[string]interface{}) error {
	switch attrType {
	case "string":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", value)
		}

		// Validate string constraints
		if validation != nil {
			if minLen, ok := validation["min_length"].(float64); ok {
				if len(str) < int(minLen) {
					return fmt.Errorf("string too short (min %d characters)", int(minLen))
				}
			}
			if maxLen, ok := validation["max_length"].(float64); ok {
				if len(str) > int(maxLen) {
					return fmt.Errorf("string too long (max %d characters)", int(maxLen))
				}
			}
			if pattern, ok := validation["pattern"].(string); ok {
				// Simple pattern matching (for production, use regex)
				if !strings.Contains(str, pattern) && pattern != "" {
					return fmt.Errorf("string does not match required pattern")
				}
			}
		}

	case "integer":
		var num float64
		switch v := value.(type) {
		case float64:
			num = v
		case int:
			num = float64(v)
		case int64:
			num = float64(v)
		default:
			return fmt.Errorf("expected integer, got %T", value)
		}

		// Validate numeric constraints
		if validation != nil {
			if min, ok := validation["min"].(float64); ok {
				if num < min {
					return fmt.Errorf("value too small (min %f)", min)
				}
			}
			if max, ok := validation["max"].(float64); ok {
				if num > max {
					return fmt.Errorf("value too large (max %f)", max)
				}
			}
		}

	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}

	case "date":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected date string, got %T", value)
		}

		// Validate ISO 8601 date format
		if _, err := time.Parse("2006-01-02", str); err != nil {
			if _, err := time.Parse(time.RFC3339, str); err != nil {
				return fmt.Errorf("invalid date format (expected YYYY-MM-DD or RFC3339)")
			}
		}

	case "array":
		arr, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("expected array, got %T", value)
		}

		// Validate array items if item_type is specified
		if validation != nil {
			if itemType, ok := validation["item_type"].(string); ok {
				for i, item := range arr {
					if err := validateAttributeType(fmt.Sprintf("%s[%d]", name, i), itemType, item, nil); err != nil {
						return fmt.Errorf("array item %d: %w", i, err)
					}
				}
			}
		}

	case "object":
		obj, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected object, got %T", value)
		}

		// Objects are validated as valid JSON structure
		if len(obj) == 0 {
			return fmt.Errorf("object cannot be empty")
		}

	default:
		return fmt.Errorf("unsupported attribute type: %s", attrType)
	}

	return nil
}

// ValidateCrossFieldRules validates EA domain-specific business rules
func ValidateCrossFieldRules(entity *EAEntity, ciType *CITypeDefinition) []ValidationError {
	var errors []ValidationError

	// Extract EA domain from CI type
	domain, err := ExtractEADomain(entity.CIType)
	if err != nil {
		return []ValidationError{{
			Field:    "ci_type",
			Message:  err.Error(),
			Severity: "error",
		}}
	}

	// Domain-specific validation
	switch domain {
	case EADomainBusiness:
		if err := ValidateBusinessAttributes(entity.CIType, entity.Attributes); err != nil {
			errors = append(errors, ValidationError{
				Field:    "attributes",
				Message:  err.Error(),
				Severity: "error",
			})
		}

	case EADomainApplication:
		if err := ValidateApplicationAttributes(entity.CIType, entity.Attributes); err != nil {
			errors = append(errors, ValidationError{
				Field:    "attributes",
				Message:  err.Error(),
				Severity: "error",
			})
		}

	case EADomainData:
		if err := ValidateDataAttributes(entity.CIType, entity.Attributes); err != nil {
			errors = append(errors, ValidationError{
				Field:    "attributes",
				Message:  err.Error(),
				Severity: "error",
			})
		}

	case EADomainTechnology:
		if err := ValidateTechnologyAttributes(entity.CIType, entity.Attributes); err != nil {
			errors = append(errors, ValidationError{
				Field:    "attributes",
				Message:  err.Error(),
				Severity: "error",
			})
		}

	case EADomainInfrastructure:
		if err := ValidateInfrastructureAttributes(entity.CIType, entity.Attributes); err != nil {
			errors = append(errors, ValidationError{
				Field:    "attributes",
				Message:  err.Error(),
				Severity: "error",
			})
		}

	case EADomainSecurity:
		if err := ValidateSecurityAttributes(entity.CIType, entity.Attributes); err != nil {
			errors = append(errors, ValidationError{
				Field:    "attributes",
				Message:  err.Error(),
				Severity: "error",
			})
		}

	case EADomainGovernance:
		if err := ValidateGovernanceAttributes(entity.CIType, entity.Attributes); err != nil {
			errors = append(errors, ValidationError{
				Field:    "attributes",
				Message:  err.Error(),
				Severity: "error",
			})
		}

	case EADomainStrategy:
		if err := ValidateStrategyAttributes(entity.CIType, entity.Attributes); err != nil {
			errors = append(errors, ValidationError{
				Field:    "attributes",
				Message:  err.Error(),
				Severity: "error",
			})
		}
	}

	return errors
}

