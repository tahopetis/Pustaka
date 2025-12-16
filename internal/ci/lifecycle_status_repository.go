package ci

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LifecycleStatusRepository handles database operations for lifecycle statuses
type LifecycleStatusRepository struct {
	db *pgxpool.Pool
}

// NewLifecycleStatusRepository creates a new lifecycle status repository
func NewLifecycleStatusRepository(db *pgxpool.Pool) *LifecycleStatusRepository {
	return &LifecycleStatusRepository{
		db: db,
	}
}

// Create creates a new lifecycle status
func (r *LifecycleStatusRepository) Create(ctx context.Context, ls *LifecycleStatus) (*LifecycleStatus, error) {
	query := `
		INSERT INTO lifecycle_statuses (
			id, name, display_name, description, color, icon, sort_order,
			is_active, is_system, created_at, updated_at, created_by, updated_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		) RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		ls.ID, ls.Name, ls.DisplayName, ls.Description, ls.Color, ls.Icon, ls.SortOrder,
		ls.IsActive, ls.IsSystem, ls.CreatedAt, ls.UpdatedAt, ls.CreatedBy, ls.UpdatedBy,
	).Scan(&ls.ID, &ls.CreatedAt, &ls.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create lifecycle status: %w", err)
	}

	return ls, nil
}

// GetByID retrieves a lifecycle status by ID
func (r *LifecycleStatusRepository) GetByID(ctx context.Context, id uuid.UUID) (*LifecycleStatus, error) {
	query := `
		SELECT id, name, display_name, description, color, icon, sort_order,
			   is_active, is_system, created_at, updated_at, created_by, updated_by
		FROM lifecycle_statuses
		WHERE id = $1
	`

	ls := &LifecycleStatus{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&ls.ID, &ls.Name, &ls.DisplayName, &ls.Description, &ls.Color, &ls.Icon, &ls.SortOrder,
		&ls.IsActive, &ls.IsSystem, &ls.CreatedAt, &ls.UpdatedAt, &ls.CreatedBy, &ls.UpdatedBy,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("lifecycle status not found")
		}
		return nil, fmt.Errorf("failed to get lifecycle status: %w", err)
	}

	return ls, nil
}

// GetByName retrieves a lifecycle status by name
func (r *LifecycleStatusRepository) GetByName(ctx context.Context, name string) (*LifecycleStatus, error) {
	query := `
		SELECT id, name, display_name, description, color, icon, sort_order,
			   is_active, is_system, created_at, updated_at, created_by, updated_by
		FROM lifecycle_statuses
		WHERE name = $1
	`

	ls := &LifecycleStatus{}
	err := r.db.QueryRow(ctx, query, name).Scan(
		&ls.ID, &ls.Name, &ls.DisplayName, &ls.Description, &ls.Color, &ls.Icon, &ls.SortOrder,
		&ls.IsActive, &ls.IsSystem, &ls.CreatedAt, &ls.UpdatedAt, &ls.CreatedBy, &ls.UpdatedBy,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("lifecycle status not found")
		}
		return nil, fmt.Errorf("failed to get lifecycle status: %w", err)
	}

	return ls, nil
}

// List retrieves lifecycle statuses with pagination and filtering
func (r *LifecycleStatusRepository) List(ctx context.Context, filters *ListLifecycleStatusFilters, page, limit int) (*LifecycleStatusListResponse, error) {
	offset := (page - 1) * limit

	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if filters != nil {
		if filters.Search != "" {
			whereClause += fmt.Sprintf(" AND (name ILIKE $%d OR display_name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex+1, argIndex+2)
			searchPattern := "%" + filters.Search + "%"
			args = append(args, searchPattern, searchPattern, searchPattern)
			argIndex += 3
		}
		if filters.IsActive != nil {
			whereClause += fmt.Sprintf(" AND is_active = $%d", argIndex)
			args = append(args, *filters.IsActive)
			argIndex++
		}
		if filters.IsSystem != nil {
			whereClause += fmt.Sprintf(" AND is_system = $%d", argIndex)
			args = append(args, *filters.IsSystem)
			argIndex++
		}
	}

	// Build ORDER BY clause
	orderClause := "ORDER BY sort_order ASC, display_name ASC"
	if filters != nil && filters.Sort != "" {
		validSorts := map[string]bool{
			"name": true, "display_name": true, "sort_order": true,
			"created_at": true, "updated_at": true,
		}
		if validSorts[filters.Sort] {
			orderClause = fmt.Sprintf("ORDER BY %s", filters.Sort)
			if filters.Order == "desc" {
				orderClause += " DESC"
			}
		}
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM lifecycle_statuses %s", whereClause)
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count lifecycle statuses: %w", err)
	}

	// Get records
	query := fmt.Sprintf(`
		SELECT id, name, display_name, description, color, icon, sort_order,
			   is_active, is_system, created_at, updated_at, created_by, updated_by
		FROM lifecycle_statuses
		%s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderClause, argIndex, argIndex+1)

	args = append(args, limit, offset)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list lifecycle statuses: %w", err)
	}
	defer rows.Close()

	var statuses []LifecycleStatus
	for rows.Next() {
		var ls LifecycleStatus
		err := rows.Scan(
			&ls.ID, &ls.Name, &ls.DisplayName, &ls.Description, &ls.Color, &ls.Icon, &ls.SortOrder,
			&ls.IsActive, &ls.IsSystem, &ls.CreatedAt, &ls.UpdatedAt, &ls.CreatedBy, &ls.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lifecycle status row: %w", err)
		}
		statuses = append(statuses, ls)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating lifecycle status rows: %w", err)
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &LifecycleStatusListResponse{
		LifecycleStatuses: statuses,
		Page:              page,
		Limit:             limit,
		Total:             total,
		TotalPages:        totalPages,
	}, nil
}

