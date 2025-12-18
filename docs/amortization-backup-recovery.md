# IT Asset Amortization Module - Backup and Recovery Strategy

## Overview

This document outlines comprehensive backup and recovery procedures specifically for the IT Asset Amortization Module, ensuring financial data integrity, regulatory compliance, and business continuity for critical asset depreciation tracking.

## Data Classification and Priorities

### 1. Critical Financial Data (Highest Priority)

#### Amortization Ledger
- **Classification**: Critical Financial Records
- **Retention**: 7 years minimum (audit requirement)
- **RPO**: 15 minutes
- **RTO**: 1 hour
- **Backup Frequency**: Real-time streaming + daily snapshots

#### Configuration Items (Financial Columns)
- **Classification**: Critical Asset Data
- **Retention**: 7 years minimum
- **RPO**: 1 hour
- **RTO**: 2 hours
- **Backup Frequency**: Hourly incremental + daily full

#### Amortization Runs
- **Classification**: Critical Operational Data
- **Retention**: 3 years
- **RPO**: 15 minutes
- **RTO**: 1 hour
- **Backup Frequency**: Real-time streaming

### 2. Important Supporting Data

#### Amortization Summaries
- **Classification**: Important (Regenerable)
- **Retention**: 1 year
- **RPO**: 4 hours
- **RTO**: 6 hours
- **Backup Frequency**: Daily
- **Note**: Can be regenerated from ledger

#### CI Type Definitions (Amortizable Flag)
- **Classification**: Important Configuration
- **Retention**: 7 years
- **RPO**: 24 hours
- **RTO**: 4 hours
- **Backup Frequency**: Daily

### 3. Reference Data

#### Lifecycle Statuses (Amortization Behavior)
- **Classification**: Standard Reference
- **Retention**: 7 years
- **RPO**: 24 hours
- **RTO**: 4 hours
- **Backup Frequency**: Weekly

## Backup Architecture

### 1. Primary Backup Strategy

#### Streaming Replication (Real-time)
```yaml
# PostgreSQL Streaming Replication Configuration
postgres:
  primary:
    wal_level: replica
    max_wal_senders: 3
    wal_keep_size: 16GB
    archive_mode: on
    archive_command: 'wal-g wal-push %p'

  replica:
    hot_standby: on
    max_standby_streaming_delay: 30s
    recovery_target_timeline: 'latest'
```

#### Continuous WAL Archiving
```bash
# WAL-G Configuration for Continuous Backup
export WALG_S3_PREFIX="s3://pustaka-backups/amortization/wal"
export WALG_S3_STORAGE_CLASS="STANDARD_IA"
export WALG_DELTA_MAX_STEPS="8"
export WALG_COMPRESSION_METHOD="lz4"

# Continuous WAL backup
wal-g wal-push /var/lib/postgresql/wal/000000010000000000000001
```

#### Daily Logical Backups
```bash
#!/bin/bash
# Daily Amortization Module Backup Script
BACKUP_DATE=$(date +%Y%m%d)
BACKUP_DIR="/backups/amortization/daily/${BACKUP_DATE}"

# Create backup directory
mkdir -p "${BACKUP_DIR}"

# Export amortization tables
pg_dump -h localhost -U pustaka -d pustaka \
  --schema=public \
  --table=amortization_ledger \
  --table=configuration_items \
  --table=amortization_runs \
  --table=amortization_summaries \
  --table=ci_type_definitions \
  --table=lifecycle_statuses \
  --format=custom \
  --compress=9 \
  --file="${BACKUP_DIR}/amortization_${BACKUP_DATE}.dump"

# Verify backup integrity
pg_restore --list "${BACKUP_DIR}/amortization_${BACKUP_DATE}.dump" > "${BACKUP_DIR}/amortization_${BACKUP_DATE}.toc"

# Calculate checksums
sha256sum "${BACKUP_DIR}/amortization_${BACKUP_DATE}.dump" > "${BACKUP_DIR}/amortization_${BACKUP_DATE}.sha256"

# Upload to cloud storage
aws s3 cp "${BACKUP_DIR}/" s3://pustaka-backups/amortization/daily/${BACKUP_DATE}/ --recursive
```

