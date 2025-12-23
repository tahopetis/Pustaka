package amortization

import (
	"context"

	"github.com/google/uuid"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

// AuditLoggerInterface defines the audit logging interface
type AuditLoggerInterface interface {
	LogAmortizationConfigUpdated(ctx context.Context, ciID uuid.UUID, userID uuid.UUID, changes map[string]interface{}) error
	LogAmortizationProcessed(ctx context.Context, ciID uuid.UUID, runID uuid.UUID, amount float64, bookValue float64) error
	LogAmortizationWriteOff(ctx context.Context, ciID uuid.UUID, userID uuid.UUID, amount float64, reason string) error
	LogAmortizationAdjustmentCreated(ctx context.Context, ciID uuid.UUID, userID uuid.UUID, amount float64, description string) error
	LogAmortizationRestructured(ctx context.Context, ciID uuid.UUID, userID uuid.UUID, oldUsefulLife, newUsefulLife int, reason string) error
	LogAmortizationRunStarted(ctx context.Context, runID uuid.UUID, triggeredBy uuid.UUID, totalCIs int) error
	LogAmortizationRunCompleted(ctx context.Context, runID uuid.UUID, processed, failed, skipped int) error
}

// auditLogger implements audit logging for amortization operations
type auditLogger struct {
	auditService interface{} // This would be the actual audit service
	logger      *pustakaLogger.Logger
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(auditService interface{}, logger *pustakaLogger.Logger) AuditLoggerInterface {
	return &auditLogger{
		auditService: auditService,
		logger:      logger,
	}
}

// Audit logging methods
func (a *auditLogger) LogAmortizationConfigUpdated(ctx context.Context, ciID uuid.UUID, userID uuid.UUID, changes map[string]interface{}) error {
	// Fixed implementation using proper zerolog methods
	a.logger.Info().
		Str("ci_id", ciID.String()).
		Str("user_id", userID.String()).
		Interface("changes", changes).
		Msg("Amortization configuration updated")
	return nil
}

func (a *auditLogger) LogAmortizationProcessed(ctx context.Context, ciID uuid.UUID, runID uuid.UUID, amount float64, bookValue float64) error {
	// Fixed implementation
	a.logger.Info().
		Str("ci_id", ciID.String()).
		Str("run_id", runID.String()).
		Float64("amount", amount).
		Float64("book_value", bookValue).
		Msg("Amortization processed")
	return nil
}

func (a *auditLogger) LogAmortizationWriteOff(ctx context.Context, ciID uuid.UUID, userID uuid.UUID, amount float64, reason string) error {
	// Fixed implementation
	a.logger.Info().
		Str("ci_id", ciID.String()).
		Str("user_id", userID.String()).
		Float64("amount", amount).
		Str("reason", reason).
		Msg("Amortization write-off")
	return nil
}

func (a *auditLogger) LogAmortizationAdjustmentCreated(ctx context.Context, ciID uuid.UUID, userID uuid.UUID, amount float64, description string) error {
	// Fixed implementation
	a.logger.Info().
		Str("ci_id", ciID.String()).
		Str("user_id", userID.String()).
		Float64("amount", amount).
		Str("description", description).
		Msg("Amortization adjustment created")
	return nil
}

func (a *auditLogger) LogAmortizationRestructured(ctx context.Context, ciID uuid.UUID, userID uuid.UUID, oldUsefulLife, newUsefulLife int, reason string) error {
	// Fixed implementation
	a.logger.Info().
		Str("ci_id", ciID.String()).
		Str("user_id", userID.String()).
		Int("old_useful_life_months", oldUsefulLife).
		Int("new_useful_life_months", newUsefulLife).
		Str("reason", reason).
		Msg("Amortization restructured")
	return nil
}

func (a *auditLogger) LogAmortizationRunStarted(ctx context.Context, runID uuid.UUID, triggeredBy uuid.UUID, totalCIs int) error {
	// Fixed implementation
	a.logger.Info().
		Str("run_id", runID.String()).
		Str("triggered_by", triggeredBy.String()).
		Int("total_cis", totalCIs).
		Msg("Amortization run started")
	return nil
}

func (a *auditLogger) LogAmortizationRunCompleted(ctx context.Context, runID uuid.UUID, processed, failed, skipped int) error {
	// Fixed implementation
	a.logger.Info().
		Str("run_id", runID.String()).
		Int("processed", processed).
		Int("failed", failed).
		Int("skipped", skipped).
		Msg("Amortization run completed")
	return nil
}