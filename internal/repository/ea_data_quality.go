package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

// EADataQualityRepository handles data quality queries for EA entities
type EADataQualityRepository struct {
	db     *pgxpool.Pool
	logger *pustakaLogger.Logger
}

// NewEADataQualityRepository creates a new EA data quality repository
func NewEADataQualityRepository(db *pgxpool.Pool, logger *pustakaLogger.Logger) *EADataQualityRepository {
	return &EADataQualityRepository{
		db:     db,
		logger: logger,
	}
}

// DataQualityMetrics represents aggregated data quality metrics
type DataQualityMetrics struct {
	TotalEntities          int64              `json:"total_entities"`
	CompletenessPct        float64            `json:"completeness_pct"`
	StaleEntitiesCount     int                `json:"stale_entities_count"`
	EntitiesWithErrorsCount int               `json:"entities_with_errors_count"`
	LifecycleBreakdown     map[string]int     `json:"lifecycle_breakdown"`
	ErrorBreakdownByDomain map[string]int     `json:"error_breakdown_by_domain"`
}

// StaleEntityCriteria defines criteria for identifying stale entities
type StaleEntityCriteria struct {
	DaysThreshold     int  `json:"days_threshold"`      // Default: 90
	IncludeIncomplete bool `json:"include_incomplete"`  // Include incomplete entities
}

// EAEntitySummary represents a simplified EA entity for lists
type EAEntitySummary struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	CIType            string  `json:"ci_type"`
	EADomain          string  `json:"ea_domain"`
	DataQualityScore  float64 `json:"data_quality_score"`
	UpdatedAt         string  `json:"updated_at"`
}

// GetCompletenessMetrics calculates average completeness percentage for EA entities
// Completeness = (valid_attributes / total_required_attributes) * 100
func (r *EADataQualityRepository) GetCompletenessMetrics(ctx context.Context, domain string) (float64, error) {
	// For EA entities, data_quality_score is stored in attributes
	// We calculate the average of all non-null scores
	query := `
		SELECT COALESCE(AVG(
			CASE
				WHEN (attributes->>'data_quality_score') IS NOT NULL
				THEN (attributes->>'data_quality_score')::float
				ELSE NULL
			END
		), 0) as avg_completeness
		FROM configuration_items
		WHERE ci_type LIKE 'EA.%%'
	`

	args := []interface{}{}
	argIndex := 1

	if domain != "" {
		// Filter by EA domain extracted from ci_type
		// EA.Application-* -> Application domain
		query += fmt.Sprintf(" AND ci_type LIKE $%d", argIndex)
		args = append(args, "EA."+domain+"-%")
		argIndex++
	}

	var avgCompleteness float64
	err := r.db.QueryRow(ctx, query, args...).Scan(&avgCompleteness)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, map[string]interface{}{
			"domain": domain,
		})
		return 0, fmt.Errorf("failed to get completeness metrics: %w", err)
	}

	return avgCompleteness, nil
}