### 2. Point-in-Time Recovery (PITR) Configuration

#### Recovery Timeline Strategy
```sql
-- Create recovery points for critical business events
CREATE TABLE recovery_points (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    point_name VARCHAR(100) NOT NULL,
    recovery_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id)
);

-- Example recovery points
INSERT INTO recovery_points (point_name, recovery_timestamp, description) VALUES
('pre_amortization_migration', NOW() - INTERVAL '1 day', 'State before amortization module deployment'),
('quarter_end_q4_2024', '2024-12-31 23:59:59 UTC', 'End of Q4 2024 for financial reporting'),
('year_end_2024', '2024-12-31 23:59:59 UTC', 'End of Year 2024 for annual audit');
```

#### Automated Recovery Point Creation
```bash
#!/bin/bash
# Create named recovery points for financial reporting
create_recovery_point() {
    local point_name=$1
    local description=$2

    psql -h localhost -U pustaka -d pustaka << EOF
INSERT INTO recovery_points (point_name, recovery_timestamp, description)
VALUES ('${point_name}', NOW(), '${description}')
RETURNING id, recovery_timestamp;
EOF
}

# Monthly recovery points (end of month)
create_recovery_point "monthly_$(date +%Y%m)" "End of month financial state"

# Quarterly recovery points
if [[ $(date +%m) =~ ^(03|06|09|12)$ ]]; then
    create_recovery_point "quarterly_q$(date +%q)_$(date +%Y)" "End of Q$(date +%q) $(date +%Y)"
fi
```

## Disaster Recovery Procedures

### 1. Complete System Recovery

#### Recovery Time Objective: 4 Hours
```bash
#!/bin/bash
# Complete System Recovery Script
RECOVERY_TIMESTAMP=$1
BACKUP_DATE=$(date +%Y%m%d)

echo "Starting complete system recovery to: ${RECOVERY_TIMESTAMP}"

# Step 1: Stop application services
docker-compose down

# Step 2: Clear existing data directories
rm -rf /var/lib/postgresql/data/*

# Step 3: Restore base backup
echo "Restoring base backup..."
pg_restore -h localhost -U pustaka -d pustaka \
  --clean --if-exists --verbose \
  "/backups/amortization/daily/${BACKUP_DATE}/amortization_${BACKUP_DATE}.dump"

# Step 4: Apply WAL archives for PITR
echo "Applying WAL archives for point-in-time recovery..."
wal-g wal-fetch "${RECOVERY_TIMESTAMP}" /tmp/recovery.tar.gz
tar -xzf /tmp/recovery.tar.gz -C /var/lib/postgresql/data/

# Step 5: Configure recovery
cat > /var/lib/postgresql/data/recovery.conf << EOF
restore_command = 'wal-g wal-fetch %f %p'
recovery_target_time = '${RECOVERY_TIMESTAMP}'
recovery_target_action = 'promote'
EOF

# Step 6: Start database with recovery mode
docker-compose up -d postgres

# Step 7: Monitor recovery progress
while ! pg_isready -h localhost -p 5432 -U pustaka; do
    echo "Waiting for database to become ready..."
    sleep 10
done

# Step 8: Verify data integrity
echo "Verifying data integrity..."
psql -h localhost -U pustaka -d pustaka << EOF
-- Verify ledger integrity
SELECT COUNT(*) as total_ledger_entries FROM amortization_ledger;
SELECT COUNT(*) as total_cis_with_financials
FROM configuration_items
WHERE purchase_cost IS NOT NULL;

-- Verify no data gaps
SELECT
    MIN(entry_date) as earliest_entry,
    MAX(entry_date) as latest_entry,
    COUNT(DISTINCT DATE(entry_date)) as unique_dates
FROM amortization_ledger;
EOF

# Step 9: Start remaining services
docker-compose up -d

echo "System recovery completed successfully"
```

