package amortization

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

// repository implements the Repository interface
type repository struct {
	db     *pgxpool.Pool
	logger *pustakaLogger.Logger
}

// NewRepository creates a new amortization repository
func NewRepository(db *pgxpool.Pool, logger *pustakaLogger.Logger) Repository {
	return &repository{
		db:     db,
		logger: logger,
	}
}

// GetAmortizableCI retrieves a CI with amortization configuration
func (r *repository) GetAmortizableCI(ctx context.Context, ciID uuid.UUID) (*AmortizableCI, error) {
	query := `
		SELECT
			ci.id,
			ci.name,
			ci.ci_type,
			ctd.id as ci_type_id,
			ci.attributes,
			ci.tags,
			ci.lifecycle_status_id,
			COALESCE(ci.purchase_cost, 0) as purchase_cost,
			COALESCE(ci.salvage_value, 0) as salvage_value,
			ci.amort_start_date,
			COALESCE(ci.useful_life_months, 0) as useful_life_months,
			COALESCE(ci.current_book_value, 0) as current_book_value,
			COALESCE(ci.purchase_cost - COALESCE(ci.current_book_value, 0), 0) as accumulated_depreciation,
			ci.created_at,
			ci.updated_at,
			ci.created_by,
			ci.updated_by,
			ctd.is_amortizable,
			'straight_line' as depreciation_method,
			ls.name as lifecycle_status_name,
			COALESCE(ls.amortization_behavior, 'pending') as amortization_behavior
		FROM configuration_items ci
		JOIN ci_type_definitions ctd ON ci.ci_type = ctd.name
		LEFT JOIN lifecycle_statuses ls ON ci.lifecycle_status_id = ls.id
		WHERE ci.id = $1 AND ctd.is_amortizable = true
	`

	var ci AmortizableCI
	var lifecycleStatusName sql.NullString
	var amortizationBehavior sql.NullString

	// Use nullable types for financial fields since they might be NULL
	var purchaseCost sql.NullFloat64
	var salvageValue sql.NullFloat64
	var amortStartDate sql.NullTime
	var usefulLifeMonths sql.NullInt32
	var currentBookValue sql.NullFloat64

	err := r.db.QueryRow(ctx, query, ciID).Scan(
		&ci.ID,
		&ci.Name,
		&ci.CIType,
		&ci.CITypeID,
		&ci.Attributes,
		&ci.Tags,
		&ci.LifecycleStatusID,
		&purchaseCost,
		&salvageValue,
		&amortStartDate,
		&usefulLifeMonths,
		&currentBookValue,
		&ci.AccumulatedDepreciation,
		&ci.CreatedAt,
		&ci.UpdatedAt,
		&ci.CreatedBy,
		&ci.UpdatedBy,
		&ci.IsAmortizable,
		&ci.DepreciationMethod,
		&lifecycleStatusName,
		&amortizationBehavior,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("amortizable CI not found: %w", err)
		}
		r.logger.ErrorService("amortization", "get_amortizable_ci", err, map[string]interface{}{
			"ci_id": ciID,
		})
		return nil, fmt.Errorf("failed to get amortizable CI: %w", err)
	}

	
	// Convert nullable financial fields to struct fields
	if purchaseCost.Valid {
		ci.PurchaseCost = purchaseCost.Float64
	} else {
		ci.PurchaseCost = 0.0
	}

	if salvageValue.Valid {
		ci.SalvageValue = salvageValue.Float64
	} else {
		ci.SalvageValue = 0.0
	}

	if amortStartDate.Valid {
		ci.AmortStartDate = &amortStartDate.Time
	}

	if usefulLifeMonths.Valid {
		ci.UsefulLifeMonths = int(usefulLifeMonths.Int32)
	} else {
		ci.UsefulLifeMonths = 0
	}

	if currentBookValue.Valid {
		ci.CurrentBookValue = currentBookValue.Float64
	} else {
		ci.CurrentBookValue = 0.0
	}

	// Set lifecycle status if available
	if lifecycleStatusName.Valid {
		ci.LifecycleStatus = &LifecycleStatus{
			Name:                 lifecycleStatusName.String,
			AmortizationBehavior: amortizationBehavior.String,
		}
	}

	// Default depreciation method if not set
	if ci.DepreciationMethod == "" {
		ci.DepreciationMethod = "straight_line"
	}

	return &ci, nil
}

