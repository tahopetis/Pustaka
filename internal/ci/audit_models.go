package ci

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID          uuid.UUID            `json:"id" db:"id"`
	EntityType  string               `json:"entity_type" db:"entity_type"`
	EntityID    *uuid.UUID           `json:"entity_id" db:"entity_id"`
	Action      string               `json:"action" db:"action"`
	PerformedBy uuid.UUID            `json:"performed_by" db:"performed_by"`
	Timestamp   time.Time            `json:"timestamp" db:"timestamp"`
	Details     map[string]interface{} `json:"details" db:"details"`
	IPAddress   string               `json:"ip_address" db:"ip_address"`
	UserAgent   string               `json:"user_agent" db:"user_agent"`
}

// AuditLogFilters represents filters for querying audit logs
type AuditLogFilters struct {
	EntityType  string     `json:"entity_type,omitempty"`
	EntityID   *uuid.UUID `json:"entity_id,omitempty"`
	Action     string     `json:"action,omitempty"`
	PerformedBy *uuid.UUID `json:"performed_by,omitempty"`
	StartDate  *time.Time `json:"start_date,omitempty"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	Search     string     `json:"search,omitempty"`
	Sort       string     `json:"sort,omitempty"`
	Order      string     `json:"order,omitempty"`
	Page       int        `json:"page,omitempty"`
	Limit      int        `json:"limit,omitempty"`
}

// AuditLogListResponse represents a paginated list of audit logs
type AuditLogListResponse struct {
	AuditLogs  []AuditLog         `json:"audit_logs"`
	Pagination PaginationResponse `json:"pagination"`
}

// AuditLogStats represents audit log statistics
type AuditLogStats struct {
	TotalEvents       int64                         `json:"total_events"`
	EventsByType      map[string]int64              `json:"events_by_type"`
	EventsByAction    map[string]int64              `json:"events_by_action"`
	EventsByUser      map[string]int64              `json:"events_by_user"`
	RecentActivity    []AuditLog                    `json:"recent_activity"`
	DailyActivity     map[string]int64              `json:"daily_activity"`
}

// AuditLogRepository defines the interface for audit log operations
type AuditLogRepository interface {
	Create(ctx context.Context, auditLog *AuditLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*AuditLog, error)
	List(ctx context.Context, filters AuditLogFilters) (*AuditLogListResponse, error)
	GetStats(ctx context.Context, filters AuditLogFilters) (*AuditLogStats, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteOld(ctx context.Context, olderThan time.Time) (int64, error)
}