// GetStaleEntities returns EA entities that haven't been updated in the specified threshold
// Stale = active entities not updated in X days OR proposed/active entities with missing required fields
func (r *EADataQualityRepository) GetStaleEntities(ctx context.Context, criteria StaleEntityCriteria) ([]EAEntitySummary, int, error) {
	daysThreshold := criteria.DaysThreshold
	if daysThreshold <= 0 {
		daysThreshold = 90 // Default to 90 days
	}

	// Build WHERE clause for stale entities
	// 1. Active entities not updated in X days
	// 2. Proposed/Active entities with data_quality_score < 100 (incomplete)
	whereClause := fmt.Sprintf(`
		WHERE ci_type LIKE 'EA.%%'
		AND (
			-- Active entities not updated in threshold days
			(ls.name IN ('active', 'proposed') AND ci.updated_at < NOW() - INTERVAL '%d days')
	`, daysThreshold)

	if criteria.IncludeIncomplete {
		// Also include entities with low data quality scores
		whereClause += fmt.Sprintf(`
			OR
			-- Proposed/Active entities with incomplete data
			(ls.name IN ('proposed', 'active') AND
			 (attributes->>'data_quality_score')::float < 100)
		`)
	}

	whereClause += ")"

	// Get total count first
	countQuery := `
		SELECT COUNT(*)
		FROM configuration_items ci
		LEFT JOIN lifecycle_statuses ls ON ci.lifecycle_status_id = ls.id
	` + whereClause

	var totalCount int
	err := r.db.QueryRow(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return nil, 0, fmt.Errorf("failed to count stale entities: %w", err)
	}

	// Get the actual entities (oldest first)
	query := `
		SELECT
			ci.id,
			ci.name,
			ci.ci_type,
			COALESCE(ci.attributes->>'ea_domain',
				SUBSTRING(ci.ci_type FROM 4 FOR POSITION('-' IN ci.ci_type) - 4)) as ea_domain,
			COALESCE((ci.attributes->>'data_quality_score')::float, 0) as data_quality_score,
			ci.updated_at
		FROM configuration_items ci
		LEFT JOIN lifecycle_statuses ls ON ci.lifecycle_status_id = ls.id
	` + whereClause + `
		ORDER BY ci.updated_at ASC
		LIMIT 100
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return nil, 0, fmt.Errorf("failed to get stale entities: %w", err)
	}
	defer rows.Close()

	var entities []EAEntitySummary
	for rows.Next() {
		var entity EAEntitySummary
		err := rows.Scan(
			&entity.ID,
			&entity.Name,
			&entity.CIType,
			&entity.EADomain,
			&entity.DataQualityScore,
			&entity.UpdatedAt,
		)
		if err != nil {
			r.logger.ErrorDatabase("SCAN", "configuration_items", err, nil)
			return nil, 0, fmt.Errorf("failed to scan stale entity: %w", err)
		}
		entities = append(entities, entity)
	}

	return entities, totalCount, nil
}

// GetEntitiesWithErrors returns EA entities with data quality issues
// Errors = data_quality_score < 80 OR validation_errors present
func (r *EADataQualityRepository) GetEntitiesWithErrors(ctx context.Context, domain string) ([]EAEntitySummary, int, error) {
	whereClause := "WHERE ci_type LIKE 'EA.%%' AND ("
	args := []interface{}{}
	argIndex := 1

	// Quality score < 80 OR validation errors exist
	whereClause += fmt.Sprintf("(attributes->>'data_quality_score')::float < 80")

	if domain != "" {
		whereClause += fmt.Sprintf(" AND ci_type LIKE $%d", argIndex)
		args = append(args, "EA."+domain+"-%")
		argIndex++
	}

	whereClause += " OR (attributes->>'validation_errors') IS NOT NULL"
	whereClause += ")"

	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM configuration_items ci
	` + whereClause

	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, map[string]interface{}{
			"domain": domain,
		})
		return nil, 0, fmt.Errorf("failed to count entities with errors: %w", err)
	}

	// Get the actual entities (worst first)
	query := `
		SELECT
			ci.id,
			ci.name,
			ci.ci_type,
			COALESCE(ci.attributes->>'ea_domain',
				SUBSTRING(ci.ci_type FROM 4 FOR POSITION('-' IN ci.ci_type) - 4)) as ea_domain,
			COALESCE((ci.attributes->>'data_quality_score')::float, 0) as data_quality_score,
			ci.updated_at
		FROM configuration_items ci
	` + whereClause + `
		ORDER BY (attributes->>'data_quality_score')::float ASC
		LIMIT 100
	`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return nil, 0, fmt.Errorf("failed to get entities with errors: %w", err)
	}
	defer rows.Close()

	var entities []EAEntitySummary
	for rows.Next() {
		var entity EAEntitySummary
		err := rows.Scan(
			&entity.ID,
			&entity.Name,
			&entity.CIType,
			&entity.EADomain,
			&entity.DataQualityScore,
			&entity.UpdatedAt,
		)
		if err != nil {
			r.logger.ErrorDatabase("SCAN", "configuration_items", err, nil)
			return nil, 0, fmt.Errorf("failed scan entity with errors: %w", err)
		}
		entities = append(entities, entity)
	}

	return entities, totalCount, nil
}

