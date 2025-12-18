# IT Asset Amortization Module - Database Design

## Overview

This document describes the comprehensive database schema and design decisions for the IT Asset Amortization Module integrated with the Pustaka CMDB system. The module provides automated depreciation tracking, financial audit trails, and status-aware amortization calculations.

## Architecture Summary

### Core Components

1. **Enhanced CI Types** - Amortizable flag on `ci_type_definitions`
2. **Status-Aware Processing** - Amortization behavior on `lifecycle_statuses`
3. **Financial CI Data** - Extended `configuration_items` with financial columns
4. **Immutable Ledger** - `amortization_ledger` for complete audit trail
5. **Scheduler Tracking** - `amortization_runs` for batch execution monitoring
6. **Reporting Optimizations** - `amortization_summaries` and materialized views

### Design Principles

- **Financial Integrity**: Never delete financial records - use adjustment entries
- **Audit Compliance**: Complete audit trail with user attribution and timestamps
- **Performance**: Optimized indexes and summary tables for reporting
- **Scalability**: Designed for millions of assets and daily processing
- **Recovery**: Checkpoint mechanism for scheduler restart capability

## Database Schema

### 1. Enhanced CI Type Definitions

```sql
ALTER TABLE ci_type_definitions
    ADD COLUMN is_amortizable BOOLEAN DEFAULT false NOT NULL;
```

**Purpose**: Marks which CI types can participate in amortization
**Default**: `false` (opt-in approach)
**Index**: `idx_ci_types_amortizable` for fast filtering

### 2. Enhanced Lifecycle Statuses

```sql
ALTER TABLE lifecycle_statuses
    ADD COLUMN amortization_behavior VARCHAR(20) DEFAULT 'pending' NOT NULL
    CONSTRAINT valid_amortization_behavior
    CHECK (amortization_behavior IN ('pending', 'active', 'terminal'));
```

**Behavior Types**:
- **`pending`**: Asset not yet in amortization (planned, on_order, in_stock)
- **`active`**: Asset currently being amortized (operational, maintenance)
- **`terminal`**: Amortization stops (retired, disposed, missing)

**Index**: `idx_lifecycle_amortization_behavior` for behavior-based queries

### 3. Enhanced Configuration Items

```sql
ALTER TABLE configuration_items
    ADD COLUMN purchase_cost DECIMAL(15,2) NULL,
    ADD COLUMN salvage_value DECIMAL(15,2) NULL,
    ADD COLUMN amort_start_date DATE NULL,
    ADD COLUMN useful_life_months INTEGER NULL,
    ADD COLUMN current_book_value DECIMAL(15,2) NULL;
```

**Financial Fields**:
- **`purchase_cost`**: Initial acquisition cost (nullable for backward compatibility)
- **`salvage_value`**: Expected value at end of useful life
- **`amort_start_date`**: Date when amortization begins
- **`useful_life_months`**: Total months for depreciation
- **`current_book_value`**: Real-time book value (calculated and maintained)

**Constraints**:
- Non-negative costs and values
- Salvage value ≤ purchase cost
- Positive useful life months
- Date consistency validation

**Indexes**:
- Individual indexes on financial columns
- Composite index for active amortization queries

### 4. Amortization Ledger (Append-Only)

```sql
CREATE TABLE amortization_ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ci_id UUID NOT NULL REFERENCES configuration_items(id) ON DELETE CASCADE,
    entry_date DATE NOT NULL,
    entry_type VARCHAR(20) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    book_value_before DECIMAL(15,2) NOT NULL,
    book_value_after DECIMAL(15,2) NOT NULL,
    period_start_date DATE,
    period_end_date DATE,
    days_in_period INTEGER,
    adjustment_reason TEXT,
    adjustment_reference VARCHAR(100),
    corrects_entry_id UUID REFERENCES amortization_ledger(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    is_system_generated BOOLEAN DEFAULT false,
    batch_run_id UUID REFERENCES amortization_runs(id),
    sequence_number INTEGER NOT NULL
);
```

**Entry Types**:
- **`depreciation`**: Regular monthly depreciation
- **`write_off`**: Asset disposal or terminal status
- **`adjustment`**: Manual value corrections
- **`correction`**: Error corrections with audit trail

**Key Features**:
- **Immutable**: Records never deleted
- **Complete Audit**: Before/after values always stored
- **Adjustment Tracking**: References to corrected entries
- **Batch Association**: Links to scheduler runs
- **Sequencing**: Multiple entries per day supported

**Indexes**:
- CI and date queries (most common)
- Entry type filtering
- Batch run association
- Correction tracking

### 5. Amortization Runs (Scheduler Tracking)

