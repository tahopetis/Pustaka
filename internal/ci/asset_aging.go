package ci

import "time"

// AssetAgingMetrics represents asset aging analysis for the dashboard
type AssetAgingMetrics struct {
	Distribution      AgeDistribution      `json:"distribution"`
	ApproachingEOL    []ApproachingEOLAsset `json:"approaching_eol"`
	AverageAgeMonths  float64              `json:"average_age_months"`
	OldestAsset       *OldestAsset         `json:"oldest_asset,omitempty"`
}

// AgeDistribution represents the distribution of assets by age
type AgeDistribution struct {
	LessThan1Year  int `json:"less_than_1_year"`
	OneTo3Years    int `json:"one_to_3_years"`
	ThreeTo5Years  int `json:"three_to_5_years"`
	MoreThan5Years int `json:"more_than_5_years"`
}

// ApproachingEOLAsset represents an asset approaching end-of-life
type ApproachingEOLAsset struct {
	CI         CIReference `json:"ci"`
	EOLDate    time.Time   `json:"eol_date"`
	DaysUntilEOL int       `json:"days_until_eol"`
	Type       string      `json:"type"`
}

// CIReference is a minimal CI reference for linked assets
type CIReference struct {
	ID   string    `json:"id"`
	Name string    `json:"name"`
}

// OldestAsset represents the oldest asset in the CMDB
type OldestAsset struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	AgeMonths   int       `json:"age_months"`
	CreatedAt   time.Time `json:"created_at"`
}

// AssetAgingFilterOptions represents filtering options for asset aging queries
type AssetAgingFilterOptions struct {
	EOLThresholdMonths int    `json:"eol_threshold_months"` // Default: 6 months
	Limit              int    `json:"limit"`                 // Max EOL assets to return
	SortBy             string `json:"sort_by"`               // days_until_eol, eol_date, name
	SortOrder          string `json:"sort_order"`            // asc, desc
}
