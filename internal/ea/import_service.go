package ea

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gocarina/gocsv"
	"github.com/google/uuid"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
	"github.com/pustaka/pustaka/internal/ci"
)

// ImportService handles CSV import operations for EA entities
type ImportService struct {
	repo                 *Repository
	ciRepo               *ci.Repository
	lifecycleStatusRepo  *ci.LifecycleStatusRepository
	logger               *pustakaLogger.Logger
}

// NewImportService creates a new import service
func NewImportService(repo *Repository, ciRepo *ci.Repository, lifecycleStatusRepo *ci.LifecycleStatusRepository, logger *pustakaLogger.Logger) *ImportService {
	return &ImportService{
		repo:                repo,
		ciRepo:              ciRepo,
		lifecycleStatusRepo: lifecycleStatusRepo,
		logger:              logger,
	}
}

// ImportRow represents a single row in the import CSV
type ImportRow struct {
	RowNumber       int              `csv:"-"` // Added during parsing
	Name            string           `csv:"Name"`
	CIType          string           `csv:"CI_Type"`
	Domain          string           `csv:"Domain"`
	LifecycleStatus string           `csv:"Lifecycle_Status"`
	Owner           string           `csv:"Owner"`
	Team            string           `csv:"Team"`
	AttributesJSON  string           `csv:"Attributes"` // JSON string
	Tags            string           `csv:"Tags"`       // Comma-separated
	Attributes      map[string]interface{} `csv:"-"` // Parsed from AttributesJSON
}

// ImportError represents a validation error for a specific row
type ImportError struct {
	RowNumber      int    `json:"row_number"`
	FieldName      string `json:"field_name"`
	ErrorMessage   string `json:"error_message"`
	ExpectedFormat string `json:"expected_format,omitempty"`
	ActualValue    string `json:"actual_value,omitempty"`
}

// ImportResult represents the result of an import operation
type ImportResult struct {
	SuccessCount int          `json:"success_count"`
	ErrorCount   int          `json:"error_count"`
	Errors       []ImportError `json:"errors,omitempty"`
	RowErrors    map[int][]ValidationError `json:"row_errors,omitempty"`
	TotalRows    int          `json:"total_rows"`
}

// ParseCSV parses CSV file content into ImportRow slice
func (s *ImportService) ParseCSV(file []byte, ciType string) ([]ImportRow, error) {
	// Validate that the file is not empty
	if len(file) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// Parse CSV using gocsv
	var rows []ImportRow
	if err := gocsv.UnmarshalBytes(file, &rows); err != nil {
		s.logger.Error().Err(err).Msg("Failed to parse CSV file")
		return nil, fmt.Errorf("failed to parse CSV file: %w", err)
	}

	// Validate required columns exist
	if len(rows) > 0 {
		if rows[0].Name == "" {
			return nil, fmt.Errorf("missing required column: Name")
		}
		if rows[0].CIType == "" {
			return nil, fmt.Errorf("missing required column: CI_Type")
		}
		if rows[0].Domain == "" {
			return nil, fmt.Errorf("missing required column: Domain")
		}
	}

	// Add row numbers and parse attributes JSON
	for i := range rows {
		rows[i].RowNumber = i + 2 // +2 because CSV rows are 1-indexed and header is row 1

		// Parse Attributes JSON if provided
		if rows[i].AttributesJSON != "" {
			var attrs map[string]interface{}
			if err := json.Unmarshal([]byte(rows[i].AttributesJSON), &attrs); err != nil {
				s.logger.Warn().
					Int("row", rows[i].RowNumber).
					Err(err).
					Msg("Failed to parse Attributes JSON, will be treated as text")
				// Store as string if not valid JSON
				rows[i].Attributes = map[string]interface{}{
					"_raw": rows[i].AttributesJSON,
				}
			} else {
				rows[i].Attributes = attrs
			}
		}

		// Validate CI_Type matches requested ciType parameter
		if rows[i].CIType != ciType {
			return nil, fmt.Errorf("row %d: CI_Type mismatch (expected %s, got %s)",
				rows[i].RowNumber, ciType, rows[i].CIType)
		}
	}

	s.logger.Info().
		Int("row_count", len(rows)).
		Str("ci_type", ciType).
		Msg("CSV file parsed successfully")

	return rows, nil
}

