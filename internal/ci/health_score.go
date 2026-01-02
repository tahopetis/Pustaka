package ci

import (
	"time"

	"github.com/google/uuid"
)

// HealthScore represents the overall CMDB health score with sub-scores
type HealthScore struct {
	Overall      float64    `json:"overall"`       // 0-100
	Completeness float64    `json:"completeness"`  // 0-100
	Correctness  float64    `json:"correctness"`   // 0-100
	Compliance   float64    `json:"compliance"`    // 0-100
	Trend        string     `json:"trend"`         // 'improving' | 'declining' | 'stable'
	CalculatedAt time.Time  `json:"calculated_at"`
}

// HealthScoreMetrics represents the raw metrics used to calculate health scores
type HealthScoreMetrics struct {
	TotalCIs      int `json:"total_cis"`
	CompleteCIs   int `json:"complete_cis"`   // CIs with all required attributes
	CurrentCIs    int `json:"current_cis"`    // CIs updated in last 90 days
	CompliantCIs  int `json:"compliant_cis"`  // CIs following naming standards
}

// HealthScoreHistory represents a historical health score record
type HealthScoreHistory struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	CalculatedAt      time.Time  `json:"calculated_at" db:"calculated_at"`
	OverallScore      float64    `json:"overall_score" db:"overall_score"`
	CompletenessScore float64    `json:"completeness_score" db:"completeness_score"`
	CorrectnessScore  float64    `json:"correctness_score" db:"correctness_score"`
	ComplianceScore   float64    `json:"compliance_score" db:"compliance_score"`
	TotalCIs          int        `json:"total_cis" db:"total_cis"`
	CompleteCIs       int        `json:"complete_cis" db:"complete_cis"`
	CurrentCIs        int        `json:"current_cis" db:"current_cis"`
	CompliantCIs      int        `json:"compliant_cis" db:"compliant_cis"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

// calculateTrend determines the trend by comparing current score with historical data
func calculateTrend(currentScore float64, historicalScores []HealthScoreHistory) string {
	if len(historicalScores) == 0 {
		return "stable"
	}

	// Get the score from 30 days ago (or closest available)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	var previousScore float64

	for _, record := range historicalScores {
		if record.CalculatedAt.Before(thirtyDaysAgo) || record.CalculatedAt.Equal(thirtyDaysAgo) {
			previousScore = record.OverallScore
			break
		}
	}

	// If no 30-day data, use the oldest available
	if previousScore == 0 && len(historicalScores) > 0 {
		previousScore = historicalScores[len(historicalScores)-1].OverallScore
	}

	// Calculate trend
	diff := currentScore - previousScore
	if diff > 2 { // More than 2% improvement
		return "improving"
	} else if diff < -2 { // More than 2% decline
		return "declining"
	}
	return "stable"
}

// CalculateHealthScore computes the health score from metrics
func CalculateHealthScore(metrics HealthScoreMetrics) HealthScore {
	var completeness, correctness, compliance float64

	// Calculate completeness
	if metrics.TotalCIs > 0 {
		completeness = float64(metrics.CompleteCIs) / float64(metrics.TotalCIs) * 100
	}

	// Calculate correctness
	if metrics.TotalCIs > 0 {
		correctness = float64(metrics.CurrentCIs) / float64(metrics.TotalCIs) * 100
	}

	// Calculate compliance
	if metrics.TotalCIs > 0 {
		compliance = float64(metrics.CompliantCIs) / float64(metrics.TotalCIs) * 100
	}

	// Overall score is the average of the three sub-scores
	overall := (completeness + correctness + compliance) / 3

	return HealthScore{
		Overall:      roundToTwoDecimals(overall),
		Completeness: roundToTwoDecimals(completeness),
		Correctness:  roundToTwoDecimals(correctness),
		Compliance:   roundToTwoDecimals(compliance),
		CalculatedAt: time.Now(),
		Trend:        "stable", // Will be updated after comparing with historical data
	}
}

// roundToTwoDecimals rounds a float64 to two decimal places
func roundToTwoDecimals(num float64) float64 {
	return float64(int(num*100+0.5)) / 100
}