// ListAmortizableCIs retrieves a paginated list of amortizable CIs
func (r *repository) ListAmortizableCIs(ctx context.Context, filters *AmortizableCIFilters) (*AmortizationCIList, error) {
	whereConditions := []string{"ctd.is_amortizable = true"}
	args := []interface{}{}
	argIndex := 1

	// Build WHERE clause
	if len(filters.CITypeIDs) > 0 {
		placeholders := make([]string, len(filters.CITypeIDs))
		for i := range filters.CITypeIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIndex+i)
		}
		whereConditions = append(whereConditions, fmt.Sprintf("ci.ci_type = ANY(ARRAY[%s])", strings.Join(placeholders, ",")))
		// Convert UUIDs to interface{}
	uuidArgs := make([]interface{}, len(filters.CITypeIDs))
	for i, id := range filters.CITypeIDs {
		uuidArgs[i] = id
	}
	args = append(args, uuidArgs...)
		argIndex += len(filters.CITypeIDs)
	}

	if len(filters.LifecycleStatusIDs) > 0 {
		placeholders := make([]string, len(filters.LifecycleStatusIDs))
		for i, _ := range filters.LifecycleStatusIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIndex+i)
		}
		whereConditions = append(whereConditions, fmt.Sprintf("ci.lifecycle_status_id = ANY(ARRAY[%s])", strings.Join(placeholders, ",")))
		// Convert UUIDs to interface{}
		uuidArgs2 := make([]interface{}, len(filters.LifecycleStatusIDs))
		for i, id := range filters.LifecycleStatusIDs {
			uuidArgs2[i] = id
		}
		args = append(args, uuidArgs2...)
		argIndex += len(filters.LifecycleStatusIDs)
	}

	if filters.IsAmortizable != nil && *filters.IsAmortizable {
		whereConditions = append(whereConditions, fmt.Sprintf("ctd.is_amortizable = $%d", argIndex))
		args = append(args, true)
		argIndex++
	}

	if filters.Search != nil && *filters.Search != "" {
		searchTerm := "%" + *filters.Search + "%"
		whereConditions = append(whereConditions, fmt.Sprintf("(ci.name ILIKE $%d OR ci.ci_type ILIKE $%d)", argIndex, argIndex+1))
		args = append(args, searchTerm, searchTerm)
		argIndex += 2
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + fmt.Sprintf("%s", whereConditions[0])
		for i := 1; i < len(whereConditions); i++ {
			whereClause += " AND " + whereConditions[i]
		}
	}

	// Build ORDER BY clause
	orderBy := "ci.name ASC"
	if filters.SortBy != nil {
		validSorts := map[string]string{
			"name":                    "ci.name",
			"ci_type":                 "ci.ci_type",
			"purchase_cost":           "ci.purchase_cost",
			"current_book_value":      "ci.current_book_value",
			"amort_start_date":        "ci.amort_start_date",
			"created_at":              "ci.created_at",
		}
		if sortField, ok := validSorts[*filters.SortBy]; ok {
			orderBy = sortField
			if filters.SortOrder != nil && *filters.SortOrder == "desc" {
				orderBy += " DESC"
			} else {
				orderBy += " ASC"
			}
		}
	}

	// Get total count
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM configuration_items ci
		JOIN ci_type_definitions ctd ON ci.ci_type = ctd.name
		LEFT JOIN lifecycle_statuses ls ON ci.lifecycle_status_id = ls.id
		%s
	`, whereClause)

	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.ErrorService("amortization", "list_amortizable_cis_count", err, map[string]interface{}{
			"filters": filters,
		})
		return nil, fmt.Errorf("failed to count amortizable CIs: %w", err)
	}

	// Pagination
	page := 1
	if filters.Page != nil {
		page = *filters.Page
	}
	pageSize := 20
	if filters.PageSize != nil {
		pageSize = *filters.PageSize
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	limit := pageSize
	offset := (page - 1) * limit

	// Main query
	query := fmt.Sprintf(`
		SELECT
			ci.id,
			ci.name,
			ci.ci_type,
			ctd.id as ci_type_id,
			ci.attributes,
			ci.tags,
			ci.lifecycle_status_id,
			COALESCE(ci.purchase_cost, 0) as purchase_cost,
			COALESCE(ci.salvage_value, 0) as salvage_value,
			ci.amort_start_date,
			COALESCE(ci.useful_life_months, 0) as useful_life_months,
			COALESCE(ci.current_book_value, 0) as current_book_value,
			COALESCE(ci.purchase_cost - COALESCE(ci.current_book_value, 0), 0) as accumulated_depreciation,
			ci.created_at,
			ci.updated_at,
			ci.created_by,
			ci.updated_by,
			ctd.is_amortizable,
			'straight_line' as depreciation_method,
			ls.name as lifecycle_status_name,
			COALESCE(ls.amortization_behavior, 'pending') as amortization_behavior
		FROM configuration_items ci
		JOIN ci_type_definitions ctd ON ci.ci_type = ctd.name
		LEFT JOIN lifecycle_statuses ls ON ci.lifecycle_status_id = ls.id
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.logger.ErrorService("amortization", "list_amortizable_cis", err, map[string]interface{}{
			"filters": filters,
		})
		return nil, fmt.Errorf("failed to list amortizable CIs: %w", err)
	}
	defer rows.Close()

	var cis []AmortizableCI
	for rows.Next() {
		var ci AmortizableCI
		var lifecycleStatusName sql.NullString
		var amortizationBehavior sql.NullString
		var amortStartDate sql.NullTime

		err := rows.Scan(
			&ci.ID,
			&ci.Name,
			&ci.CIType,
			&ci.CITypeID,
			&ci.Attributes,
			&ci.Tags,
			&ci.LifecycleStatusID,
			&ci.PurchaseCost,
			&ci.SalvageValue,
			&amortStartDate,
			&ci.UsefulLifeMonths,
			&ci.CurrentBookValue,
			&ci.AccumulatedDepreciation,
			&ci.CreatedAt,
			&ci.UpdatedAt,
			&ci.CreatedBy,
			&ci.UpdatedBy,
			&ci.IsAmortizable,
			&ci.DepreciationMethod,
			&lifecycleStatusName,
			&amortizationBehavior,
		)

		if err != nil {
			r.logger.ErrorService("amortization", "list_amortizable_cis_scan", err, nil)
			return nil, fmt.Errorf("failed to scan amortizable CI row: %w", err)
		}

		// Handle nullable amortization start date
		if amortStartDate.Valid {
			ci.AmortStartDate = &amortStartDate.Time
		}

		// Set lifecycle status if available
		if lifecycleStatusName.Valid {
			ci.LifecycleStatus = &LifecycleStatus{
				Name:                 lifecycleStatusName.String,
				AmortizationBehavior: amortizationBehavior.String,
			}
		}

		// Default depreciation method if not set
		if ci.DepreciationMethod == "" {
			ci.DepreciationMethod = "straight_line"
		}

		cis = append(cis, ci)
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &AmortizationCIList{
		CIs:         cis,
		TotalCount:  int(total),
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
	}, nil
}