### 2. Partial Recovery Scenarios

#### Scenario A: Corrupted Ledger Entries
```bash
#!/bin/bash
# Recover corrupted ledger entries for specific CI
CI_ID=$1
CORRUPTION_DATE=$2

echo "Recovering ledger entries for CI: ${CI_ID} from ${CORRUPTION_DATE}"

# Step 1: Create backup of current state
pg_dump -h localhost -U pustaka -d pustaka \
  --data-only --table=amortization_ledger \
  --where="ci_id = '${CI_ID}'" \
  > "/tmp/ci_${CI_ID}_backup_$(date +%Y%m%d_%H%M%S).sql"

# Step 2: Remove corrupted entries
psql -h localhost -U pustaka -d pustaka << EOF
DELETE FROM amortization_ledger
WHERE ci_id = '${CI_ID}'
AND entry_date >= '${CORRUPTION_DATE}';

-- Regenerate summaries
DELETE FROM amortization_summaries
WHERE ci_id = '${CI_ID}'
AND reporting_date >= '${CORRUPTION_DATE}';
EOF

# Step 3: Recalculate current book value
psql -h localhost -U pustaka -d pustaka << EOF
-- Update CI current book value from last valid ledger entry
UPDATE configuration_items
SET current_book_value = (
    SELECT book_value_after
    FROM amortization_ledger
    WHERE ci_id = '${CI_ID}'
    AND entry_date < '${CORRUPTION_DATE}'
    ORDER BY entry_date DESC, sequence_number DESC
    LIMIT 1
)
WHERE id = '${CI_ID}';
EOF

# Step 4: Recalculate missing amortization
echo "Recalculating missing amortization..."
# This would call the backend service to regenerate amortization
curl -X POST "http://localhost:8080/api/v1/amortization/recalculate/${CI_ID}" \
  -H "Authorization: Bearer ${API_TOKEN}"

echo "Ledger recovery completed for CI: ${CI_ID}"
```

#### Scenario B: Missing Summary Data
```sql
-- Function to regenerate amortization summaries from ledger
CREATE OR REPLACE FUNCTION regenerate_ci_summaries(p_ci_id UUID)
RETURNS VOID AS $$
DECLARE
    current_book_value DECIMAL(15,2);
    accumulated_depreciation DECIMAL(15,2);
BEGIN
    -- Clear existing summaries
    DELETE FROM amortization_summaries WHERE ci_id = p_ci_id;

    -- Get initial book value
    SELECT purchase_cost INTO current_book_value
    FROM configuration_items
    WHERE id = p_ci_id;

    accumulated_depreciation := 0;

    -- Process ledger entries chronologically
    FOR entry IN
        SELECT entry_date, entry_type, amount, book_value_after
        FROM amortization_ledger
        WHERE ci_id = p_ci_id
        ORDER BY entry_date, sequence_number
    LOOP
        IF entry.entry_type = 'depreciation' THEN
            accumulated_depreciation := accumulated_depreciation + entry.amount;
        END IF;

        current_book_value := entry.book_value_after;

        -- Insert summary record
        INSERT INTO amortization_summaries (
            ci_id, reporting_date, current_book_value,
            accumulated_depreciation, amortization_status
        ) VALUES (
            p_ci_id, entry.entry_date, current_book_value,
            accumulated_depreciation,
            CASE
                WHEN entry.entry_type = 'write_off' THEN 'written_off'
                WHEN current_book_value <= 0 THEN 'completed'
                ELSE 'active'
            END
        );
    END LOOP;

    -- Log the regeneration
    INSERT INTO audit_logs (
        entity_type, entity_id, action, performed_by, details
    ) VALUES (
        'amortization_summary', p_ci_id, 'regenerated',
        current_setting('app.current_user_id')::UUID,
        json_build_object('regenerated_at', NOW(), 'entries_processed',
            (SELECT COUNT(*) FROM amortization_ledger WHERE ci_id = p_ci_id))
    );
END;
$$ LANGUAGE plpgsql;
```