```sql
CREATE TABLE amortization_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_date DATE NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    total_cis_processed INTEGER DEFAULT 0,
    successful_depreciations INTEGER DEFAULT 0,
    write_offs_generated INTEGER DEFAULT 0,
    errors_encountered INTEGER DEFAULT 0,
    last_processed_ci_id UUID,
    last_processed_at TIMESTAMP WITH TIME ZONE,
    checkpoint_data JSONB DEFAULT '{}',
    error_summary TEXT,
    error_details JSONB DEFAULT '{}',
    run_config JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id)
);
```

**Status Flow**:
1. **`pending`**: Scheduled but not started
2. **`running`**: Currently processing
3. **`completed`**: Successful completion
4. **`failed`**: Error during processing
5. **`cancelled`**: Manual cancellation

**Recovery Features**:
- **Checkpoint Data**: JSON state for restart capability
- **Progress Tracking**: Last processed CI for resume
- **Error Details**: Comprehensive error logging
- **Configuration Snapshot**: Run parameters preserved

**Indexes**:
- Status-based queries for monitoring
- Date-based queries for run history
- Performance metrics analysis

### 6. Amortization Summaries (Reporting Optimization)

```sql
CREATE TABLE amortization_summaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ci_id UUID NOT NULL REFERENCES configuration_items(id) ON DELETE CASCADE,
    reporting_date DATE NOT NULL,
    current_book_value DECIMAL(15,2) NOT NULL,
    accumulated_depreciation DECIMAL(15,2) NOT NULL,
    period_depreciation DECIMAL(15,2) DEFAULT 0,
    period_adjustments DECIMAL(15,2) DEFAULT 0,
    amortization_status VARCHAR(20) NOT NULL,
    last_updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (ci_id, reporting_date)
);
```

**Purpose**: Pre-calculated summaries for fast reporting
**Update Strategy**: Trigger-based real-time updates
**Status Types**: `pending`, `active`, `completed`, `written_off`
**Indexes**: Optimized for date-based and CI-based queries

## Performance Optimization Strategy

### 1. Indexing Strategy

#### Primary Access Patterns

**Active Asset Queries**:
```sql
-- Find all assets currently being amortized
CREATE INDEX idx_cis_amortization_active ON configuration_items(
    amort_start_date, useful_life_months, current_book_value
) WHERE amort_start_date IS NOT NULL
AND useful_life_months IS NOT NULL
AND current_book_value > 0;
```

**Financial Reporting**:
```sql
-- Support range queries on financial data
CREATE INDEX idx_cis_purchase_cost ON configuration_items(purchase_cost)
WHERE purchase_cost IS NOT NULL;
CREATE INDEX idx_ledger_ci_date_type ON amortization_ledger(ci_id, entry_date, entry_type);
```

**Audit Trail Queries**:
```sql
-- Fast lookup of CI's complete amortization history
CREATE INDEX idx_ledger_ci_id ON amortization_ledger(ci_id);
CREATE INDEX idx_ledger_entry_date ON amortization_ledger(entry_date);
```

#### Specialized Indexes

**Partial Indexes**:
- Only index non-NULL financial values
- Status-based filtering for lifecycle queries
- Batch run association for system entries

**Composite Indexes**:
- Multi-column indexes for common query patterns
- Ordering columns for report generation

### 2. Query Optimization

#### Materialized Views

**Active Assets View**:
```sql
CREATE VIEW active_amortization_assets AS
SELECT
    ci.*, calculated_fields...
FROM configuration_items ci
JOIN ci_type_definitions ctd ON ci.ci_type = ctd.name
LEFT JOIN lifecycle_statuses ls ON ci.lifecycle_status_id = ls.id
WHERE ctd.is_amortizable = true
AND ci.current_book_value > 0;
```

**Financial Reporting View**:
```sql
CREATE VIEW amortization_report AS
SELECT
    ci.*, al.*, ytd_calculations...
FROM configuration_items ci
JOIN amortization_summaries al ON ci.id = al.ci_id
WHERE al.reporting_date = (latest_date_subquery);
```

#### Partitioning Strategy (Future Enhancement)

**Ledger Partitioning by Year**:
- Improves query performance for historical data
- Simplifies archival processes
- Reduces index size for active data

## Data Validation and Constraints

### 1. Business Rules

#### Financial Data Integrity
```sql
-- Cost validations
ALTER TABLE configuration_items
    ADD CONSTRAINT valid_purchase_cost CHECK (purchase_cost IS NULL OR purchase_cost >= 0),
    ADD CONSTRAINT valid_salvage_value CHECK (salvage_value IS NULL OR salvage_value >= 0),
    ADD CONSTRAINT valid_salvage_vs_purchase CHECK (
        salvage_value IS NULL OR purchase_cost IS NULL OR salvage_value <= purchase_cost
    );

-- Time-based validations
ALTER TABLE configuration_items
    ADD CONSTRAINT valid_useful_life_months CHECK (useful_life_months IS NULL OR useful_life_months > 0),
    ADD CONSTRAINT amortization_dates_consistency CHECK (
        amort_start_date IS NULL OR
        useful_life_months IS NULL OR
        (purchase_cost IS NOT NULL AND useful_life_months IS NOT NULL)
    );
```

