package amortization

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tahopetis/porta/pkg/logger"
)

func testAmortizableCI(start time.Time) *AmortizableCI {
	return &AmortizableCI{
		ID:                      uuid.New(),
		Name:                    "laptop-01",
		PurchaseCost:            1200,
		SalvageValue:            0,
		AmortStartDate:          &start,
		UsefulLifeMonths:        24,
		CurrentBookValue:        1050,
		AccumulatedDepreciation: 150,
		DepreciationMethod:      "straight_line",
	}
}

func TestCalculateMonthlyDepreciationStraightLine(t *testing.T) {
	start := time.Now().AddDate(0, -3, 0)
	ci := testAmortizableCI(start)

	result, err := NewCalculator(logger.Default()).CalculateMonthlyDepreciation(context.Background(), ci, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Independent expectation: (1200 - 0) / 24 = 50 per month.
	if result.Amount != 50 {
		t.Fatalf("amount = %v, want 50", result.Amount)
	}
	if result.BookValueBefore != 1050 {
		t.Fatalf("book value before = %v, want 1050", result.BookValueBefore)
	}
	if result.BookValueAfter != 1000 {
		t.Fatalf("book value after = %v, want 1000", result.BookValueAfter)
	}
	if result.AccumulatedDepreciationAfter != 200 {
		t.Fatalf("accumulated after = %v, want 200", result.AccumulatedDepreciationAfter)
	}
	if result.Method != "straight_line" {
		t.Fatalf("method = %q, want straight_line", result.Method)
	}
}

func TestCalculateMonthlyDepreciationClampsToSalvageValue(t *testing.T) {
	start := time.Now().AddDate(0, -23, 0)
	ci := testAmortizableCI(start)
	ci.CurrentBookValue = 30
	ci.AccumulatedDepreciation = 1170

	result, err := NewCalculator(logger.Default()).CalculateMonthlyDepreciation(context.Background(), ci, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Monthly amount would be 50, but only 30 remains above salvage 0.
	if result.Amount != 30 {
		t.Fatalf("amount = %v, want 30", result.Amount)
	}
	if result.BookValueAfter != 0 {
		t.Fatalf("book value after = %v, want 0", result.BookValueAfter)
	}
}

func TestCalculateMonthlyDepreciationRejectsInvalidConfig(t *testing.T) {
	start := time.Now().AddDate(0, -1, 0)

	tests := []struct {
		name   string
		mutate func(*AmortizableCI)
	}{
		{name: "no purchase cost", mutate: func(ci *AmortizableCI) { ci.PurchaseCost = 0 }},
		{name: "no useful life", mutate: func(ci *AmortizableCI) { ci.UsefulLifeMonths = 0 }},
		{name: "no start date", mutate: func(ci *AmortizableCI) { ci.AmortStartDate = nil }},
		{name: "book value at salvage", mutate: func(ci *AmortizableCI) {
			ci.CurrentBookValue = ci.SalvageValue
		}},
	}

	calculator := NewCalculator(logger.Default())
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ci := testAmortizableCI(start)
			tc.mutate(ci)

			if _, err := calculator.CalculateMonthlyDepreciation(context.Background(), ci, time.Now()); err == nil {
				t.Fatal("expected an error, got nil")
			}
		})
	}
}