## Data Integrity Verification

### 1. Automated Integrity Checks

#### Daily Integrity Verification
```sql
-- Comprehensive daily integrity check function
CREATE OR REPLACE FUNCTION verify_amortization_integrity()
RETURNS TABLE(
    check_name TEXT,
    status TEXT,
    affected_records BIGINT,
    details TEXT
) AS $$
BEGIN
    -- Check 1: Ledger balance consistency
    RETURN QUERY
    WITH ledger_latest AS (
        SELECT DISTINCT ON (ci_id)
            ci_id,
            book_value_after as ledger_book_value
        FROM amortization_ledger
        ORDER BY ci_id, entry_date DESC, sequence_number DESC
    )
    SELECT
        'Ledger Balance Consistency'::TEXT,
        CASE WHEN COUNT(*) = 0 THEN 'PASS' ELSE 'FAIL' END::TEXT,
        COUNT(*)::BIGINT,
        'CI book values do not match latest ledger entries'::TEXT
    FROM configuration_items ci
    JOIN ledger_latest ll ON ci.id = ll.ci_id
    WHERE ci.current_book_value IS NOT NULL
    AND ci.current_book_value != ll.ledger_book_value;

    -- Check 2: No negative book values
    RETURN QUERY
    SELECT
        'No Negative Book Values'::TEXT,
        CASE WHEN MIN(book_value_after) >= 0 THEN 'PASS' ELSE 'FAIL' END::TEXT,
        COUNT(*)::BIGINT,
        'Negative book values found in ledger'::TEXT
    FROM amortization_ledger
    WHERE book_value_after < 0;

    -- Check 3: Sequential depreciation processing
    RETURN QUERY
    WITH depreciation_gaps AS (
        SELECT ci_id, entry_date,
            LAG(entry_date) OVER (PARTITION BY ci_id ORDER BY entry_date) as prev_date
        FROM amortization_ledger
        WHERE entry_type = 'depreciation'
    )
    SELECT
        'Sequential Depreciation'::TEXT,
        CASE WHEN COUNT(*) = 0 THEN 'PASS' ELSE 'FAIL' END::TEXT,
        COUNT(DISTINCT ci_id)::BIGINT,
        'Gaps found in monthly depreciation sequences'::TEXT
    FROM depreciation_gaps
    WHERE prev_date IS NOT NULL
    AND entry_date > prev_date + INTERVAL '1 month' + INTERVAL '7 days';

    -- Check 4: Write-off consistency
    RETURN QUERY
    SELECT
        'Write-off Consistency'::TEXT,
        CASE WHEN COUNT(*) = 0 THEN 'PASS' ELSE 'FAIL' END::TEXT,
        COUNT(*)::BIGINT,
        'Assets with write-off entries still have positive book value'::TEXT
    FROM amortization_ledger al1
    JOIN configuration_items ci ON al1.ci_id = ci.id
    WHERE al1.entry_type = 'write-off'
    AND ci.current_book_value > 0;

    -- Check 5: Summary data freshness
    RETURN QUERY
    SELECT
        'Summary Data Freshness'::TEXT,
        CASE WHEN MAX(reporting_date) >= CURRENT_DATE - INTERVAL '2 days'
             THEN 'PASS' ELSE 'FAIL' END::TEXT,
        COUNT(DISTINCT ci_id)::BIGINT,
        'Amortization summaries are not up to date'::TEXT
    FROM amortization_summaries;
END;
$$ LANGUAGE plpgsql;
```

