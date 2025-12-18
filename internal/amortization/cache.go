package amortization

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

// CacheRepository implements caching operations for amortization using Redis
type CacheRepository struct {
	client    *redis.Client
	logger    *pustakaLogger.Logger
	keyPrefix string
}

// NewCacheRepository creates a new cache repository
func NewCacheRepository(client *redis.Client, logger *pustakaLogger.Logger) *CacheRepository {
	return &CacheRepository{
		client:    client,
		logger:    logger,
		keyPrefix: "amortization:",
	}
}

// Cache key patterns
const (
	// Amortizable CI cache keys
	CIKeyPattern        = "ci:%s"
	CILatestLedgerKey   = "ci:%s:latest_ledger"
	CISummariesKey      = "summaries:%s"

	// Type and status cache keys (reference data)
	CITypeAmortizableKey = "ci_type:%s:amortizable"
	StatusBehaviorKey   = "status:%s:behavior"

	// Scheduler cache keys
	ProcessingLockKey   = "scheduler:processing_lock"
	JobQueueKey         = "jobs:queue"
	JobStatsKey         = "jobs:stats"
	LastRunKey          = "scheduler:last_run"
	SchedulerConfigKey  = "scheduler:config"

	// Session cache keys (for pagination, filters, etc.)
	UserSessionPrefix   = "session:user:%s:"
	ActiveFiltersKey    = "filters:active"
	RecentSearchesKey   = "searches:recent"

	// Report cache keys
	ReportPrefix         = "report:%s"
	DepreciationScheduleKey = "report:deprecation_schedule:%s"
)

// Default TTL values
const (
	DefaultTTL            = 5 * time.Minute
	ReferenceDataTTL      = 30 * time.Minute
	LedgerTTL             = 2 * time.Minute
	SummaryTTL           = 10 * time.Minute
	SessionTTL           = 30 * time.Minute
	ReportTTL            = 1 * time.Hour
	LockTTL              = 1 * time.Hour
	JobTTL               = 24 * time.Hour
)

// AmortizableCI caching operations

// GetAmortizableCI retrieves an amortizable CI from cache
func (c *CacheRepository) GetAmortizableCI(ctx context.Context, ciID uuid.UUID) (*AmortizableCI, error) {
	key := c.buildKey(CIKeyPattern, ciID.String())

	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		c.logger.Error().Err(err).Str("key", key).Msg("Failed to get amortizable CI from cache")
		return nil, fmt.Errorf("failed to get amortizable CI from cache: %w", err)
	}

	var ci AmortizableCI
	if err := json.Unmarshal([]byte(data), &ci); err != nil {
		c.logger.Error().Err(err).Str("key", key).Msg("Failed to unmarshal amortizable CI")
		return nil, fmt.Errorf("failed to unmarshal amortizable CI: %w", err)
	}

	c.logger.Info().Str("key", key).Msg("Cache HIT for amortizable CI")
	return &ci, nil
}

// SetAmortizableCI stores an amortizable CI in cache
func (c *CacheRepository) SetAmortizableCI(ctx context.Context, ciID uuid.UUID, ci *AmortizableCI, ttl time.Duration) error {
	if ttl == 0 {
		ttl = DefaultTTL
	}

	key := c.buildKey(CIKeyPattern, ciID.String())
	data, err := json.Marshal(ci)
	if err != nil {
		c.logger.Error().Err(err).Str("key", key).Msg("Failed to marshal amortizable CI")
		return fmt.Errorf("failed to marshal amortizable CI: %w", err)
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		c.logger.Error().Err(err).Str("key", key).Msg("Failed to set amortizable CI in cache")
		return fmt.Errorf("failed to set amortizable CI in cache: %w", err)
	}

	c.logger.Info().Str("key", key).Dur("ttl", ttl).Msg("Cache SET for amortizable CI")
	return nil
}