// ValidateRow validates a single import row
func (s *ImportService) ValidateRow(ctx context.Context, row ImportRow, ciType *CITypeDefinition, rowNumber int) ([]ValidationError, error) {
	var errors []ValidationError

	// Validate required fields
	if row.Name == "" {
		errors = append(errors, ValidationError{
			Field:    "Name",
			Message:  "Name is required",
			Severity: "error",
		})
	}

	if row.CIType == "" {
		errors = append(errors, ValidationError{
			Field:    "CI_Type",
			Message:  "CI_Type is required",
			Severity: "error",
		})
	}

	if row.Domain == "" {
		errors = append(errors, ValidationError{
			Field:    "Domain",
			Message:  "Domain is required",
			Severity: "error",
		})
	}

	// Validate CI_Type exists in database
	if row.CIType != "" {
		dbCIType, err := s.repo.GetCITypeByName(ctx, row.CIType)
		if err != nil {
			errors = append(errors, ValidationError{
				Field:    "CI_Type",
				Message:  fmt.Sprintf("CI_Type '%s' does not exist in database", row.CIType),
				Severity: "error",
			})
		} else if dbCIType != nil {
			ciType = dbCIType // Use database version for validation
		}
	}

	// Validate Lifecycle_Status exists if provided
	if row.LifecycleStatus != "" {
		_, err := s.lifecycleStatusRepo.GetByName(ctx, row.LifecycleStatus)
		if err != nil {
			errors = append(errors, ValidationError{
				Field:    "Lifecycle_Status",
				Message:  fmt.Sprintf("Lifecycle_Status '%s' does not exist", row.LifecycleStatus),
				Severity: "error",
			})
		}
	}

	// Validate Owner exists if provided
	if row.Owner != "" {
		_, err := s.repo.GetTeamByName(ctx, row.Owner)
		if err != nil {
			errors = append(errors, ValidationError{
				Field:    "Owner",
				Message:  fmt.Sprintf("Owner team '%s' does not exist", row.Owner),
				Severity: "error",
			})
		}
	}

	// Validate Team exists if provided
	if row.Team != "" && row.Team != row.Owner {
		_, err := s.repo.GetTeamByName(ctx, row.Team)
		if err != nil {
			errors = append(errors, ValidationError{
				Field:    "Team",
				Message:  fmt.Sprintf("Team '%s' does not exist", row.Team),
				Severity: "error",
			})
		}
	}

	// Validate attributes against CI type schema
	if ciType != nil && len(row.Attributes) > 0 {
		entity := &EAEntity{
			Name:       row.Name,
			CIType:     row.CIType,
			Attributes: row.Attributes,
		}

		result, err := ValidateEntityAttributes(entity, ciType)
		if err != nil {
			return nil, fmt.Errorf("attribute validation error: %w", err)
		}

		if !result.IsValid {
			errors = append(errors, result.Errors...)
		}
	}

	return errors, nil
}

// ValidateImport validates all rows in the import
func (s *ImportService) ValidateImport(ctx context.Context, rows []ImportRow, ciTypeName string) (*ImportResult, error) {
	result := &ImportResult{
		TotalRows: len(rows),
		Errors:    []ImportError{},
		RowErrors: make(map[int][]ValidationError),
	}

	// Fetch CI type definition
	ciType, err := s.repo.GetCITypeByName(ctx, ciTypeName)
	if err != nil {
		return nil, fmt.Errorf("failed to get CI type: %w", err)
	}

	// Validate each row
	for _, row := range rows {
		validationErrors, err := s.ValidateRow(ctx, row, ciType, row.RowNumber)
		if err != nil {
			s.logger.Error().Err(err).Int("row", row.RowNumber).Msg("Validation error")
			return nil, fmt.Errorf("row %d: %w", row.RowNumber, err)
		}

		if len(validationErrors) > 0 {
			result.ErrorCount++
			result.RowErrors[row.RowNumber] = validationErrors

			// Convert to ImportError format
			for _, ve := range validationErrors {
				result.Errors = append(result.Errors, ImportError{
					RowNumber:      row.RowNumber,
					FieldName:      ve.Field,
					ErrorMessage:   ve.Message,
					ExpectedFormat: "",
					ActualValue:    "",
				})
			}
		} else {
			result.SuccessCount++
		}
	}

	s.logger.Info().
		Int("total_rows", result.TotalRows).
		Int("success_count", result.SuccessCount).
		Int("error_count", result.ErrorCount).
		Msg("Import validation completed")

	return result, nil
}