#### Daily Integrity Check Schedule
```bash
#!/bin/bash
# Daily integrity check script
INTEGRITY_LOG="/var/log/amortization/integrity_check_$(date +%Y%m%d).log"

echo "Starting daily amortization integrity check at $(date)" >> ${INTEGRITY_LOG}

# Run integrity checks
psql -h localhost -U pustuka -d pustaka -c "
SELECT * FROM verify_amortization_integrity();
" >> ${INTEGRITY_LOG} 2>&1

# Check for failures
FAILED_CHECKS=$(psql -h localhost -U pustaka -d pustaka -t -c "
SELECT COUNT(*) FROM verify_amortization_integrity() WHERE status = 'FAIL';
")

if [ ${FAILED_CHECKS} -gt 0 ]; then
    echo "ALERT: ${FAILED_CHECKS} integrity checks failed" >> ${INTEGRITY_LOG}
    # Send alert to administrators
    curl -X POST "https://hooks.slack.com/your-webhook" \
      -H 'Content-type: application/json' \
      --data "{\"text\":\"Amortization integrity check failed with ${FAILED_CHECKS} failures\"}"
fi

echo "Daily integrity check completed at $(date)" >> ${INTEGRITY_LOG}
```

### 2. Manual Verification Procedures

#### Quarterly Financial Reconciliation
```sql
-- Quarterly reconciliation report
CREATE OR REPLACE FUNCTION quarterly_reconciliation(p_quarter_start DATE, p_quarter_end DATE)
RETURNS TABLE(
    ci_name TEXT,
    beginning_book_value DECIMAL(15,2),
    depreciation_amount DECIMAL(15,2),
    adjustments_amount DECIMAL(15,2),
    ending_book_value DECIMAL(15,2),
    status TEXT
) AS $$
BEGIN
    RETURN QUERY
    WITH quarterly_ledger AS (
        SELECT
            al.ci_id,
            ci.name,
            SUM(CASE WHEN al.entry_type = 'depreciation' THEN al.amount ELSE 0 END) as depreciation,
            SUM(CASE WHEN al.entry_type IN ('adjustment', 'correction') THEN al.amount ELSE 0 END) as adjustments,
            MIN(al.entry_date) as period_start,
            MAX(al.entry_date) as period_end
        FROM amortization_ledger al
        JOIN configuration_items ci ON al.ci_id = ci.id
        WHERE al.entry_date BETWEEN p_quarter_start AND p_quarter_end
        GROUP BY al.ci_id, ci.name
    ),
    beginning_values AS (
        SELECT
            ci.id as ci_id,
            ci.name,
            COALESCE(al.book_value_after, ci.purchase_cost) as beginning_value
        FROM configuration_items ci
        LEFT JOIN LATERAL (
            SELECT book_value_after
            FROM amortization_ledger
            WHERE ci_id = ci.id AND entry_date < p_quarter_start
            ORDER BY entry_date DESC, sequence_number DESC
            LIMIT 1
        ) al ON true
        WHERE ci.purchase_cost IS NOT NULL
    )
    SELECT
        bv.name,
        bv.beginning_value,
        COALESCE(ql.depreciation, 0),
        COALESCE(ql.adjustments, 0),
        ci.current_book_value,
        CASE
            WHEN ci.current_book_value = (bv.beginning_value - COALESCE(ql.depreciation, 0) - COALESCE(ql.adjustments, 0))
            THEN 'BALANCED'
            ELSE 'MISMATCH'
        END
    FROM beginning_values bv
    LEFT JOIN quarterly_ledger ql ON bv.ci_id = ql.ci_id
    JOIN configuration_items ci ON bv.ci_id = ci.id
    ORDER BY bv.name;
END;
$$ LANGUAGE plpgsql;
```

## Archival and Retention

### 1. Data Archival Strategy