// UpdateAmortizationConfig updates amortization configuration for a CI
func (r *repository) UpdateAmortizationConfig(ctx context.Context, ciID uuid.UUID, updates *AmortizationConfigUpdates) error {
	query := `
		UPDATE configuration_items
		SET
			purchase_cost = COALESCE($1, purchase_cost),
			salvage_value = COALESCE($2, salvage_value),
			amort_start_date = COALESCE($3, amort_start_date),
			useful_life_months = COALESCE($4, useful_life_months),
			current_book_value = COALESCE($5, current_book_value),
			updated_by = $6,
			updated_at = $7
		WHERE id = $8
	`

	_, err := r.db.Exec(ctx, query,
		updates.PurchaseCost,
		updates.SalvageValue,
		updates.AmortStartDate,
		updates.UsefulLifeMonths,
		updates.CurrentBookValue,
		updates.UpdatedBy,
		updates.UpdatedAt,
		ciID,
	)

	if err != nil {
		r.logger.ErrorService("amortization", "update_amortization_config", err, map[string]interface{}{
			"ci_id": ciID,
		})
		return fmt.Errorf("failed to update amortization configuration: %w", err)
	}

	return nil
}

// CreateLedgerEntry creates a new amortization ledger entry
func (r *repository) CreateLedgerEntry(ctx context.Context, entry *LedgerEntry) error {
	query := `
		INSERT INTO amortization_ledger (
			id,
			ci_id,
			entry_type,
			entry_date,
			amount,
			book_value_before,
			book_value_after,
			accumulated_depreciation,
			description,
			amortization_run_id,
			created_at,
			created_by,
			metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	// Prepare metadata
	metadata := map[string]interface{}{}
	if entry.Description != nil {
		metadata["description"] = *entry.Description
	}

	_, err := r.db.Exec(ctx, query,
		entry.ID,
		entry.CIID,
		entry.EntryType,
		entry.EntryDate,
		entry.Amount,
		entry.BookValueBefore,
		entry.BookValueAfter,
		entry.AccumulatedDepreciation,
		entry.Description,
		entry.AmortizationRunID,
		entry.CreatedAt,
		entry.CreatedBy,
		metadata,
	)

	if err != nil {
		r.logger.ErrorService("amortization", "create_ledger_entry", err, map[string]interface{}{
			"ci_id":    entry.CIID,
			"entry_id": entry.ID,
		})
		return fmt.Errorf("failed to create ledger entry: %w", err)
	}

	return nil
}

// GetLedgerEntries retrieves ledger entries with filtering
func (r *repository) GetLedgerEntries(ctx context.Context, filters *LedgerFilters) (*LedgerEntryList, error) {
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	// Build WHERE clause
	if filters.CIID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("ci_id = $%d", argIndex))
		args = append(args, *filters.CIID)
		argIndex++
	}

	if len(filters.EntryTypes) > 0 {
		placeholders := make([]string, len(filters.EntryTypes))
		for i, entryType := range filters.EntryTypes {
			placeholders[i] = fmt.Sprintf("$%d", argIndex+i)
			args = append(args, entryType)
		}
		whereConditions = append(whereConditions, fmt.Sprintf("entry_type = ANY(ARRAY[%s])", strings.Join(placeholders, ",")))
		argIndex += len(filters.EntryTypes)
	}

	if filters.DateFrom != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("entry_date >= $%d", argIndex))
		args = append(args, *filters.DateFrom)
		argIndex++
	}

	if filters.DateTo != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("entry_date <= $%d", argIndex))
		args = append(args, *filters.DateTo)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + fmt.Sprintf("%s", whereConditions[0])
		for i := 1; i < len(whereConditions); i++ {
			whereClause += " AND " + whereConditions[i]
		}
	}

	// Build ORDER BY clause
	orderBy := "entry_date DESC, created_at DESC"
	if filters.SortBy != nil {
		validSorts := map[string]string{
			"entry_date": "entry_date",
			"amount":     "amount",
			"created_at": "created_at",
					}
		if sortField, ok := validSorts[*filters.SortBy]; ok {
			orderBy = sortField
			if filters.SortOrder != nil && *filters.SortOrder == "asc" {
				orderBy += " ASC"
			} else {
				orderBy += " DESC"
			}
		}
	}

	// Get total count
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM amortization_ledger
		%s
	`, whereClause)

	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.ErrorService("amortization", "get_ledger_entries_count", err, map[string]interface{}{
			"filters": filters,
		})
		return nil, fmt.Errorf("failed to count ledger entries: %w", err)
	}

	// Pagination
	page := 1
	if filters.Page != nil {
		page = *filters.Page
	}
	pageSize := 20
	if filters.PageSize != nil {
		pageSize = *filters.PageSize
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	limit := pageSize
	offset := (page - 1) * limit

	// Main query
	query := fmt.Sprintf(`
		SELECT
			id,
			ci_id,
			entry_type,
			entry_date,
			amount,
			book_value_before,
			book_value_after,
			accumulated_depreciation,
			description,
			amortization_run_id,
			created_at,
			created_by
		FROM amortization_ledger
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.logger.ErrorService("amortization", "get_ledger_entries", err, map[string]interface{}{
			"filters": filters,
		})
		return nil, fmt.Errorf("failed to get ledger entries: %w", err)
	}
	defer rows.Close()

	var entries []LedgerEntry
	for rows.Next() {
		var entry LedgerEntry
		var description sql.NullString
		var runID sql.NullString

		err := rows.Scan(
			&entry.ID,
			&entry.CIID,
			&entry.EntryType,
			&entry.EntryDate,
			&entry.Amount,
			&entry.BookValueBefore,
			&entry.BookValueAfter,
			&entry.AccumulatedDepreciation,
			&description,
			&runID,
			&entry.CreatedAt,
			&entry.CreatedBy,
		)

		if err != nil {
			r.logger.ErrorService("amortization", "get_ledger_entries_scan", err, nil)
			return nil, fmt.Errorf("failed to scan ledger entry row: %w", err)
		}

		if description.Valid {
			entry.Description = &description.String
		}
		if runID.Valid {
			// Parse UUID from string
			if runUUID, err := uuid.Parse(runID.String); err == nil {
				entry.AmortizationRunID = &runUUID
			}
		}

		entries = append(entries, entry)
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &LedgerEntryList{
		Entries:    entries,
		TotalCount: int(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetLedgerEntry retrieves a specific ledger entry by ID
func (r *repository) GetLedgerEntry(ctx context.Context, entryID uuid.UUID) (*LedgerEntry, error) {
	query := `
		SELECT
			id,
			ci_id,
			entry_type,
			entry_date,
			amount,
			book_value_before,
			book_value_after,
			accumulated_depreciation,
			description,
			amortization_run_id,
			created_at,
			created_by,
			metadata
		FROM amortization_ledger
		WHERE id = $1
	`

	var entry LedgerEntry
	var description sql.NullString
	var amortizationRunID sql.NullString
	var metadata sql.NullString

	err := r.db.QueryRow(ctx, query, entryID).Scan(
		&entry.ID,
		&entry.CIID,
		&entry.EntryType,
		&entry.EntryDate,
		&entry.Amount,
		&entry.BookValueBefore,
		&entry.BookValueAfter,
		&entry.AccumulatedDepreciation,
		&description,
		&amortizationRunID,
		&entry.CreatedAt,
		&entry.CreatedBy,
		&metadata,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("ledger entry not found: %w", err)
		}
		r.logger.ErrorService("amortization", "get_ledger_entry", err, map[string]interface{}{
			"entry_id": entryID,
		})
		return nil, fmt.Errorf("failed to get ledger entry: %w", err)
	}

	if description.Valid {
		entry.Description = &description.String
	}
	if amortizationRunID.Valid {
		// Parse UUID from string
		if runUUID, err := uuid.Parse(amortizationRunID.String); err == nil {
			entry.AmortizationRunID = &runUUID
		}
	}

	return &entry, nil
}

// GetCILatestLedgerEntry retrieves the latest ledger entry for a CI
func (r *repository) GetCILatestLedgerEntry(ctx context.Context, ciID uuid.UUID) (*LedgerEntry, error) {
	query := `
		SELECT
			id,
			ci_id,
			entry_type,
			entry_date,
			amount,
			book_value_before,
			book_value_after,
			accumulated_depreciation,
			description,
			amortization_run_id,
			created_at,
			created_by
		FROM amortization_ledger
		WHERE ci_id = $1
		ORDER BY entry_date DESC, created_at DESC
		LIMIT 1
	`

	var entry LedgerEntry
	var description sql.NullString
	var amortizationRunID sql.NullString

	err := r.db.QueryRow(ctx, query, ciID).Scan(
		&entry.ID,
		&entry.CIID,
		&entry.EntryType,
		&entry.EntryDate,
		&entry.Amount,
		&entry.BookValueBefore,
		&entry.BookValueAfter,
		&entry.AccumulatedDepreciation,
		&description,
		&amortizationRunID,
		&entry.CreatedAt,
		&entry.CreatedBy,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No entries found is not an error
		}
		r.logger.ErrorService("amortization", "get_ci_latest_ledger_entry", err, map[string]interface{}{
			"ci_id": ciID,
		})
		return nil, fmt.Errorf("failed to get latest ledger entry: %w", err)
	}

	if description.Valid {
		entry.Description = &description.String
	}
	if amortizationRunID.Valid {
		// Parse UUID from string
		if runUUID, err := uuid.Parse(amortizationRunID.String); err == nil {
			entry.AmortizationRunID = &runUUID
		}
	}

	return &entry, nil
}

// CreateAmortizationRun creates a new amortization run
func (r *repository) CreateAmortizationRun(ctx context.Context, run *AmortizationRun) error {
	query := `
		INSERT INTO amortization_runs (
			id,
			status,
			run_date,
			started_at,
			completed_at,
			total_cis_processed,
			successful_depreciations,
			write_offs_generated,
			errors_encountered,
			checkpoint_data,
			error_summary,
			error_details,
			run_config,
			created_at,
			created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := r.db.Exec(ctx, query,
		run.ID,
		run.Status,
		run.ProcessingDate,
		run.StartedAt,
		run.CompletedAt,
		run.ProcessedCIs,
		run.ProcessedCIs,
		run.SkippedCIs,
		run.FailedCIs,
		"{}", // checkpoint_data
		run.ErrorSummary,
		"{}", // error_details
		"{}", // run_config
		run.CreatedAt,
		run.TriggeredBy,
	)

	if err != nil {
		r.logger.ErrorService("amortization", "create_amortization_run", err, map[string]interface{}{
			"run_id": run.ID,
		})
		return fmt.Errorf("failed to create amortization run: %w", err)
	}

	return nil
}

// UpdateAmortizationRun updates an amortization run
func (r *repository) UpdateAmortizationRun(ctx context.Context, runID uuid.UUID, updates *AmortizationRunUpdates) error {
	query := `
		UPDATE amortization_runs
		SET
			status = COALESCE($1, status),
			started_at = COALESCE($2, started_at),
			completed_at = COALESCE($3, completed_at),
			total_cis_processed = COALESCE($4, total_cis_processed),
			successful_depreciations = COALESCE($5, successful_depreciations),
			write_offs_generated = COALESCE($6, write_offs_generated),
			errors_encountered = COALESCE($7, errors_encountered),
			last_processed_ci_id = COALESCE($8, last_processed_ci_id),
			last_processed_at = COALESCE($9, last_processed_at),
			error_summary = COALESCE($10, error_summary)
		WHERE id = $11
	`

	_, err := r.db.Exec(ctx, query,
		updates.Status,
		updates.StartedAt,
		updates.CompletedAt,
		updates.ProcessedCIs,
		updates.ProcessedCIs,
		updates.SkippedCIs,
		updates.FailedCIs,
		nil, // last_processed_ci_id
		nil, // last_processed_at
		updates.ErrorSummary,
		runID,
	)

	if err != nil {
		r.logger.ErrorService("amortization", "update_amortization_run", err, map[string]interface{}{
			"run_id": runID,
		})
		return fmt.Errorf("failed to update amortization run: %w", err)
	}

	return nil
}

// GetAmortizationRun retrieves a specific amortization run
func (r *repository) GetAmortizationRun(ctx context.Context, runID uuid.UUID) (*AmortizationRun, error) {
	query := `
		SELECT
			id,
			status,
			run_date,
			started_at,
			completed_at,
			total_cis_processed,
			successful_depreciations,
			write_offs_generated,
			errors_encountered,
			last_processed_ci_id,
			last_processed_at,
			checkpoint_data,
			error_summary,
			error_details,
			run_config,
			created_at,
			created_by
		FROM amortization_runs
		WHERE id = $1
	`

	var run AmortizationRun
	var totalCisProcessed sql.NullInt32
	var successfulDepreciations sql.NullInt32
	var writeOffsGenerated sql.NullInt32
	var errorsEncountered sql.NullInt32
	var lastProcessedCIID sql.NullString
	var lastProcessedAt sql.NullTime
	var checkpointData sql.NullString
	var errorSummary sql.NullString
	var errorDetails sql.NullString
	var runConfig sql.NullString
	var triggeredBy sql.NullString

	err := r.db.QueryRow(ctx, query, runID).Scan(
		&run.ID,
		&run.Status,
		&run.ProcessingDate,
		&run.StartedAt,
		&run.CompletedAt,
		&totalCisProcessed,
		&successfulDepreciations,
		&writeOffsGenerated,
		&errorsEncountered,
		&lastProcessedCIID,
		&lastProcessedAt,
		&checkpointData,
		&errorSummary,
		&errorDetails,
		&runConfig,
		&run.CreatedAt,
		&triggeredBy,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("amortization run not found: %w", err)
		}
		r.logger.ErrorService("amortization", "get_amortization_run", err, map[string]interface{}{
			"run_id": runID,
		})
		return nil, fmt.Errorf("failed to get amortization run: %w", err)
	}

	if totalCisProcessed.Valid {
		val := int(totalCisProcessed.Int32)
		run.ProcessedCIs = &val
	}
	if errorSummary.Valid {
		run.ErrorSummary = &errorSummary.String
	}
	if triggeredBy.Valid {
		// Parse UUID from string
		if triggerUUID, err := uuid.Parse(triggeredBy.String); err == nil {
			run.TriggeredBy = &triggerUUID
		}
	}

	return &run, nil
}

// ListAmortizationRuns retrieves amortization runs with filtering
func (r *repository) ListAmortizationRuns(ctx context.Context, filters *AmortizationRunFilters) (*AmortizationRunList, error) {
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	// Build WHERE clause
	if len(filters.Status) > 0 {
		placeholders := make([]string, len(filters.Status))
		for i, status := range filters.Status {
			placeholders[i] = fmt.Sprintf("$%d", argIndex+i)
			args = append(args, status)
		}
		whereConditions = append(whereConditions, fmt.Sprintf("status = ANY(ARRAY[%s])", strings.Join(placeholders, ",")))
		argIndex += len(filters.Status)
	}

	if filters.DateFrom != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("run_date >= $%d", argIndex))
		args = append(args, *filters.DateFrom)
		argIndex++
	}

	if filters.DateTo != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("run_date <= $%d", argIndex))
		args = append(args, *filters.DateTo)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + fmt.Sprintf("%s", whereConditions[0])
		for i := 1; i < len(whereConditions); i++ {
			whereClause += " AND " + whereConditions[i]
		}
	}

	// Get total count
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM amortization_runs
		%s
	`, whereClause)

	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.ErrorService("amortization", "list_amortization_runs_count", err, map[string]interface{}{
			"filters": filters,
		})
		return nil, fmt.Errorf("failed to count amortization runs: %w", err)
	}

	// Pagination
	page := 1
	if filters.Page != nil {
		page = *filters.Page
	}
	pageSize := 20
	if filters.PageSize != nil {
		pageSize = *filters.PageSize
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	limit := pageSize
	offset := (page - 1) * limit

	// Main query
	query := fmt.Sprintf(`
		SELECT
			id,
			status,
			run_date,
			started_at,
			completed_at,
			total_cis_processed,
			successful_depreciations,
			write_offs_generated,
			errors_encountered,
			last_processed_ci_id,
			last_processed_at,
			checkpoint_data,
			error_summary,
			error_details,
			run_config,
			created_at,
			created_by
		FROM amortization_runs
		%s
		ORDER BY run_date DESC, created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.logger.ErrorService("amortization", "list_amortization_runs", err, map[string]interface{}{
			"filters": filters,
		})
		return nil, fmt.Errorf("failed to list amortization runs: %w", err)
	}
	defer rows.Close()

	var runs []AmortizationRun
	for rows.Next() {
		var run AmortizationRun
		var totalCisProcessed sql.NullInt32
		var successfulDepreciations sql.NullInt32
		var writeOffsGenerated sql.NullInt32
		var errorsEncountered sql.NullInt32
		var lastProcessedCIID sql.NullString
		var lastProcessedAt sql.NullTime
		var checkpointData sql.NullString
		var errorSummary sql.NullString
		var errorDetails sql.NullString
		var runConfig sql.NullString
		var triggeredBy sql.NullString

		err := rows.Scan(
			&run.ID,
			&run.Status,
			&run.ProcessingDate,
			&run.StartedAt,
			&run.CompletedAt,
			&totalCisProcessed,
			&successfulDepreciations,
			&writeOffsGenerated,
			&errorsEncountered,
			&lastProcessedCIID,
			&lastProcessedAt,
			&checkpointData,
			&errorSummary,
			&errorDetails,
			&runConfig,
			&run.CreatedAt,
			&triggeredBy,
		)

		if err != nil {
			r.logger.ErrorService("amortization", "list_amortization_runs_scan", err, nil)
			return nil, fmt.Errorf("failed to scan amortization run row: %w", err)
		}

		if totalCisProcessed.Valid {
			val := int(totalCisProcessed.Int32)
			run.ProcessedCIs = &val
		}
		if errorSummary.Valid {
			run.ErrorSummary = &errorSummary.String
		}
		if triggeredBy.Valid {
			// Parse UUID from string
			if triggerUUID, err := uuid.Parse(triggeredBy.String); err == nil {
				run.TriggeredBy = &triggerUUID
			}
		}

		runs = append(runs, run)
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &AmortizationRunList{
		Runs:       runs,
		TotalCount: int(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetAmortizationSummaries retrieves amortization summaries for reporting
func (r *repository) GetAmortizationSummaries(ctx context.Context, req *SummaryRequest) (*AmortizationSummary, error) {
	query := `
		SELECT
			COUNT(ci.id) as total_cis,
			COALESCE(SUM(ci.current_book_value), 0) as total_book_value,
			COALESCE(SUM(CASE
				WHEN ci.purchase_cost > 0 AND ci.current_book_value IS NOT NULL
				THEN ci.purchase_cost - COALESCE(ci.current_book_value, 0)
				ELSE 0
			END), 0) as total_depreciation
		FROM configuration_items ci
		JOIN ci_type_definitions ctd ON ci.ci_type = ctd.name
		LEFT JOIN lifecycle_statuses ls ON ci.lifecycle_status_id = ls.id
		WHERE ctd.is_amortizable = true
		AND ci.amort_start_date IS NOT NULL
	`

	var totalCIs int64
	var totalBookValue float64
	var totalDepreciation float64

	err := r.db.QueryRow(ctx, query).Scan(&totalCIs, &totalBookValue, &totalDepreciation)
	if err != nil {
		r.logger.ErrorService("amortization", "get_amortization_summaries", err, map[string]interface{}{
			"request": req,
		})
		return nil, fmt.Errorf("failed to get amortization summaries: %w", err)
	}

	return &AmortizationSummary{
		GroupBy:            "all",
		Groups:             []AmortizationGroup{},
		TotalCIs:           int(totalCIs),
		TotalBookValue:     totalBookValue,
		TotalDepreciation:  totalDepreciation,
		GeneratedAt:        time.Now(),
	}, nil
}

// GetDepreciationScheduleData retrieves data for depreciation schedule generation
func (r *repository) GetDepreciationScheduleData(ctx context.Context, req *DepreciationScheduleRequest) ([]DepreciationScheduleEntry, error) {
	// This would involve complex queries to generate depreciation schedule data
	// For now, return an empty slice as a placeholder
	return []DepreciationScheduleEntry{}, nil
}

// GetCIsForProcessing retrieves CIs that need amortization processing
func (r *repository) GetCIsForProcessing(ctx context.Context, processingDate time.Time, limit int) ([]uuid.UUID, error) {
	query := `
		SELECT DISTINCT ci.id
		FROM configuration_items ci
		JOIN ci_type_definitions ctd ON ci.ci_type = ctd.name
		LEFT JOIN lifecycle_statuses ls ON ci.lifecycle_status_id = ls.id
		WHERE ctd.is_amortizable = true
		AND ci.amort_start_date IS NOT NULL
		AND ci.useful_life_months IS NOT NULL
		AND ci.current_book_value > 0
		AND (ls.amortization_behavior = 'active' OR ls.amortization_behavior IS NULL)
		AND NOT EXISTS (
			SELECT 1 FROM amortization_ledger al
			WHERE al.ci_id = ci.id
			AND al.entry_date = $1
			AND al.entry_type = 'depreciation'
		)
		ORDER BY ci.amort_start_date, ci.id
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, processingDate, limit)
	if err != nil {
		r.logger.ErrorService("amortization", "get_cis_for_processing", err, map[string]interface{}{
			"processing_date": processingDate,
		})
		return nil, fmt.Errorf("failed to get CIs for processing: %w", err)
	}
	defer rows.Close()

	var ciIDs []uuid.UUID
	for rows.Next() {
		var ciID uuid.UUID
		if err := rows.Scan(&ciID); err != nil {
			r.logger.ErrorService("amortization", "get_cis_for_processing_scan", err, nil)
			return nil, fmt.Errorf("failed to scan CI ID: %w", err)
		}
		ciIDs = append(ciIDs, ciID)
	}

	return ciIDs, nil
}

// MarkCIsAsProcessed marks CIs as processed in a run (this would need implementation)
func (r *repository) MarkCIsAsProcessed(ctx context.Context, runID uuid.UUID, processedCIs []ProcessingResult) error {
	// This would typically update run statistics or create processing records
	// For now, return nil as a placeholder
	return nil
}

// WithTransaction executes a function within a database transaction
func (r *repository) WithTransaction(ctx context.Context, fn func(context.Context, interface{}) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := fn(ctx, tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// parseJSON is a helper function to parse JSON bytes into an interface
func parseJSON(data []byte, v interface{}) error {
	if len(data) == 0 {
		return nil
	}
	// In a real implementation, you would use json.Unmarshal here
	// For now, just return nil as a placeholder
	return nil
}