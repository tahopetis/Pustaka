package ci

import (
	"context"
	"time"

	"github.com/google/uuid"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

type AuditService struct {
	repo   AuditLogRepository
	logger *pustakaLogger.Logger
}

// NewAuditService creates a new audit service
func NewAuditService(repo AuditLogRepository, logger *pustakaLogger.Logger) *AuditService {
	return &AuditService{
		repo:   repo,
		logger: logger,
	}
}

// CreateAuditLog creates a new audit log entry
func (s *AuditService) CreateAuditLog(ctx context.Context, entityType string, entityID *uuid.UUID, action string, performedBy uuid.UUID, details map[string]interface{}, ipAddress, userAgent string) error {
	auditLog := &AuditLog{
		EntityType:  entityType,
		EntityID:    entityID,
		Action:      action,
		PerformedBy: performedBy,
		Timestamp:   time.Now(),
		Details:     details,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	}

	if err := s.repo.Create(ctx, auditLog); err != nil {
		s.logger.Error().
			Err(err).
			Str("entity_type", entityType).
			Str("action", action).
			Msg("Failed to create audit log")
		return err
	}

	s.logger.Debug().
		Str("entity_type", entityType).
		Str("action", action).
		Str("performed_by", performedBy.String()).
		Msg("Audit log created successfully")

	return nil
}

// GetAuditLog gets an audit log by ID
func (s *AuditService) GetAuditLog(ctx context.Context, id uuid.UUID) (*AuditLog, error) {
	return s.repo.GetByID(ctx, id)
}

// ListAuditLogs lists audit logs with filters
func (s *AuditService) ListAuditLogs(ctx context.Context, filters AuditLogFilters) (*AuditLogListResponse, error) {
	return s.repo.List(ctx, filters)
}

// GetAuditStats gets audit log statistics
func (s *AuditService) GetAuditStats(ctx context.Context, filters AuditLogFilters) (*AuditLogStats, error) {
	return s.repo.GetStats(ctx, filters)
}

// DeleteAuditLog deletes an audit log (for compliance/GDPR purposes)
func (s *AuditService) DeleteAuditLog(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// DeleteOldAuditLogs deletes audit logs older than the specified date
func (s *AuditService) DeleteOldAuditLogs(ctx context.Context, olderThan time.Time) (int64, error) {
	return s.repo.DeleteOld(ctx, olderThan)
}

// GetRetentionPeriod returns the default retention period for audit logs
func (s *AuditService) GetRetentionPeriod() time.Duration {
	return 365 * 24 * time.Hour // 1 year by default
}

// CleanupOldAuditLogs deletes audit logs older than the retention period
func (s *AuditService) CleanupOldAuditLogs(ctx context.Context) (int64, error) {
	retentionPeriod := s.GetRetentionPeriod()
	cutoffDate := time.Now().Add(-retentionPeriod)

	s.logger.Info().
		Time("cutoff_date", cutoffDate).
		Dur("retention_period", retentionPeriod).
		Msg("Starting audit log cleanup")

	return s.DeleteOldAuditLogs(ctx, cutoffDate)
}