#### Ledger Partitioning for Archival
```sql
-- Create partitioned amortization_ledger table for archival
CREATE TABLE amortization_ledger_partitioned (
    LIKE amortization_ledger INCLUDING ALL
) PARTITION BY RANGE (entry_date);

-- Create yearly partitions
CREATE TABLE amortization_ledger_2024 PARTITION OF amortization_ledger_partitioned
    FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');

CREATE TABLE amortization_ledger_2025 PARTITION OF amortization_ledger_partitioned
    FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');

-- Migration procedure for existing data
INSERT INTO amortization_ledger_partitioned SELECT * FROM amortization_ledger;
DROP TABLE amortization_ledger;
ALTER TABLE amortization_ledger_partitioned RENAME TO amortization_ledger;
```

#### Automated Archival Process
```bash
#!/bin/bash
# Annual archival process for old amortization data
ARCHIVAL_YEAR=$(date +%Y)
ARCHIVAL_CUTOFF_YEAR=$((${ARCHIVAL_YEAR} - 7))

echo "Starting archival process for data older than ${ARCHIVAL_CUTOFF_YEAR}"

# Step 1: Create archival backup
pg_dump -h localhost -U pustaka -d pustaka \
  --table=amortization_ledger \
  --where="entry_date < '${ARCHIVAL_CUTOFF_YEAR}-01-01'" \
  --format=custom \
  --compress=9 \
  --file="/backups/amortization/archival/ledger_before_${ARCHIVAL_CUTOFF_YEAR}.dump"

# Step 2: Export to cold storage
aws s3 cp "/backups/amortization/archival/ledger_before_${ARCHIVAL_CUTOFF_YEAR}.dump" \
  "s3://pustaka-archives/amortization/ledger_before_${ARCHIVAL_CUTOFF_YEAR}.dump" \
  --storage-class GLACIER

# Step 3: Create summary of archived data
psql -h localhost -U pustaka -d pustaka << EOF
CREATE TABLE amortization_ledger_archived_${ARCHIVAL_CUTOFF_YEAR} AS
SELECT
    ci_id,
    MIN(entry_date) as first_entry_date,
    MAX(entry_date) as last_entry_date,
    COUNT(*) as total_entries,
    SUM(CASE WHEN entry_type = 'depreciation' THEN amount ELSE 0 END) as total_depreciation,
    SUM(CASE WHEN entry_type = 'write_off' THEN amount ELSE 0 END) as total_write_offs,
    SUM(CASE WHEN entry_type IN ('adjustment', 'correction') THEN amount ELSE 0 END) as total_adjustments
FROM amortization_ledger
WHERE entry_date < '${ARCHIVAL_CUTOFF_YEAR}-01-01'
GROUP BY ci_id;
EOF

# Step 4: Remove archived data from main table (only after verification)
echo "WARNING: About to delete archived data. Verify backups before proceeding."
read -p "Continue with deletion? (y/N): " -n 1 -r
echo

if [[ $REPLY =~ ^[Yy]$ ]]; then
    psql -h localhost -U pustaka -d pustaka << EOF
DELETE FROM amortization_ledger
WHERE entry_date < '${ARCHIVAL_CUTOFF_YEAR}-01-01';

-- Verify deletion
SELECT COUNT(*) as remaining_entries FROM amortization_ledger;
EOF
    echo "Archival deletion completed"
else
    echo "Archival deletion cancelled"
fi
```

### 2. Long-term Storage Strategy

#### Cold Storage Retention
- **Format**: Compressed PostgreSQL dumps
- **Storage**: Amazon S3 Glacier Deep Archive
- **Retention**: 10 years minimum
- **Accessibility**: 12-hour retrieval time
- **Cost**: $0.00099 per GB/month

#### Metadata Retention
```sql
-- Create archival metadata table for tracking
CREATE TABLE amortization_archival_metadata (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    archival_date DATE NOT NULL,
    data_type VARCHAR(50) NOT NULL,
    year_affected INTEGER NOT NULL,
    record_count BIGINT NOT NULL,
    archive_location TEXT NOT NULL,
    checksum_md5 TEXT NOT NULL,
    retention_until DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id)
);

-- Index for efficient archival tracking
CREATE INDEX idx_archival_metadata_date ON amortization_archival_metadata(archival_date);
CREATE INDEX idx_archival_metadata_year ON amortization_archival_metadata(year_affected);
```