// BulkCreateEntities creates entities from valid import rows
func (s *ImportService) BulkCreateEntities(ctx context.Context, rows []ImportRow, ciTypeName string, userID string) (*ImportResult, error) {
	// First validate all rows
	validationResult, err := s.ValidateImport(ctx, rows, ciTypeName)
	if err != nil {
		return nil, err
	}

	// If there are validation errors, return them
	if validationResult.ErrorCount > 0 {
		return validationResult, nil
	}

	// Get CI type definition
	ciType, err := s.repo.GetCITypeByName(ctx, ciTypeName)
	if err != nil {
		return nil, fmt.Errorf("failed to get CI type: %w", err)
	}

	// Get default lifecycle status
	lifecycleStatus, err := s.lifecycleStatusRepo.GetByName(ctx, "Draft")
	if err != nil {
		// Try "Active" as fallback
		lifecycleStatus, err = s.lifecycleStatusRepo.GetByName(ctx, "Active")
		if err != nil {
			return nil, fmt.Errorf("failed to get default lifecycle status: %w", err)
		}
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Create entities
	successCount := 0
	errorCount := 0

	for _, row := range rows {
		// Prepare entity
		entity := &EAEntity{
			Name:       row.Name,
			CIType:     row.CIType,
			Attributes: row.Attributes,
			Tags:       parseTags(row.Tags),
			CreatedBy:  userUUID,
		}

		// Set lifecycle status
		if row.LifecycleStatus != "" {
			lc, err := s.lifecycleStatusRepo.GetByName(ctx, row.LifecycleStatus)
			if err == nil {
				entity.LifecycleStatusID = &lc.ID
			} else {
				// Use default
				entity.LifecycleStatusID = &lifecycleStatus.ID
			}
		} else {
			entity.LifecycleStatusID = &lifecycleStatus.ID
		}

		// Add EA metadata to attributes
		entity.Attributes["ea_domain"] = row.Domain
		if row.Owner != "" {
			entity.Attributes["ea_owner"] = row.Owner
		}
		if row.Team != "" {
			entity.Attributes["ea_team"] = row.Team
		}

		// Validate entity
		vResult, err := ValidateEntityAttributes(entity, ciType)
		if err == nil && vResult.IsValid {
			crossFieldErrors := ValidateCrossFieldRules(entity, ciType)
			if len(crossFieldErrors) == 0 {
				entity.DataQualityScore = vResult.DataQualityScore
			}
		}

		// Create entity
		_, err = s.repo.Create(ctx, entity)
		if err != nil {
			s.logger.Error().Err(err).
				Int("row", row.RowNumber).
				Str("name", row.Name).
				Msg("Failed to create entity from import")
			errorCount++
		} else {
			successCount++
		}
	}

	result := &ImportResult{
		SuccessCount: successCount,
		ErrorCount:   errorCount,
		TotalRows:    len(rows),
	}

	s.logger.Info().
		Int("total_rows", len(rows)).
		Int("success_count", successCount).
		Int("error_count", errorCount).
		Msg("Bulk import completed")

	return result, nil
}

// GenerateErrorCSV generates a CSV file containing error details
func (s *ImportService) GenerateErrorCSV(result *ImportResult) ([]byte, error) {
	if len(result.Errors) == 0 {
		return nil, fmt.Errorf("no errors to export")
	}

	// Create CSV buffer
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{"RowNumber", "FieldName", "ErrorMessage", "ExpectedFormat", "ActualValue"}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write error rows
	for _, err := range result.Errors {
		row := []string{
			strconv.Itoa(err.RowNumber),
			err.FieldName,
			err.ErrorMessage,
			err.ExpectedFormat,
			err.ActualValue,
		}
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateTemplate generates a CSV template for the given CI type
func (s *ImportService) GenerateTemplate(ctx context.Context, ciTypeName string) ([]byte, error) {
	// Fetch CI type definition
	ciType, err := s.repo.GetCITypeByName(ctx, ciTypeName)
	if err != nil {
		return nil, fmt.Errorf("CI type not found: %w", err)
	}

	// Create CSV buffer
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Build header row
	header := []string{"Name*", "CI_Type*", "Domain*", "Lifecycle_Status", "Owner", "Team", "Tags"}

	// Add attribute columns for required attributes
	for _, attr := range ciType.RequiredAttributes {
		header = append(header, attr.Name)
	}

	// Write header
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write example row
	example := []string{
		fmt.Sprintf("Example %s", strings.TrimPrefix(ciTypeName, "EA.")),
		ciTypeName,
		strings.Split(ciTypeName, ".")[1], // Extract domain from CI type name
		"Active",
		"Business Architecture",
		"Enterprise Architecture",
		"tag1,tag2",
	}

	// Add example values for required attributes
	for _, attr := range ciType.RequiredAttributes {
		example = append(example, getExampleValue(attr.Type))
	}

	if err := writer.Write(example); err != nil {
		return nil, fmt.Errorf("failed to write example row: %w", err)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	s.logger.Info().
		Str("ci_type", ciTypeName).
		Msg("CSV template generated")

	return buf.Bytes(), nil
}

// parseTags parses comma-separated tags string
func parseTags(tagsStr string) []string {
	if tagsStr == "" {
		return []string{}
	}

	tags := strings.Split(tagsStr, ",")
	result := make([]string, 0, len(tags))

	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// getExampleValue returns an example value for a given attribute type
func getExampleValue(attrType string) string {
	switch attrType {
	case "string":
		return "example_value"
	case "integer":
		return "100"
	case "boolean":
		return "true"
	case "date":
		return "2026-01-01"
	case "array":
		return `["item1", "item2"]`
	case "object":
		return `{"key": "value"}`
	default:
		return "example"
	}
}
