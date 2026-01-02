package ci

import (
	"time"
)

// DataQualityMetrics represents all data quality issues in the CMDB
type DataQualityMetrics struct {
	MissingAttributes  QualityIssue `json:"missing_attributes"`
	OrphanedCIs        QualityIssue `json:"orphaned_cis"`
	NoLifecycleStatus QualityIssue `json:"no_lifecycle_status"`
	NoTags             QualityIssue `json:"no_tags"`
	Stale30Days        int         `json:"stale_30_days"`
	Stale60Days        int         `json:"stale_60_days"`
	Stale90Days        int         `json:"stale_90_days"`
	Duplicates         int         `json:"duplicates"`
}

// QualityIssue represents a specific type of data quality issue
type QualityIssue struct {
	Count      int              `json:"count"`
	Percentage float64          `json:"percentage"`
	CIs        []QualityIssueCI `json:"cis,omitempty"`
}

// QualityIssueCI represents a CI with a data quality issue
type QualityIssueCI struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	CIType       string      `json:"ci_type"`
	MissingFields []string   `json:"missing_fields,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

// DataQualityMetricsSummary represents summary stats without detailed CI lists
// Used for performance when detailed lists aren't needed
type DataQualityMetricsSummary struct {
	MissingAttributes  QualityIssueSummary `json:"missing_attributes"`
	OrphanedCIs        QualityIssueSummary `json:"orphaned_cis"`
	NoLifecycleStatus QualityIssueSummary `json:"no_lifecycle_status"`
	NoTags             QualityIssueSummary `json:"no_tags"`
	Stale30Days        int                  `json:"stale_30_days"`
	Stale60Days        int                  `json:"stale_60_days"`
	Stale90Days        int                  `json:"stale_90_days"`
	Duplicates         int                  `json:"duplicates"`
}

// QualityIssueSummary represents summary stats without CI details
type QualityIssueSummary struct {
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// DataQualityFilterOptions represents filtering options for data quality queries
type DataQualityFilterOptions struct {
	IncludeDetails bool   `json:"include_details"` // If true, include list of affected CIs
	Limit           int    `json:"limit"`           // Max number of CIs to return per category
	Offset          int    `json:"offset"`
	SortBy          string `json:"sort_by"`         // count, percentage, name
	SortOrder       string `json:"sort_order"`      // asc, desc
}
