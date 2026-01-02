package ci

// RiskMetrics contains risk assessment metrics for the dashboard
type RiskMetrics struct {
	RiskScore            int              `json:"risk_score"`            // 0-100
	SPOFCount            int              `json:"spof_count"`            // Single points of failure
	CriticalAssetsCount  int              `json:"critical_assets_count"` // Assets without redundancy
	NoRedundancyCount    int              `json:"no_redundancy_count"`   // Assets without backup
	ComplianceViolations int              `json:"compliance_violations"` // Compliance issues
	HighRiskCIs          []HighRiskCI     `json:"high_risk_cis"`        // Top high-risk assets
	LastUpdated          string           `json:"last_updated"`
}

// HighRiskCI represents a CI with elevated risk
type HighRiskCI struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	CIType          string  `json:"ci_type"`
	RiskScore       int     `json:"risk_score"`       // 0-100
	IsAmortizable   bool    `json:"is_amortizable"`   // Whether it's an asset
	HasRedundancy   bool    `json:"has_redundancy"`   // Has backup/redundancy
	IsCritical      bool    `json:"is_critical"`      // No redundancy + amortizable
	AgeMonths       int     `json:"age_months"`      // Age in months
	HasTags         bool    `json:"has_tags"`        // Whether it has tags
}
