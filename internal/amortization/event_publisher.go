package amortization

import (
	"context"

	"github.com/google/uuid"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

// EventPublisherInterface defines the event publishing interface
type EventPublisherInterface interface {
	PublishAmortizationConfigUpdated(ctx context.Context, event *AmortizationConfigUpdatedEvent) error
	PublishAmortizationProcessed(ctx context.Context, event *AmortizationProcessedEvent) error
	PublishAmortizationWrittenOff(ctx context.Context, event *AmortizationWrittenOffEvent) error
	PublishAmortizationAdjustmentCreated(ctx context.Context, event *AmortizationAdjustmentCreatedEvent) error
	PublishAmortizationRunStarted(ctx context.Context, event *AmortizationRunStartedEvent) error
	PublishAmortizationRunCompleted(ctx context.Context, event *AmortizationRunCompletedEvent) error
	PublishAmortizationRunFailed(ctx context.Context, event *AmortizationRunFailedEvent) error
}

// eventPublisher implements event publishing for amortization operations
type eventPublisher struct {
	logger *pustakaLogger.Logger
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher(logger *pustakaLogger.Logger) EventPublisherInterface {
	return &eventPublisher{
		logger: logger,
	}
}

// Event structures
type AmortizationConfigUpdatedEvent struct {
	CIID      uuid.UUID              `json:"ci_id"`
	UserID    uuid.UUID              `json:"user_id"`
	Changes   map[string]interface{} `json:"changes"`
	Timestamp string                 `json:"timestamp"`
}

type AmortizationProcessedEvent struct {
	CIID         uuid.UUID `json:"ci_id"`
	RunID        uuid.UUID `json:"run_id"`
	Amount       float64   `json:"amount"`
	BookValue    float64   `json:"book_value"`
	Timestamp    string    `json:"timestamp"`
}

type AmortizationWrittenOffEvent struct {
	CIID          uuid.UUID `json:"ci_id"`
	UserID        uuid.UUID `json:"user_id"`
	WriteOffAmount float64   `json:"write_off_amount"`
	Reason        string    `json:"reason"`
	Timestamp     string    `json:"timestamp"`
}

type AmortizationAdjustmentCreatedEvent struct {
	CIID        uuid.UUID `json:"ci_id"`
	UserID      uuid.UUID `json:"user_id"`
	AdjustmentID uuid.UUID `json:"adjustment_id"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	Timestamp   string    `json:"timestamp"`
}

type AmortizationRunStartedEvent struct {
	RunID        uuid.UUID `json:"run_id"`
	TriggeredBy  *uuid.UUID `json:"triggered_by"`
	TotalCIs     int       `json:"total_cis"`
	Timestamp    string    `json:"timestamp"`
}

type AmortizationRunCompletedEvent struct {
	RunID              uuid.UUID `json:"run_id"`
	Processed          int       `json:"processed"`
	Failed             int       `json:"failed"`
	Skipped            int       `json:"skipped"`
	TotalDepreciation  float64   `json:"total_depreciation"`
	Timestamp          string    `json:"timestamp"`
}

type AmortizationRunFailedEvent struct {
	RunID        uuid.UUID `json:"run_id"`
	ErrorMessage string    `json:"error_message"`
	Processed    int       `json:"processed"`
	Failed       int       `json:"failed"`
	Skipped      int       `json:"skipped"`
	Timestamp    string    `json:"timestamp"`
}

// Event publishing methods
func (e *eventPublisher) PublishAmortizationConfigUpdated(ctx context.Context, event *AmortizationConfigUpdatedEvent) error {
	// Stub implementation - would publish to message broker
	e.logger.Info().Interface("event", event).Msg("Publishing amortization config updated event")
	return nil
}

func (e *eventPublisher) PublishAmortizationProcessed(ctx context.Context, event *AmortizationProcessedEvent) error {
	// Stub implementation
	e.logger.Info().Interface("event", event).Msg("Publishing amortization processed event")
	return nil
}

func (e *eventPublisher) PublishAmortizationWrittenOff(ctx context.Context, event *AmortizationWrittenOffEvent) error {
	// Stub implementation
	e.logger.Info().Interface("event", event).Msg("Publishing amortization written off event")
	return nil
}

func (e *eventPublisher) PublishAmortizationAdjustmentCreated(ctx context.Context, event *AmortizationAdjustmentCreatedEvent) error {
	// Stub implementation
	e.logger.Info().Interface("event", event).Msg("Publishing amortization adjustment created event")
	return nil
}

func (e *eventPublisher) PublishAmortizationRunStarted(ctx context.Context, event *AmortizationRunStartedEvent) error {
	// Stub implementation
	e.logger.Info().Interface("event", event).Msg("Publishing amortization run started event")
	return nil
}

func (e *eventPublisher) PublishAmortizationRunCompleted(ctx context.Context, event *AmortizationRunCompletedEvent) error {
	// Stub implementation
	e.logger.Info().Interface("event", event).Msg("Publishing amortization run completed event")
	return nil
}

func (e *eventPublisher) PublishAmortizationRunFailed(ctx context.Context, event *AmortizationRunFailedEvent) error {
	// Stub implementation
	e.logger.Info().Interface("event", event).Msg("Publishing amortization run failed event")
	return nil
}