#### Ledger Consistency
```sql
-- Prevent duplicate entries on same date
CONSTRAINT unique_ledger_entry UNIQUE (ci_id, entry_date, sequence_number);

-- Ensure valid entry types
CONSTRAINT valid_entry_type CHECK (entry_type IN ('depreciation', 'write_off', 'adjustment', 'correction'));
```

### 2. Trigger-Based Validation

#### Book Value Maintenance
```sql
CREATE OR REPLACE FUNCTION update_ci_book_value()
RETURNS TRIGGER AS $$
BEGIN
    -- Set initial book value for new assets
    IF NEW.purchase_cost IS NOT NULL AND NEW.amort_start_date IS NULL THEN
        NEW.current_book_value = NEW.purchase_cost;
    ELSIF NEW.purchase_cost IS NULL THEN
        NEW.current_book_value = NULL;
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';
```

#### Sequence Number Management
```sql
CREATE OR REPLACE FUNCTION set_ledger_sequence_number()
RETURNS TRIGGER AS $$
BEGIN
    -- Get next sequence for CI and date
    NEW.sequence_number = (
        SELECT COALESCE(MAX(sequence_number), 0) + 1
        FROM amortization_ledger
        WHERE ci_id = NEW.ci_id AND entry_date = NEW.entry_date
    );
    RETURN NEW;
END;
$$ language 'plpgsql';
```

### 3. Application-Level Validation

#### Status Transitions
- Prevent amortization start for `pending` lifecycle statuses
- Auto-generate write-offs for `terminal` status transitions
- Validate useful life changes (prospective only)

#### Financial Calculations
- Straight-line depreciation: `(purchase_cost - salvage_value) / useful_life_months`
- Monthly depreciation rounding to 2 decimal places
- Minimum book value never goes below salvage value

## Security and Access Control

### 1. RBAC Integration

#### New Permissions
```sql
INSERT INTO permissions (name, description, resource_type) VALUES
    ('amortization:read', 'Read amortization data and reports', 'amortization'),
    ('amortization:configure', 'Configure amortization settings', 'amortization'),
    ('amortization:adjust', 'Make manual adjustments', 'amortization'),
    ('amortization:run', 'Trigger amortization calculations', 'amortization'),
    ('amortization:admin', 'Full amortization administration', 'amortization');
```

#### Role Assignments
- **Admin**: All amortization permissions
- **Editor**: Read, configure, adjust permissions
- **Viewer**: Read-only access

### 2. Data Privacy

#### Sensitive Financial Data
- All financial columns access-controlled via RBAC
- Audit logging for all financial modifications
- Row-level security for multi-tenant scenarios

#### Audit Trail
- Complete user attribution for all changes
- IP address and user agent tracking
- Change reasons for adjustments and corrections

## Backup and Recovery Strategy

### 1. Financial Data Protection

#### High-Priority Tables
1. **amortization_ledger** - Critical financial records
2. **configuration_items** (financial columns) - Asset values
3. **amortization_runs** - Execution history
4. **amortization_summaries** - Can be regenerated from ledger

#### Backup Requirements
- **Daily Incremental**: Ledger changes (append-only, small)
- **Weekly Full**: All amortization tables
- **Point-in-Time Recovery**: Must support financial audit requirements
- **Cross-Region Replication**: For disaster recovery

### 2. Recovery Procedures

#### Ledger Integrity
```sql
-- Verify ledger balance matches CI book values
SELECT ci.id, ci.name, ci.current_book_value, latest_ledger.book_value_after
FROM configuration_items ci
LEFT JOIN LATERAL (
    SELECT book_value_after
    FROM amortization_ledger
    WHERE ci_id = ci.id
    ORDER BY entry_date DESC, sequence_number DESC
    LIMIT 1
) latest_ledger ON true
WHERE ci.current_book_value IS DISTINCT FROM latest_ledger.book_value_after;
```

#### Summary Regeneration
```sql
-- Function to regenerate summaries from ledger
CREATE OR REPLACE FUNCTION regenerate_amortization_summaries(p_ci_id UUID)
RETURNS VOID AS $$
BEGIN
    DELETE FROM amortization_summaries WHERE ci_id = p_ci_id;

    INSERT INTO amortization_summaries (ci_id, reporting_date, ...)
    SELECT
        ci_id,
        entry_date,
        book_value_after,
        accumulated_depreciation,
        amortization_status
    FROM amortization_ledger_processed_view
    WHERE ci_id = p_ci_id;
END;
$$ language 'plpgsql';
```

