package amortization

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
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
			COALESCE(
				(SELECT COALESCE(SUM(amount), 0)
				 FROM amortization_ledger ale
				 WHERE ale.ci_id = ci.id AND ale.entry_type IN ('monthly_depreciation', 'catch_up_depreciation', 'depreciation')
				),
				0
			) as accumulated_depreciation,
			CASE
				WHEN COALESCE(ci.useful_life_months, 0) > 0
				THEN ROUND((COALESCE(ci.purchase_cost, 0) - COALESCE(ci.salvage_value, 0)) / NULLIF(ci.useful_life_months, 0)::numeric, 2)
				ELSE NULL
			END as monthly_depreciation,
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
	var monthlyDepreciation sql.NullFloat64

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
		&monthlyDepreciation,
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

	if monthlyDepreciation.Valid {
		ci.MonthlyDepreciation = &monthlyDepreciation.Float64
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
			"name":               "ci.name",
			"ci_type":            "ci.ci_type",
			"purchase_cost":      "ci.purchase_cost",
			"current_book_value": "ci.current_book_value",
			"amort_start_date":   "ci.amort_start_date",
			"created_at":         "ci.created_at",
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
			ci.id as ci_id,
			ci.name as ci_name,
			ci.ci_type as ci_type_name,
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
			CASE
				WHEN COALESCE(ci.useful_life_months, 0) > 0
				THEN ROUND((COALESCE(ci.purchase_cost, 0) - COALESCE(ci.salvage_value, 0)) / NULLIF(ci.useful_life_months, 0)::numeric, 2)
				ELSE NULL
			END as monthly_depreciation,
			CASE
				WHEN ci.amort_start_date IS NOT NULL AND ci.useful_life_months IS NOT NULL AND ci.useful_life_months > 0
				THEN GREATEST(0, ci.useful_life_months - EXTRACT(YEAR FROM AGE(CURRENT_DATE, ci.amort_start_date)) * 12 - EXTRACT(MONTH FROM AGE(CURRENT_DATE, ci.amort_start_date)))
				ELSE NULL
			END as remaining_months,
			ci.created_at,
			ci.updated_at,
			ci.created_by,
			ci.updated_by,
			ctd.is_amortizable,
			'straight_line' as depreciation_method,
			ls.name as lifecycle_status_name,
			COALESCE(ls.amortization_behavior, 'pending') as status,
			ci.amort_start_date
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
		var amortStartDate sql.NullTime
		var monthlyDepreciation sql.NullFloat64
		var remainingMonths sql.NullInt64
		var status sql.NullString

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
			&monthlyDepreciation,
			&remainingMonths,
			&ci.CreatedAt,
			&ci.UpdatedAt,
			&ci.CreatedBy,
			&ci.UpdatedBy,
			&ci.IsAmortizable,
			&ci.DepreciationMethod,
			&lifecycleStatusName,
			&status,
			&amortStartDate, // Scan amort_start_date again for the AmortizableCI struct field
		)

		if err != nil {
			r.logger.ErrorService("amortization", "list_amortizable_cis_scan", err, nil)
			return nil, fmt.Errorf("failed to scan amortizable CI row: %w", err)
		}

		// Handle nullable amortization start date
		if amortStartDate.Valid {
			ci.AmortStartDate = &amortStartDate.Time
		}

		// Handle nullable monthly depreciation
		if monthlyDepreciation.Valid {
			ci.MonthlyDepreciation = &monthlyDepreciation.Float64
		}

		// Set computed remaining months field
		if remainingMonths.Valid {
			ci.RemainingMonths = &remainingMonths.Int64
		}

		// Set status field (alias for amortization_behavior)
		if status.Valid {
			statusStr := status.String
			ci.Status = &statusStr
			ci.AmortizationBehavior = statusStr
		} else {
			defaultStatus := "pending"
			ci.Status = &defaultStatus
			ci.AmortizationBehavior = "pending"
		}

		// Populate frontend-specific aliases for backward compatibility
		ciIDStr := ci.ID.String()
		ci.CIID = &ciIDStr
		ci.CIName = &ci.Name
		ci.CITypeName = &ci.CIType

		// Set lifecycle status if available
		if lifecycleStatusName.Valid {
			ci.LifecycleStatus = &LifecycleStatus{
				Name:                 lifecycleStatusName.String,
				AmortizationBehavior: ci.AmortizationBehavior,
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
		CIs:        cis,
		TotalCount: int(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
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
			created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

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

// HasLedgerEntriesForCI checks if a CI has any existing ledger entries
func (r *repository) HasLedgerEntriesForCI(ctx context.Context, ciID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM amortization_ledger WHERE ci_id = $1 LIMIT 1)`

	var exists bool
	err := r.db.QueryRow(ctx, query, ciID).Scan(&exists)
	if err != nil {
		r.logger.Error().
			Err(err).
			Str("ci_id", ciID.String()).
			Msg("Failed to check if CI has ledger entries")
		return false, fmt.Errorf("failed to check if CI has ledger entries: %w", err)
	}

	r.logger.Info().
		Str("ci_id", ciID.String()).
		Bool("has_entries", exists).
		Msg("Checked for existing ledger entries")

	return exists, nil
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

	if filters.CITypeID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("ctd.id = $%d", argIndex))
		args = append(args, *filters.CITypeID)
		argIndex++
	}

	if filters.CINameSearch != nil && *filters.CINameSearch != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("ci.name ILIKE $%d", argIndex))
		args = append(args, "%"+*filters.CINameSearch+"%")
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
			ale.id,
			ale.ci_id,
			ci.name as ci_name,
			ale.entry_type,
			ale.entry_date,
			ale.amount,
			ale.book_value_before,
			ale.book_value_after,
			ale.accumulated_depreciation,
			ale.description,
			ale.amortization_run_id,
			ale.created_at,
			ale.created_by,
			u.username as created_by_name
		FROM amortization_ledger ale
		JOIN configuration_items ci ON ale.ci_id = ci.id
		LEFT JOIN ci_type_definitions ctd ON ci.ci_type = ctd.name
		LEFT JOIN users u ON ale.created_by = u.id
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
		var createdBy sql.NullString
		var createdByName sql.NullString

		err := rows.Scan(
			&entry.ID,
			&entry.CIID,
			&entry.CIName,
			&entry.EntryType,
			&entry.EntryDate,
			&entry.Amount,
			&entry.BookValueBefore,
			&entry.BookValueAfter,
			&entry.AccumulatedDepreciation,
			&description,
			&runID,
			&entry.CreatedAt,
			&createdBy,
			&createdByName,
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
		if createdBy.Valid {
			// Parse UUID from string
			if createdByUUID, err := uuid.Parse(createdBy.String); err == nil {
				entry.CreatedBy = &createdByUUID
			}
		}
		if createdByName.Valid {
			entry.CreatedByName = &createdByName.String
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
			created_by
		FROM amortization_ledger
		WHERE id = $1
	`

	var entry LedgerEntry
	var description sql.NullString
	var amortizationRunID sql.NullString

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
			is_manual,
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
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err := r.db.Exec(ctx, query,
		run.ID,
		run.Status,
		run.ProcessingDate,
		run.IsManual,
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
			is_manual,
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
		&run.IsManual,
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
			is_manual,
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
			&run.IsManual,
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
			END), 0) as total_depreciation,
			COALESCE(SUM(
				CASE
					WHEN COALESCE(ci.useful_life_months, 0) > 0
					THEN ROUND((COALESCE(ci.purchase_cost, 0) - COALESCE(ci.salvage_value, 0)) / NULLIF(ci.useful_life_months, 0)::numeric, 2)
					ELSE 0
				END
			), 0) as total_monthly_depreciation
		FROM configuration_items ci
		JOIN ci_type_definitions ctd ON ci.ci_type = ctd.name
		LEFT JOIN lifecycle_statuses ls ON ci.lifecycle_status_id = ls.id
		WHERE ctd.is_amortizable = true
		AND ci.amort_start_date IS NOT NULL
	`

	var totalCIs int64
	var totalBookValue float64
	var totalDepreciation float64
	var totalMonthlyDepreciation float64

	err := r.db.QueryRow(ctx, query).Scan(&totalCIs, &totalBookValue, &totalDepreciation, &totalMonthlyDepreciation)
	if err != nil {
		r.logger.ErrorService("amortization", "get_amortization_summaries", err, map[string]interface{}{
			"request": req,
		})
		return nil, fmt.Errorf("failed to get amortization summaries: %w", err)
	}

	return &AmortizationSummary{
		GroupBy:                  "all",
		Groups:                   []AmortizationGroup{},
		TotalCIs:                 int(totalCIs),
		TotalBookValue:           totalBookValue,
		TotalDepreciation:        totalDepreciation,
		TotalMonthlyDepreciation: totalMonthlyDepreciation,
		GeneratedAt:              time.Now(),
	}, nil
}

// GetDepreciationScheduleData retrieves data for depreciation schedule generation
func (r *repository) GetDepreciationScheduleData(ctx context.Context, req *DepreciationScheduleRequest) (*DepreciationScheduleResponse, error) {
	// Get amortization settings for currency
	// If settings don't exist yet, use default currency
	currency := "USD"
	settings, err := r.GetAmortizationSettings(ctx)
	if err == nil && settings != nil {
		currency = settings.Currency
	}
	// If settings table doesn't exist, we continue with default currency

	response := &DepreciationScheduleResponse{
		Currency:  currency,
		StartDate: req.DateFrom,
		EndDate:   req.DateTo,
		MonthlyData: []MonthlyScheduleEntry{},
	}

	// Build WHERE clause for filters
	whereClause := "WHERE ale.entry_date >= $1 AND ale.entry_date <= $2"
	args := []interface{}{req.DateFrom, req.DateTo}
	argIndex := 3

	// Add CI type filter if specified
	if len(req.CITypeIDs) > 0 {
		placeholders := make([]string, len(req.CITypeIDs))
		for i, id := range req.CITypeIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, id)
			argIndex++
		}
		whereClause += fmt.Sprintf(" AND ctd.id IN (%s)", strings.Join(placeholders, ", "))
	}

	// Add CI filter if specified
	if len(req.CIIDs) > 0 {
		placeholders := make([]string, len(req.CIIDs))
		for i, id := range req.CIIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, id)
			argIndex++
		}
		whereClause += fmt.Sprintf(" AND ale.ci_id IN (%s)", strings.Join(placeholders, ", "))
	}

	// Query 1: Get historical monthly data from ledger
	// For each month, get opening value (first entry) and closing value (last entry)
	historicalQuery := `
		WITH monthly_opening AS (
			SELECT DISTINCT ON (date_trunc('month', ale.entry_date), ale.ci_id)
				date_trunc('month', ale.entry_date) as month,
				ale.ci_id,
				ale.book_value_before as opening_value
			FROM amortization_ledger ale
		` + whereClause + `
			ORDER BY date_trunc('month', ale.entry_date), ale.ci_id, ale.entry_date ASC, ale.created_at ASC
		),
		monthly_closing AS (
			SELECT DISTINCT ON (date_trunc('month', ale.entry_date), ale.ci_id)
				date_trunc('month', ale.entry_date) as month,
				ale.ci_id,
				ale.book_value_after as closing_value
			FROM amortization_ledger ale
		` + whereClause + `
			ORDER BY date_trunc('month', ale.entry_date), ale.ci_id, ale.entry_date DESC, ale.created_at DESC
		),
		monthly_amounts AS (
			SELECT
				date_trunc('month', ale.entry_date) as month,
				SUM(CASE WHEN ale.entry_type IN ('depreciation', 'monthly_depreciation', 'catch_up_depreciation') THEN ale.amount ELSE 0 END) as depreciation_amount,
				SUM(CASE WHEN ale.entry_type = 'write_off' THEN ale.amount ELSE 0 END) as write_off_amount,
				SUM(CASE WHEN ale.entry_type = 'adjustment' THEN ale.amount ELSE 0 END) as adjustment_amount,
				COUNT(DISTINCT ale.ci_id) as active_assets_count
			FROM amortization_ledger ale
			JOIN configuration_items ci ON ale.ci_id = ci.id
			LEFT JOIN ci_type_definitions ctd ON ci.ci_type = ctd.name
		` + whereClause + `
			GROUP BY date_trunc('month', ale.entry_date)
		)
		SELECT
			mc.month,
			COALESCE(ma.depreciation_amount, 0) as depreciation_amount,
			COALESCE(ma.write_off_amount, 0) as write_off_amount,
			COALESCE(ma.adjustment_amount, 0) as adjustment_amount,
			SUM(mo.opening_value) as opening_book_value,
			SUM(mc.closing_value) as closing_book_value,
			COALESCE(ma.active_assets_count, 0) as active_assets_count
		FROM monthly_closing mc
		JOIN monthly_opening mo ON mc.month = mo.month AND mc.ci_id = mo.ci_id
		LEFT JOIN monthly_amounts ma ON mc.month = ma.month
		GROUP BY mc.month, ma.depreciation_amount, ma.write_off_amount, ma.adjustment_amount, ma.active_assets_count
		ORDER BY mc.month
	`

	rows, err := r.db.Query(ctx, historicalQuery, args...)
	if err != nil {
		r.logger.Error().Err(err).Str("query", "historical_data").Msg("Failed to query historical data")
		return nil, fmt.Errorf("failed to query historical data: %w", err)
	}
	defer rows.Close()

	historicalData := map[string]MonthlyScheduleEntry{}
	for rows.Next() {
		var entry MonthlyScheduleEntry
		var openingBookValue, closingBookValue sql.NullFloat64

		err := rows.Scan(
			&entry.Month,
			&entry.DepreciationAmount,
			&entry.WriteOffAmount,
			&entry.AdjustmentAmount,
			&openingBookValue,
			&closingBookValue,
			&entry.ActiveAssetsCount,
		)
		if err != nil {
			r.logger.Error().Err(err).Msg("Failed to scan historical row")
			return nil, fmt.Errorf("failed to scan historical row: %w", err)
		}

		entry.IsProjected = false
		entry.OpeningBookValue = openingBookValue.Float64
		entry.ClosingBookValue = closingBookValue.Float64
		historicalData[entry.Month.Format("2006-01")] = entry
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating historical rows: %w", err)
	}

	// Find the last month with actual ledger entries
	lastHistoricalMonth := time.Time{}
	for monthStr := range historicalData {
		entry := historicalData[monthStr]
		if entry.Month.After(lastHistoricalMonth) {
			lastHistoricalMonth = entry.Month
		}
	}

	// Query 2: Get projected data from active assets
	// Get all active amortizable assets
	assetWhereClause := "WHERE ci.current_book_value > COALESCE(ci.salvage_value, 0)"
	assetArgs := []interface{}{}
	assetArgIndex := 1

	if len(req.CITypeIDs) > 0 {
		placeholders := make([]string, len(req.CITypeIDs))
		for i, id := range req.CITypeIDs {
			placeholders[i] = fmt.Sprintf("$%d", assetArgIndex)
			assetArgs = append(assetArgs, id)
			assetArgIndex++
		}
		assetWhereClause += fmt.Sprintf(" AND ctd.id IN (%s)", strings.Join(placeholders, ", "))
	}

	if len(req.CIIDs) > 0 {
		placeholders := make([]string, len(req.CIIDs))
		for i, id := range req.CIIDs {
			placeholders[i] = fmt.Sprintf("$%d", assetArgIndex)
			assetArgs = append(assetArgs, id)
			assetArgIndex++
		}
		assetWhereClause += fmt.Sprintf(" AND ci.id IN (%s)", strings.Join(placeholders, ", "))
	}

	// Query for active assets to project
	assetQuery := `
		SELECT
			ci.id,
			ci.name,
			ci.current_book_value,
			COALESCE(ci.salvage_value, 0) as salvage_value,
			COALESCE(ci.purchase_cost, 0) as purchase_cost,
			ci.useful_life_months,
			ci.amort_start_date,
			ctd.id as ci_type_id,
			ctd.name as ci_type_name
		FROM configuration_items ci
		JOIN ci_type_definitions ctd ON ci.ci_type = ctd.name
		LEFT JOIN lifecycle_statuses ls ON ci.lifecycle_status_id = ls.id
		` + assetWhereClause + `
		AND ci.amort_start_date IS NOT NULL
		AND ci.useful_life_months IS NOT NULL
		AND ci.purchase_cost IS NOT NULL
		AND (ls.amortization_behavior IN ('active', 'pending', NULL) OR ls.amortization_behavior != 'terminal')
	`

	assetRows, err := r.db.Query(ctx, assetQuery, assetArgs...)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to query active assets")
		return nil, fmt.Errorf("failed to query active assets: %w", err)
	}
	defer assetRows.Close()

	type assetInfo struct {
		ID                 uuid.UUID
		Name               string
		CurrentBookValue   float64
		SalvageValue       float64
		PurchaseCost       float64
		UsefulLifeMonths   int
		AmortStartDate     time.Time
		CITypeID           uuid.UUID
		CITypeName         string
		MonthlyDepreciation float64
	}

	var assets []assetInfo
	totalBookValue := 0.0
	totalMonthlyDepreciation := 0.0
	totalSalvageValue := 0.0
	totalOriginalCost := 0.0 // OCC - Original Capitalized Cost
	totalGrossBookValue := 0.0 // GVB - Gross Book Value

	for assetRows.Next() {
		var asset assetInfo
		var salvageValue, purchaseCost sql.NullFloat64
		var amortStartDate sql.NullTime

		err := assetRows.Scan(
			&asset.ID,
			&asset.Name,
			&asset.CurrentBookValue,
			&salvageValue,
			&purchaseCost,
			&asset.UsefulLifeMonths,
			&amortStartDate,
			&asset.CITypeID,
			&asset.CITypeName,
		)
		if err != nil {
			r.logger.Error().Err(err).Msg("Failed to scan asset row")
			return nil, fmt.Errorf("failed to scan asset row: %w", err)
		}

		asset.SalvageValue = salvageValue.Float64
		asset.PurchaseCost = purchaseCost.Float64
		if amortStartDate.Valid {
			asset.AmortStartDate = amortStartDate.Time
		}

		// Calculate monthly depreciation: (purchase_cost - salvage_value) / useful_life_months
		if asset.UsefulLifeMonths > 0 && asset.PurchaseCost > asset.SalvageValue {
			asset.MonthlyDepreciation = (asset.PurchaseCost - asset.SalvageValue) / float64(asset.UsefulLifeMonths)
		}

		assets = append(assets, asset)
		totalBookValue += asset.CurrentBookValue
		totalMonthlyDepreciation += asset.MonthlyDepreciation
		totalSalvageValue += asset.SalvageValue
		totalOriginalCost += asset.PurchaseCost // OCC calculation
	}

	// Calculate GVB: OCC + net adjustments from ledger
	// Query for all adjustments within the date range
	adjustmentsQuery := `
		SELECT
			COALESCE(SUM(CASE WHEN amount > 0 THEN amount ELSE 0 END), 0) as positive_adj,
			COALESCE(SUM(CASE WHEN amount < 0 THEN ABS(amount) ELSE 0 END), 0) as negative_adj
		FROM amortization_ledger ale
		JOIN configuration_items ci ON ale.ci_id = ci.id
		LEFT JOIN ci_type_definitions ctd ON ci.ci_type = ctd.name
	` + whereClause + `
		AND ale.entry_type = 'adjustment'
	`

	var positiveAdj, negativeAdj sql.NullFloat64
	err = r.db.QueryRow(ctx, adjustmentsQuery, args...).Scan(&positiveAdj, &negativeAdj)
	if err != nil && err != pgx.ErrNoRows {
		r.logger.Error().Err(err).Msg("Failed to query adjustments")
		// Continue with zero adjustments
	}

	netAdjustments := 0.0
	if positiveAdj.Valid {
		netAdjustments += positiveAdj.Float64
	}
	if negativeAdj.Valid {
		netAdjustments -= negativeAdj.Float64
	}

	totalGrossBookValue = totalOriginalCost + netAdjustments

	if err := assetRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating asset rows: %w", err)
	}

	// Generate monthly entries from start date to end date
	// Strategy:
	// 1. For months at or before lastHistoricalMonth: only show actual ledger entries
	// 2. For months after lastHistoricalMonth: show projections starting from last actual book value

	currentDate := req.DateFrom
	runningBookValue := 0.0      // Will be initialized from last historical entry
	runningGrossBookValue := totalGrossBookValue // GVB stays constant unless there are adjustments
	accumulatedDepreciation := 0.0 // Running total of accumulated depreciation

	for currentDate.Before(req.DateTo) || currentDate.Equal(req.DateTo) {
		monthKey := currentDate.Format("2006-01")

		// Check if we have historical data for this month
		if histEntry, exists := historicalData[monthKey]; exists {
			// This is a historical month with actual ledger data
			// Add accumulated depreciation to this entry
			accumulatedDepreciation += histEntry.DepreciationAmount + histEntry.WriteOffAmount + histEntry.AdjustmentAmount
			histEntry.AccumulatedDepreciation = accumulatedDepreciation
			histEntry.GrossBookValue = runningGrossBookValue // GVB for this month

			response.MonthlyData = append(response.MonthlyData, histEntry)
			runningBookValue = histEntry.ClosingBookValue
		} else {
			// No historical data for this month
			// Only create projections if this month is AFTER the last historical month
			if !lastHistoricalMonth.IsZero() && (currentDate.Month() == lastHistoricalMonth.Month() && currentDate.Year() == lastHistoricalMonth.Year()) {
				// This is the same month as the last historical entry but has no data
				// Skip it to avoid duplicates
			} else if !lastHistoricalMonth.IsZero() && currentDate.Before(lastHistoricalMonth) {
				// This is a past month before the last historical entry
				// Don't show projections for past periods - skip it
			} else {
				// This is a future month - create projected entry
				// Calculate total depreciation for this month from active assets
				monthlyDepreciation := 0.0
				activeCount := 0

				// Get the first day of current month for comparison
				currentMonthStart := time.Date(currentDate.Year(), currentDate.Month(), 1, 0, 0, 0, 0, time.UTC)

				for _, asset := range assets {
					// Only depreciate if:
					// 1. Current month is on or after amortization start date
					// 2. Book value is above salvage value
					if currentMonthStart.After(asset.AmortStartDate) || currentMonthStart.Equal(asset.AmortStartDate) {
						if asset.CurrentBookValue > asset.SalvageValue {
							monthlyDepreciation += asset.MonthlyDepreciation
							activeCount++
						}
					}
				}

				// Update accumulated depreciation
				accumulatedDepreciation += monthlyDepreciation

				// Use running book value (from last actual entry) as opening
				openingBookValue := runningBookValue
				closingBookValue := openingBookValue - monthlyDepreciation
				if closingBookValue < 0 {
					closingBookValue = 0
				}

				projectedEntry := MonthlyScheduleEntry{
					Month:                   currentDate,
					IsProjected:             true,
					OpeningBookValue:        openingBookValue,
					GrossBookValue:          runningGrossBookValue, // GVB for this month
					DepreciationAmount:      monthlyDepreciation,
					WriteOffAmount:          0,
					AdjustmentAmount:        0,
					ClosingBookValue:        closingBookValue,
					AccumulatedDepreciation: accumulatedDepreciation,
					ActiveAssetsCount:       activeCount,
				}

				response.MonthlyData = append(response.MonthlyData, projectedEntry)
				runningBookValue = closingBookValue
			}
		}

		currentDate = currentDate.AddDate(0, 1, 0)
	}

	// Calculate summary statistics
	totalDepreciation := 0.0
	totalWriteOffs := 0.0
	totalAdjustments := 0.0

	for _, entry := range response.MonthlyData {
		totalDepreciation += entry.DepreciationAmount
		totalWriteOffs += entry.WriteOffAmount
		totalAdjustments += entry.AdjustmentAmount
	}

	monthsCount := len(response.MonthlyData)
	averageMonthlyExpense := 0.0
	if monthsCount > 0 {
		averageMonthlyExpense = totalDepreciation / float64(monthsCount)
	}

	// Get final book value as projected end value
	projectedEndValue := 0.0
	if len(response.MonthlyData) > 0 {
		projectedEndValue = response.MonthlyData[len(response.MonthlyData)-1].ClosingBookValue
	}

	// Calculate accounting percentages
	totalNetBookValue := totalBookValue // NBV is current book value
	totalAccumulatedDepreciation := totalDepreciation + totalWriteOffs // AD
	depreciationPercentage := 0.0
	if totalGrossBookValue > 0 {
		depreciationPercentage = (totalAccumulatedDepreciation / totalGrossBookValue) * 100
	}
	remainingPercentage := 0.0
	if totalGrossBookValue > 0 {
		remainingPercentage = (totalNetBookValue / totalGrossBookValue) * 100
	}

	response.Summary = ScheduleSummary{
		TotalOriginalCost:        totalOriginalCost,          // OCC
		TotalGrossBookValue:      totalGrossBookValue,        // GVB
		TotalNetBookValue:        totalNetBookValue,          // NBV
		TotalDepreciation:        totalDepreciation,          // AD
		TotalWriteOffs:           totalWriteOffs,
		TotalAdjustments:         totalAdjustments,           // Net ±
		TotalSalvageValue:        totalSalvageValue,          // SV
		AverageMonthlyExpense:    averageMonthlyExpense,
		ProjectedEndValue:        projectedEndValue,
		DepreciationPercentage:   depreciationPercentage,     // AD/GVB × 100
		RemainingPercentage:      remainingPercentage,        // NBV/GVB × 100
	}

	// Update top-level response fields
	response.TotalOriginalCost = totalOriginalCost           // OCC
	response.TotalGrossBookValue = totalGrossBookValue       // GVB
	response.TotalNetBookValue = totalNetBookValue           // NBV
	response.TotalSalvageValue = totalSalvageValue           // SV
	response.TotalAccumulatedDepreciation = totalAccumulatedDepreciation // AD

	// Group by CI type
	byCITypeMap := make(map[uuid.UUID]*CITypeScheduleSummary)
	for _, asset := range assets {
		if _, exists := byCITypeMap[asset.CITypeID]; !exists {
			byCITypeMap[asset.CITypeID] = &CITypeScheduleSummary{
				CITypeID:   asset.CITypeID,
				CITypeName: asset.CITypeName,
			}
		}
		byCITypeMap[asset.CITypeID].AssetCount++
		byCITypeMap[asset.CITypeID].TotalBookValue += asset.CurrentBookValue
		byCITypeMap[asset.CITypeID].MonthlyDepreciation += asset.MonthlyDepreciation
	}

	for _, summary := range byCITypeMap {
		response.ByCIType = append(response.ByCIType, *summary)
	}

	// Sort by CI type name
	sort.Slice(response.ByCIType, func(i, j int) bool {
		return response.ByCIType[i].CITypeName < response.ByCIType[j].CITypeName
	})

	// Set total salvage value
	response.TotalSalvageValue = totalSalvageValue

	r.logger.Info().
		Int("total_months", monthsCount).
		Float64("total_depreciation", totalDepreciation).
		Float64("total_salvage_value", totalSalvageValue).
		Int("ci_types", len(response.ByCIType)).
		Msg("Generated depreciation schedule")

	return response, nil
}

// GetCIsForProcessing retrieves CIs that need amortization processing
func (r *repository) GetCIsForProcessing(ctx context.Context, processingDate time.Time, limit int) ([]uuid.UUID, error) {
	query := `
		SELECT ci.id
		FROM configuration_items ci
		JOIN ci_type_definitions ctd ON ci.ci_type = ctd.name
		LEFT JOIN lifecycle_statuses ls ON ci.lifecycle_status_id = ls.id
		WHERE ctd.is_amortizable = true
		AND ci.amort_start_date IS NOT NULL
		AND ci.useful_life_months IS NOT NULL
		AND ci.current_book_value > 0
		AND (ls.amortization_behavior = 'active' OR ls.amortization_behavior = 'terminal' OR ls.amortization_behavior IS NULL)
		AND NOT EXISTS (
			SELECT 1 FROM amortization_ledger al
			WHERE al.ci_id = ci.id
			AND al.entry_date = $1
			AND al.entry_type IN ('monthly_depreciation', 'write_off')
		)
		ORDER BY ci.amort_start_date, ci.id
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, processingDate, limit)
	if err != nil {
		r.logger.ErrorService("amortization", "get_cis_for_processing", err, map[string]interface{}{
			"processing_date": processingDate,
			"query":           query,
		})
		return nil, fmt.Errorf("failed to get CIs for processing: %w", err)
	}
	defer rows.Close()

	r.logger.Debug().Str("component", "GetCIsForProcessing").Time("processing_date", processingDate).Int("limit", limit).Msg("Executing GetCIsForProcessing query")

	var ciIDs []uuid.UUID
	for rows.Next() {
		var ciID uuid.UUID
		if err := rows.Scan(&ciID); err != nil {
			r.logger.ErrorService("amortization", "get_cis_for_processing_scan", err, nil)
			return nil, fmt.Errorf("failed to scan CI ID: %w", err)
		}
		ciIDs = append(ciIDs, ciID)
		r.logger.Debug().Str("component", "GetCIsForProcessing").Str("ci_id", ciID.String()).Msg("Found eligible CI")
	}

	r.logger.Info().Str("component", "GetCIsForProcessing").Int("found_cis", len(ciIDs)).Msg("GetCIsForProcessing completed")
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

// GetAmortizationSettings retrieves global amortization settings
func (r *repository) GetAmortizationSettings(ctx context.Context) (*AmortizationSettings, error) {
	query := `
		SELECT
			id,
			currency,
			default_useful_life_months,
			created_at,
			updated_at,
			created_by,
			updated_by
		FROM amortization_settings
		WHERE id = 'global'
	`

	var settings AmortizationSettings
	var createdBy sql.NullString
	var updatedBy sql.NullString

	err := r.db.QueryRow(ctx, query).Scan(
		&settings.ID,
		&settings.Currency,
		&settings.DefaultUsefulLifeMonths,
		&settings.CreatedAt,
		&settings.UpdatedAt,
		&createdBy,
		&updatedBy,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Return default settings if not found
			return &AmortizationSettings{
				ID:                      "global",
				Currency:                "USD",
				DefaultUsefulLifeMonths: 36,
				CreatedAt:               time.Now(),
				UpdatedAt:               time.Now(),
			}, nil
		}
		r.logger.ErrorService("amortization", "get_amortization_settings", err, nil)
		return nil, fmt.Errorf("failed to get amortization settings: %w", err)
	}

	if createdBy.Valid {
		if uuid, err := uuid.Parse(createdBy.String); err == nil {
			settings.CreatedBy = &uuid
		}
	}
	if updatedBy.Valid {
		if uuid, err := uuid.Parse(updatedBy.String); err == nil {
			settings.UpdatedBy = &uuid
		}
	}

	return &settings, nil
}

// UpdateAmortizationSettings updates global amortization settings
func (r *repository) UpdateAmortizationSettings(ctx context.Context, settings *AmortizationSettings, userID uuid.UUID) error {
	query := `
		INSERT INTO amortization_settings (
			id,
			currency,
			default_useful_life_months,
			updated_at,
			updated_by
		) VALUES ('global', $1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			currency = EXCLUDED.currency,
			default_useful_life_months = EXCLUDED.default_useful_life_months,
			updated_at = EXCLUDED.updated_at,
			updated_by = EXCLUDED.updated_by
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query,
		settings.Currency,
		settings.DefaultUsefulLifeMonths,
		now,
		userID,
	)
	if err != nil {
		r.logger.ErrorService("amortization", "update_amortization_settings", err, map[string]interface{}{
			"currency":                   settings.Currency,
			"default_useful_life_months": settings.DefaultUsefulLifeMonths,
			"user_id":                    userID,
		})
		return fmt.Errorf("failed to update amortization settings: %w", err)
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
