package ci

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

type auditLogRepository struct {
	pool   *pgxpool.Pool
	logger *pustakaLogger.Logger
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(pool *pgxpool.Pool, logger *pustakaLogger.Logger) AuditLogRepository {
	return &auditLogRepository{
		pool:   pool,
		logger: logger,
	}
}

func (r *auditLogRepository) Create(ctx context.Context, auditLog *AuditLog) error {
	query := `
		INSERT INTO audit_logs (entity_type, entity_id, action, performed_by, timestamp, details, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var detailsJSON []byte
	var err error
	if auditLog.Details != nil {
		detailsJSON, err = json.Marshal(auditLog.Details)
		if err != nil {
			return fmt.Errorf("failed to marshal audit log details: %w", err)
		}
	}

	var entityID interface{}
	if auditLog.EntityID != nil {
		entityID = auditLog.EntityID
	}

	err = r.pool.QueryRow(ctx, query,
		auditLog.EntityType,
		entityID,
		auditLog.Action,
		auditLog.PerformedBy,
		auditLog.Timestamp,
		detailsJSON,
		auditLog.IPAddress,
		auditLog.UserAgent,
	).Scan(&auditLog.ID)

	if err != nil {
		r.logger.Error().
			Err(err).
			Str("entity_type", auditLog.EntityType).
			Str("action", auditLog.Action).
			Msg("Failed to create audit log")
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	r.logger.Debug().
		Str("audit_log_id", auditLog.ID.String()).
		Str("entity_type", auditLog.EntityType).
		Str("action", auditLog.Action).
		Str("performed_by", auditLog.PerformedBy.String()).
		Msg("Audit log created")

	return nil
}

func (r *auditLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*AuditLog, error) {
	query := `
		SELECT id, entity_type, entity_id, action, performed_by, timestamp, details, ip_address, user_agent
		FROM audit_logs
		WHERE id = $1
	`

	var auditLog AuditLog
	var entityID pgtype.UUID
	var detailsJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&auditLog.ID,
		&auditLog.EntityType,
		&entityID,
		&auditLog.Action,
		&auditLog.PerformedBy,
		&auditLog.Timestamp,
		&detailsJSON,
		&auditLog.IPAddress,
		&auditLog.UserAgent,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("audit log not found")
		}
		r.logger.Error().Err(err).Str("audit_log_id", id.String()).Msg("Failed to get audit log")
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	if len(entityID.Bytes) > 0 {
		uid := uuid.UUID(entityID.Bytes)
		auditLog.EntityID = &uid
	}

	if len(detailsJSON) > 0 {
		if err := json.Unmarshal(detailsJSON, &auditLog.Details); err != nil {
			r.logger.Error().Err(err).Str("audit_log_id", id.String()).Msg("Failed to unmarshal audit log details")
			auditLog.Details = make(map[string]interface{})
		}
	}

	return &auditLog, nil
}

func (r *auditLogRepository) List(ctx context.Context, filters AuditLogFilters) (*AuditLogListResponse, error) {
	// Set defaults
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.Limit <= 0 {
		filters.Limit = 50
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}
	if filters.Sort == "" {
		filters.Sort = "timestamp"
	}
	if filters.Order == "" {
		filters.Order = "desc"
	}

	offset := (filters.Page - 1) * filters.Limit

	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if filters.EntityType != "" {
		whereClause += fmt.Sprintf(" AND entity_type = $%d", argIndex)
		args = append(args, filters.EntityType)
		argIndex++
	}

	if filters.EntityID != nil {
		whereClause += fmt.Sprintf(" AND entity_id = $%d", argIndex)
		args = append(args, *filters.EntityID)
		argIndex++
	}

	if filters.Action != "" {
		whereClause += fmt.Sprintf(" AND action = $%d", argIndex)
		args = append(args, filters.Action)
		argIndex++
	}

	if filters.PerformedBy != nil {
		whereClause += fmt.Sprintf(" AND performed_by = $%d", argIndex)
		args = append(args, *filters.PerformedBy)
		argIndex++
	}

	if filters.StartDate != nil {
		whereClause += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
		args = append(args, *filters.StartDate)
		argIndex++
	}

	if filters.EndDate != nil {
		whereClause += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
		args = append(args, *filters.EndDate)
		argIndex++
	}

	if filters.Search != "" {
		whereClause += fmt.Sprintf(" AND (LOWER(entity_type) LIKE $%d OR LOWER(action) LIKE $%d OR LOWER(details::text) LIKE $%d)", argIndex, argIndex+1, argIndex+2)
		searchTerm := "%" + filters.Search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
		argIndex += 3
	}

	// Validate sort column
	validSortColumns := map[string]bool{
		"timestamp":    true,
		"entity_type":  true,
		"action":       true,
		"performed_by": true,
		"ip_address":   true,
	}
	if !validSortColumns[filters.Sort] {
		filters.Sort = "timestamp"
	}

	// Validate order
	if filters.Order != "asc" && filters.Order != "desc" {
		filters.Order = "desc"
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM audit_logs " + whereClause
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to count audit logs")
		return nil, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Get audit logs
	query := `
		SELECT id, entity_type, entity_id, action, performed_by, timestamp, details, ip_address, user_agent
		FROM audit_logs ` + whereClause + `
		ORDER BY ` + filters.Sort + " " + filters.Order + `
		LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)

	args = append(args, filters.Limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to query audit logs")
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var auditLogs []AuditLog
	for rows.Next() {
		var auditLog AuditLog
		var entityID pgtype.UUID
		var detailsJSON []byte

		err := rows.Scan(
			&auditLog.ID,
			&auditLog.EntityType,
			&entityID,
			&auditLog.Action,
			&auditLog.PerformedBy,
			&auditLog.Timestamp,
			&detailsJSON,
			&auditLog.IPAddress,
			&auditLog.UserAgent,
		)

		if err != nil {
			r.logger.Error().Err(err).Msg("Failed to scan audit log row")
			continue
		}

		if len(entityID.Bytes) > 0 {
			uid := uuid.UUID(entityID.Bytes)
			auditLog.EntityID = &uid
		}

		if len(detailsJSON) > 0 {
			if err := json.Unmarshal(detailsJSON, &auditLog.Details); err != nil {
				r.logger.Error().Err(err).Str("audit_log_id", auditLog.ID.String()).Msg("Failed to unmarshal audit log details")
				auditLog.Details = make(map[string]interface{})
			}
		}

		auditLogs = append(auditLogs, auditLog)
	}

	totalPages := int((total + int64(filters.Limit) - 1) / int64(filters.Limit))

	return &AuditLogListResponse{
		AuditLogs: auditLogs,
		Pagination: PaginationResponse{
			Page:       filters.Page,
			Limit:      filters.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func (r *auditLogRepository) GetStats(ctx context.Context, filters AuditLogFilters) (*AuditLogStats, error) {
	// Build base WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if filters.StartDate != nil {
		whereClause += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
		args = append(args, *filters.StartDate)
		argIndex++
	}

	if filters.EndDate != nil {
		whereClause += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
		args = append(args, *filters.EndDate)
		argIndex++
	}

	stats := &AuditLogStats{
		EventsByType:   make(map[string]int64),
		EventsByAction: make(map[string]int64),
		EventsByUser:   make(map[string]int64),
		DailyActivity:  make(map[string]int64),
	}

	// Get total count
	totalQuery := "SELECT COUNT(*) FROM audit_logs " + whereClause
	err := r.pool.QueryRow(ctx, totalQuery, args...).Scan(&stats.TotalEvents)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get total audit log count")
		return nil, fmt.Errorf("failed to get total audit log count: %w", err)
	}

	// Get events by type
	typeQuery := `
		SELECT entity_type, COUNT(*)
		FROM audit_logs ` + whereClause + `
		GROUP BY entity_type
		ORDER BY COUNT(*) DESC
		LIMIT 10
	`
	typeRows, err := r.pool.Query(ctx, typeQuery, args...)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get events by type")
	} else {
		defer typeRows.Close()
		for typeRows.Next() {
			var entityType string
			var count int64
			if err := typeRows.Scan(&entityType, &count); err == nil {
				stats.EventsByType[entityType] = count
			}
		}
	}

	// Get events by action
	actionQuery := `
		SELECT action, COUNT(*)
		FROM audit_logs ` + whereClause + `
		GROUP BY action
		ORDER BY COUNT(*) DESC
		LIMIT 10
	`
	actionRows, err := r.pool.Query(ctx, actionQuery, args...)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get events by action")
	} else {
		defer actionRows.Close()
		for actionRows.Next() {
			var action string
			var count int64
			if err := actionRows.Scan(&action, &count); err == nil {
				stats.EventsByAction[action] = count
			}
		}
	}

	// Get events by user
	userQuery := `
		SELECT u.username, COUNT(*)
		FROM audit_logs al
		JOIN users u ON al.performed_by = u.id
		` + whereClause + `
		GROUP BY u.username
		ORDER BY COUNT(*) DESC
		LIMIT 10
	`
	userRows, err := r.pool.Query(ctx, userQuery, args...)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get events by user")
	} else {
		defer userRows.Close()
		for userRows.Next() {
			var username string
			var count int64
			if err := userRows.Scan(&username, &count); err == nil {
				stats.EventsByUser[username] = count
			}
		}
	}

	// Get daily activity (last 30 days)
	dailyQuery := `
		SELECT DATE(timestamp) as date, COUNT(*)
		FROM audit_logs
		WHERE timestamp >= NOW() - INTERVAL '30 days'
		GROUP BY DATE(timestamp)
		ORDER BY date DESC
	`
	dailyRows, err := r.pool.Query(ctx, dailyQuery)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get daily activity")
	} else {
		defer dailyRows.Close()
		for dailyRows.Next() {
			var date time.Time
			var count int64
			if err := dailyRows.Scan(&date, &count); err == nil {
				stats.DailyActivity[date.Format("2006-01-02")] = count
			}
		}
	}

	// Get recent activity
	recentQuery := `
		SELECT al.id, al.entity_type, al.entity_id, al.action, al.performed_by, al.timestamp, al.ip_address, u.username
		FROM audit_logs al
		JOIN users u ON al.performed_by = u.id
		` + whereClause + `
		ORDER BY al.timestamp DESC
		LIMIT 10
	`
	recentRows, err := r.pool.Query(ctx, recentQuery, args...)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get recent activity")
	} else {
		defer recentRows.Close()
		for recentRows.Next() {
			var auditLog AuditLog
			var entityID pgtype.UUID
			var username string

			err := recentRows.Scan(
				&auditLog.ID,
				&auditLog.EntityType,
				&entityID,
				&auditLog.Action,
				&auditLog.PerformedBy,
				&auditLog.Timestamp,
				&auditLog.IPAddress,
				&username,
			)

			if err == nil {
				if len(entityID.Bytes) > 0 {
					uid := uuid.UUID(entityID.Bytes)
					auditLog.EntityID = &uid
				}
				auditLog.Details = map[string]interface{}{
					"username": username,
				}
				stats.RecentActivity = append(stats.RecentActivity, auditLog)
			}
		}
	}

	return stats, nil
}

func (r *auditLogRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM audit_logs WHERE id = $1"

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		r.logger.Error().Err(err).Str("audit_log_id", id.String()).Msg("Failed to delete audit log")
		return fmt.Errorf("failed to delete audit log: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("audit log not found")
	}

	r.logger.Info().Str("audit_log_id", id.String()).Msg("Audit log deleted")
	return nil
}

func (r *auditLogRepository) DeleteOld(ctx context.Context, olderThan time.Time) (int64, error) {
	query := "DELETE FROM audit_logs WHERE timestamp < $1"

	result, err := r.pool.Exec(ctx, query, olderThan)
	if err != nil {
		r.logger.Error().Err(err).Time("older_than", olderThan).Msg("Failed to delete old audit logs")
		return 0, fmt.Errorf("failed to delete old audit logs: %w", err)
	}

	rowsAffected := result.RowsAffected()
	r.logger.Info().
		Time("older_than", olderThan).
		Int64("deleted_count", rowsAffected).
		Msg("Old audit logs deleted")

	return rowsAffected, nil
}