// InvalidateAmortizableCI invalidates amortizable CI cache entries
func (c *CacheRepository) InvalidateAmortizableCI(ctx context.Context, ciID uuid.UUID) error {
	ciKey := c.buildKey(CIKeyPattern, ciID.String())
	ledgerKey := c.buildKey(CILatestLedgerKey, ciID.String())

	// Delete both CI and its latest ledger entry
	if err := c.client.Del(ctx, ciKey, ledgerKey).Err(); err != nil {
		c.logger.Error().Err(err).Str("ci_key", ciKey).Msg("Failed to invalidate amortizable CI cache")
		return fmt.Errorf("failed to invalidate amortizable CI cache: %w", err)
	}

	c.logger.Info().Str("ci_key", ciKey).Msg("Cache INVALIDATE for amortizable CI")
	return nil
}

// Minimal implementations for required interface methods
func (c *CacheRepository) GetCILatestLedgerEntry(ctx context.Context, ciID uuid.UUID) (*LedgerEntry, error) {
	// Stub implementation
	return nil, nil
}

func (c *CacheRepository) SetCILatestLedgerEntry(ctx context.Context, ciID uuid.UUID, entry *LedgerEntry, ttl time.Duration) error {
	// Stub implementation
	return nil
}

func (c *CacheRepository) InvalidateCILedgerEntries(ctx context.Context, ciID uuid.UUID) error {
	// Stub implementation
	return nil
}

func (c *CacheRepository) GetAmortizationBehaviorForStatus(ctx context.Context, statusID uuid.UUID) (string, error) {
	// Stub implementation
	return "", nil
}

func (c *CacheRepository) SetAmortizationBehaviorForStatus(ctx context.Context, statusID uuid.UUID, behavior string) error {
	// Stub implementation
	return nil
}

func (c *CacheRepository) GetIsAmortizableForCIType(ctx context.Context, ciTypeID uuid.UUID) (bool, error) {
	// Stub implementation
	return false, nil
}

func (c *CacheRepository) SetIsAmortizableForCIType(ctx context.Context, ciTypeID uuid.UUID, isAmortizable bool) error {
	// Stub implementation
	return nil
}

func (c *CacheRepository) AcquireProcessingLock(ctx context.Context, lockKey string, ttl time.Duration) (bool, error) {
	// Stub implementation
	return true, nil
}

func (c *CacheRepository) ReleaseProcessingLock(ctx context.Context, lockKey string) error {
	// Stub implementation
	return nil
}

func (c *CacheRepository) EnqueueAmortizationJob(ctx context.Context, job *AmortizationJob) error {
	// Stub implementation
	return nil
}

func (c *CacheRepository) DequeueAmortizationJob(ctx context.Context) (*AmortizationJob, error) {
	// Stub implementation
	return nil, nil
}

func (c *CacheRepository) GetJobQueueStats(ctx context.Context) (*JobQueueStats, error) {
	// Stub implementation
	return &JobQueueStats{}, nil
}

func (c *CacheRepository) GetAmortizationSummaries(ctx context.Context, groupBy string) (*AmortizationSummary, error) {
	// Stub implementation
	return nil, nil
}

func (c *CacheRepository) SetAmortizationSummaries(ctx context.Context, groupBy string, summary *AmortizationSummary) error {
	// Stub implementation
	return nil
}

func (c *CacheRepository) CacheUserFilters(ctx context.Context, userID uuid.UUID, filters *AmortizableCIFilters) error {
	// Stub implementation
	return nil
}

func (c *CacheRepository) GetUserFilters(ctx context.Context, userID uuid.UUID) (*AmortizableCIFilters, error) {
	// Stub implementation
	return nil, nil
}

func (c *CacheRepository) InvalidatePattern(ctx context.Context, pattern string) error {
	// Stub implementation
	return nil
}

func (c *CacheRepository) WarmupCache(ctx context.Context) error {
	// Stub implementation
	return nil
}

func (c *CacheRepository) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	// Stub implementation
	return map[string]interface{}{}, nil
}

// Helper methods

// buildKey constructs a cache key with the service prefix
func (c *CacheRepository) buildKey(pattern string, args ...interface{}) string {
	if len(args) == 0 {
		return c.keyPrefix + pattern
	}
	return c.keyPrefix + fmt.Sprintf(pattern, args...)
}