### 3. Archival Strategy

#### Historical Data Management
- **7 Years**: Full retention for audit compliance
- **7+ Years**: Archive to cold storage
- **Ledger Partitioning**: By year for efficient archival

#### Archival Process
1. Verify data integrity before archival
2. Create compressed archives
3. Update partition metadata
4. Maintain search indexes for archived periods

## Integration Points

### 1. Existing CMDB Integration

#### CI Type Management
- Extend existing CI type forms with amortization checkbox
- Validate financial fields based on amortizable flag
- Maintain backward compatibility for non-amortizable types

#### Lifecycle Status Integration
- Automatic amortization behavior based on status
- Status change triggers for write-offs
- Integration with existing status workflows

#### Audit System Integration
- Financial changes logged to existing `audit_logs` table
- Consistent user attribution and timing
- Integration with existing audit reporting

### 2. External System Integration

#### ERP/Accounting System
- CSV export for general ledger integration
- Periodic reconciliation reports
- API endpoints for real-time sync

#### Procurement System
- Asset acquisition data import
- Purchase cost and date synchronization
- Vendor information cross-reference

## Migration Strategy

### 1. Zero-Downtime Migration

#### Phase 1: Schema Changes
- Add new columns with DEFAULT values
- Create indexes in CONCURRENTLY mode
- Deploy triggers and functions

#### Phase 2: Data Migration
- Backfill existing CI data if needed
- Populate amortization behavior for lifecycle statuses
- Create initial amortization summaries

#### Phase 3: Application Rollout
- Enable amortization features
- Train users on new financial fields
- Monitor system performance

### 2. Rollback Planning

#### Schema Rollback
- DROP new columns, tables, and indexes
- Remove triggers and functions
- Restore previous application code

#### Data Recovery
- Maintain pre-migration backups
- Document rollback procedures
- Test rollback scenarios

## Monitoring and Maintenance

### 1. Performance Monitoring

#### Key Metrics
- Ledger entry processing rate
- Query performance for reports
- Index usage and efficiency
- Database size growth rates

#### Alerting
- Failed amortization runs
- Long-running queries
- Index bloat detection
- Storage capacity warnings

### 2. Data Quality Monitoring

#### Consistency Checks
```sql
-- Daily integrity check
SELECT
    'Ledger Balance Mismatch' as issue,
    COUNT(*) as count
FROM (
    SELECT ci.id
    FROM configuration_items ci
    LEFT JOIN amortization_ledger al ON ci.id = al.ci_id
    WHERE ci.current_book_value IS NOT NULL
    GROUP BY ci.id, ci.current_book_value
    HAVING ci.current_book_value != MAX(al.book_value_after)
) mismatches;
```

#### Automated Validation
- Negative book value detection
- Missing amortization schedules
- Orphaned ledger entries
- Status consistency verification

### 3. Maintenance Procedures

#### Regular Tasks
- Index statistics updates
- Table vacuum and analyze
- Partition maintenance
- Archive cleanup

#### Performance Tuning
- Query plan analysis
- Index optimization
- Configuration parameter tuning
- Hardware capacity planning

## Future Enhancements

### 1. Advanced Features

#### Multi-Currency Support
- Currency conversion tables
- Historical exchange rates
- Reporting in multiple currencies

#### Enhanced Depreciation Methods
- Declining balance depreciation
- Sum-of-years-digits
- Units of production method

#### Asset Revaluation
- Fair value adjustments
- Impairment testing
- Revaluation surplus tracking

### 2. Scaling Improvements

#### Partitioning Implementation
- Ledger partitioning by year/month
- Automated partition management
- Query optimization across partitions

#### Caching Strategy
- Redis caching for summary data
- Application-level result caching
- Database query result caching

#### Read Replicas
- Reporting query offloading
- Geographic distribution
- Load balancing for read operations

## Conclusion

The IT Asset Amortization Module provides a robust, scalable, and auditable solution for tracking IT asset depreciation within the Pustaka CMDB system. The design emphasizes data integrity, performance optimization, and financial compliance while maintaining seamless integration with existing CMDB functionality.

Key strengths of the design:
- **Financial Integrity**: Immutable ledger with complete audit trail
- **Performance**: Optimized indexes and summary tables
- **Scalability**: Designed for millions of assets
- **Recovery**: Checkpoint mechanism and rollback procedures
- **Compliance**: Full audit capabilities and data validation
- **Integration**: Seamless CMDB integration with RBAC

The modular design allows for future enhancements while maintaining backward compatibility and operational stability.