// GetLifecycleStatusBreakdown returns entity count grouped by lifecycle status
func (r *EADataQualityRepository) GetLifecycleStatusBreakdown(ctx context.Context, domain string) (map[string]int, error) {
	query := `
		SELECT
			COALESCE(ls.name, 'uncategorized') as status_name,
			COUNT(*) as count
		FROM configuration_items ci
		LEFT JOIN lifecycle_statuses ls ON ci.lifecycle_status_id = ls.id
		WHERE ci_type LIKE 'EA.%%'
	`

	args := []interface{}{}
	argIndex := 1

	if domain != "" {
		query += fmt.Sprintf(" AND ci_type LIKE $%d", argIndex)
		args = append(args, "EA."+domain+"-%")
		argIndex++
	}

	query += " GROUP BY ls.name ORDER BY count DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, map[string]interface{}{
			"domain": domain,
		})
		return nil, fmt.Errorf("failed to get lifecycle status breakdown: %w", err)
	}
	defer rows.Close()

	breakdown := make(map[string]int)
	for rows.Next() {
		var statusName string
		var count int
		err := rows.Scan(&statusName, &count)
		if err != nil {
			r.logger.ErrorDatabase("SCAN", "configuration_items", err, nil)
			return nil, fmt.Errorf("failed to scan lifecycle status: %w", err)
		}
		breakdown[statusName] = count
	}

	return breakdown, nil
}

// GetErrorBreakdownByDomain returns entity count with errors grouped by EA domain
func (r *EADataQualityRepository) GetErrorBreakdownByDomain(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT
			COALESCE(ci.attributes->>'ea_domain',
				SUBSTRING(ci.ci_type FROM 4 FOR POSITION('-' IN ci.ci_type) - 4)) as ea_domain,
			COUNT(*) as count
		FROM configuration_items ci
		WHERE ci_type LIKE 'EA.%%'
			AND ((attributes->>'data_quality_score')::float < 80
				OR (attributes->>'validation_errors') IS NOT NULL)
		GROUP BY ea_domain
		ORDER BY count DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return nil, fmt.Errorf("failed to get error breakdown by domain: %w", err)
	}
	defer rows.Close()

	breakdown := make(map[string]int)
	for rows.Next() {
		var domain string
		var count int
		err := rows.Scan(&domain, &count)
		if err != nil {
			r.logger.ErrorDatabase("SCAN", "configuration_items", err, nil)
			return nil, fmt.Errorf("failed to scan domain error: %w", err)
		}
		// Capitalize first letter for display
		if len(domain) > 0 {
			domain = strings.ToUpper(domain[:1]) + domain[1:]
		}
		breakdown[domain] = count
	}

	return breakdown, nil
}

// GetOverallMetrics aggregates all data quality metrics
func (r *EADataQualityRepository) GetOverallMetrics(ctx context.Context) (*DataQualityMetrics, error) {
	metrics := &DataQualityMetrics{
		LifecycleBreakdown:     make(map[string]int),
		ErrorBreakdownByDomain: make(map[string]int),
	}

	// Get total EA entities
	countQuery := "SELECT COUNT(*) FROM configuration_items WHERE ci_type LIKE 'EA.%%'"
	err := r.db.QueryRow(ctx, countQuery).Scan(&metrics.TotalEntities)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return nil, fmt.Errorf("failed to count total EA entities: %w", err)
	}

	if metrics.TotalEntities == 0 {
		return metrics, nil
	}

	// Get completeness metrics (all domains)
	completeness, err := r.GetCompletenessMetrics(ctx, "")
	if err != nil {
		return nil, err
	}
	metrics.CompletenessPct = completeness

	// Get stale entities count (90 days, including incomplete)
	_, staleCount, err := r.GetStaleEntities(ctx, StaleEntityCriteria{
		DaysThreshold:     90,
		IncludeIncomplete: true,
	})
	if err != nil {
		return nil, err
	}
	metrics.StaleEntitiesCount = staleCount

	// Get entities with errors count
	_, errorCount, err := r.GetEntitiesWithErrors(ctx, "")
	if err != nil {
		return nil, err
	}
	metrics.EntitiesWithErrorsCount = errorCount

	// Get lifecycle status breakdown
	metrics.LifecycleBreakdown, err = r.GetLifecycleStatusBreakdown(ctx, "")
	if err != nil {
		return nil, err
	}

	// Get error breakdown by domain
	metrics.ErrorBreakdownByDomain, err = r.GetErrorBreakdownByDomain(ctx)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}