## Monitoring and Alerting

### 1. Backup Monitoring

#### Backup Success Monitoring
```bash
#!/bin/bash
# Monitor backup success and alert on failures
BACKUP_SUCCESS_LOG="/var/log/amortization/backup_status.log"
ALERT_THRESHOLD_HOURS=24

# Check last successful backup
LAST_BACKUP=$(aws s3 ls s3://pustaka-backups/amortization/daily/ --recursive | sort | tail -n 1)
BACKUP_TIMESTAMP=$(echo ${LAST_BACKUP} | awk '{print $1, $2}')

# Convert to epoch time for comparison
BACKUP_EPOCH=$(date -d "${BACKUP_TIMESTAMP}" +%s)
CURRENT_EPOCH=$(date +%s)
HOURS_SINCE_BACKUP=$(( (CURRENT_EPOCH - BACKUP_EPOCH) / 3600 ))

if [ ${HOURS_SINCE_BACKUP} -gt ${ALERT_THRESHOLD_HOURS} ]; then
    echo "ALERT: Backup is ${HOURS_SINCE_BACKUP} hours old" >> ${BACKUP_SUCCESS_LOG}

    # Send alert
    curl -X POST "https://hooks.slack.com/your-webhook" \
      -H 'Content-type: application/json' \
      --data "{\"text\":\"🚨 Amortization backup is ${HOURS_SINCE_BACKUP} hours old! Last backup: ${BACKUP_TIMESTAMP}\"}"
fi
```

### 2. Recovery Time Testing

#### Monthly Recovery Testing
```bash
#!/bin/bash
# Monthly disaster recovery test
TEST_DATE=$(date +%Y%m%d)
TEST_DB="pustaka_amortization_test_${TEST_DATE}"
TEST_CONTAINER="pustaka-test-${TEST_DATE}"

echo "Starting monthly recovery test at $(date)"

# Step 1: Create test database
docker run -d --name ${TEST_CONTAINER} \
  -e POSTGRES_DB=${TEST_DB} \
  -e POSTGRES_USER=pustaka \
  -e POSTGRES_PASSWORD=password \
  postgres:15

# Wait for test database to be ready
sleep 30

# Step 2: Test restoration from latest backup
docker exec ${TEST_CONTAINER} createdb -U pustaka ${TEST_DB}

LATEST_BACKUP=$(aws s3 ls s3://pustaka-backups/amortization/daily/ --recursive | grep "dump" | sort | tail -n 1 | awk '{print $4}')
aws s3 cp "s3://pustaka-backups/amortization/daily/${LATEST_BACKUP}" /tmp/test_restore.dump

docker cp /tmp/test_restore.dump ${TEST_CONTAINER}:/tmp/test_restore.dump

# Step 3: Restore and verify
docker exec ${TEST_CONTAINER} pg_restore -U pustaka -d ${TEST_DB} /tmp/test_restore.dump

# Step 4: Verify data integrity
RESTORE_RESULT=$(docker exec ${TEST_CONTAINER} psql -U pustaka -d ${TEST_DB} -t -c "
SELECT COUNT(*) FROM amortization_ledger;
SELECT COUNT(*) FROM configuration_items WHERE purchase_cost IS NOT NULL;
")

echo "Restore verification result: ${RESTORE_RESULT}"

# Step 5: Cleanup
docker stop ${TEST_CONTAINER}
docker rm ${TEST_CONTAINER}
rm /tmp/test_restore.dump

echo "Monthly recovery test completed at $(date)"
```

This comprehensive backup and recovery strategy ensures the amortization module's financial data is properly protected, quickly recoverable, and compliant with regulatory requirements while providing robust monitoring and testing procedures.