// GetActive retrieves all active lifecycle statuses ordered by sort_order
func (r *LifecycleStatusRepository) GetActive(ctx context.Context) ([]LifecycleStatus, error) {
	query := `
		SELECT id, name, display_name, description, color, icon, sort_order,
			   is_active, is_system, created_at, updated_at, created_by, updated_by
		FROM lifecycle_statuses
		WHERE is_active = true
		ORDER BY sort_order ASC, display_name ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active lifecycle statuses: %w", err)
	}
	defer rows.Close()

	var statuses []LifecycleStatus
	for rows.Next() {
		var ls LifecycleStatus
		err := rows.Scan(
			&ls.ID, &ls.Name, &ls.DisplayName, &ls.Description, &ls.Color, &ls.Icon, &ls.SortOrder,
			&ls.IsActive, &ls.IsSystem, &ls.CreatedAt, &ls.UpdatedAt, &ls.CreatedBy, &ls.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lifecycle status row: %w", err)
		}
		statuses = append(statuses, ls)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating lifecycle status rows: %w", err)
	}

	return statuses, nil
}

// Update updates an existing lifecycle status
func (r *LifecycleStatusRepository) Update(ctx context.Context, ls *LifecycleStatus) (*LifecycleStatus, error) {
	query := `
		UPDATE lifecycle_statuses
		SET display_name = $2, description = $3, color = $4, icon = $5, sort_order = $6,
			is_active = $7, updated_at = $8, updated_by = $9
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRow(ctx, query,
		ls.ID, ls.DisplayName, ls.Description, ls.Color, ls.Icon, ls.SortOrder,
		ls.IsActive, time.Now(), ls.UpdatedBy,
	).Scan(&ls.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update lifecycle status: %w", err)
	}

	return ls, nil
}

// Delete performs a hard delete of a lifecycle status (only for non-system statuses not in use)
func (r *LifecycleStatusRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM lifecycle_statuses WHERE id = $1 AND is_system = false`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete lifecycle status: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("lifecycle status not found or cannot be deleted")
	}

	return nil
}

// CountCIsWithStatus counts how many CIs use a specific lifecycle status
func (r *LifecycleStatusRepository) CountCIsWithStatus(ctx context.Context, statusID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM configuration_items WHERE lifecycle_status_id = $1`

	var count int64
	err := r.db.QueryRow(ctx, query, statusID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count CIs with lifecycle status: %w", err)
	}

	return count, nil
}

// GetUsageStats retrieves usage statistics for all lifecycle statuses
func (r *LifecycleStatusRepository) GetUsageStats(ctx context.Context) (*LifecycleStatusUsageResponse, error) {
	// Get total CIs count
	var totalCIs, cisWithStatus int64
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM configuration_items").Scan(&totalCIs)
	if err != nil {
		return nil, fmt.Errorf("failed to count total CIs: %w", err)
	}

	err = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM configuration_items WHERE lifecycle_status_id IS NOT NULL").Scan(&cisWithStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to count CIs with status: %w", err)
	}

	// Get usage by status
	query := `
		SELECT ls.id, ls.name, ls.display_name, ls.description, ls.color, ls.icon,
			   ls.sort_order, ls.is_active, ls.is_system, ls.created_at, ls.updated_at,
			   ls.created_by, ls.updated_by, COUNT(ci.id) as usage_count
		FROM lifecycle_statuses ls
		LEFT JOIN configuration_items ci ON ls.id = ci.lifecycle_status_id
		GROUP BY ls.id, ls.name, ls.display_name, ls.description, ls.color, ls.icon,
				 ls.sort_order, ls.is_active, ls.is_system, ls.created_at, ls.updated_at,
				 ls.created_by, ls.updated_by
		ORDER BY ls.sort_order ASC, ls.display_name ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get lifecycle status usage: %w", err)
	}
	defer rows.Close()

	var statusUsage []LifecycleStatusUsage
	statusDistribution := make(map[string]int64)

	for rows.Next() {
		var ls LifecycleStatus
		var usageCount int64
		err := rows.Scan(
			&ls.ID, &ls.Name, &ls.DisplayName, &ls.Description, &ls.Color, &ls.Icon,
			&ls.SortOrder, &ls.IsActive, &ls.IsSystem, &ls.CreatedAt, &ls.UpdatedAt,
			&ls.CreatedBy, &ls.UpdatedBy, &usageCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage row: %w", err)
		}

		statusUsage = append(statusUsage, LifecycleStatusUsage{
			LifecycleStatus: ls,
			UsageCount:      usageCount,
		})

		statusDistribution[ls.Name] = usageCount
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating usage rows: %w", err)
	}

	return &LifecycleStatusUsageResponse{
		TotalCIs:          totalCIs,
		CIsWithStatus:     cisWithStatus,
		CIsWithoutStatus:  totalCIs - cisWithStatus,
		StatusUsage:       statusUsage,
		StatusDistribution: statusDistribution,
	}, nil
}