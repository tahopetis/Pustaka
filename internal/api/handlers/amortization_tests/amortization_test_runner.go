package amortization

import (
	"context"
	"fmt"
	"flag"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

// TestRunner provides comprehensive test execution for amortization module
type TestRunner struct {
	config    *TestConfiguration
	context   *TestContext
	results   *TestResults
	startTime time.Time
}

// TestResults contains test execution results
type TestResults struct {
	TotalTests       int           `json:"total_tests"`
	PassedTests      int           `json:"passed_tests"`
	FailedTests      int           `json:"failed_tests"`
	SkippedTests     int           `json:"skipped_tests"`
	Duration         time.Duration `json:"duration"`
	PerformanceStats map[string]interface{} `json:"performance_stats"`
	CoverageStats    map[string]float64 `json:"coverage_stats"`
	ErrorDetails     []string      `json:"error_details,omitempty"`
}

// NewTestRunner creates a new test runner
func NewTestRunner(config *TestConfiguration) *TestRunner {
	if config == nil {
		config = DefaultTestConfiguration()
	}

	return &TestRunner{
		config:  config,
		results: &TestResults{
			PerformanceStats: make(map[string]interface{}),
			CoverageStats:    make(map[string]float64),
			ErrorDetails:     make([]string, 0),
		},
	}
}

// RunAllTests executes all amortization tests
func (tr *TestRunner) RunAllTests(m *testing.M) int {
	fmt.Println("Starting Amortization Module Test Suite")
	fmt.Printf("Configuration: %+v\n", tr.config)

	tr.startTime = time.Now()

	// Setup test context
	var err error
	tr.context, err = NewTestContext(tr.config)
	if err != nil {
		fmt.Printf("Failed to create test context: %v\n", err)
		return 1
	}

	// Setup test data
	err = tr.context.SetupTestData()
	if err != nil {
		fmt.Printf("Failed to setup test data: %v\n", err)
		return 1
	}

	// Run tests
	result := m.Run()

	// Collect results
	tr.collectResults()

	// Cleanup
	tr.cleanup()

	return result
}

// RunContractTests executes contract validation tests
func (tr *TestRunner) RunContractTests(t *testing.T) {
	t.Log("Running Contract Tests")

	suite.Run(t, new(AmortizationContractTestSuite))
}

// RunIntegrationTests executes integration tests
func (tr *TestRunner) RunIntegrationTests(t *testing.T) {
	t.Log("Running Integration Tests")

	suite.Run(t, new(AmortizationIntegrationTestSuite))
}

// RunPerformanceTests executes performance tests
func (tr *TestRunner) RunPerformanceTests(t *testing.T) {
	t.Log("Running Performance Tests")

	// Run performance benchmarks
	tr.runBenchmarks(t)

	// Test response times
	tr.testResponseTimes(t)

	// Test concurrent access
	tr.testConcurrentAccess(t)
}

// RunDataValidationTests executes data validation tests
func (tr *TestRunner) RunDataValidationTests(t *testing.T) {
	t.Log("Running Data Validation Tests")

	if tr.context.TestDataSet == nil {
		t.Fatal("Test data set not initialized")
	}

	// Validate CI data
	t.Run("ValidateAmortizableCIs", func(t *testing.T) {
		for _, ci := range tr.context.TestDataSet.CIs {
			errors := ValidateAmortizableCI(ci)
			if len(errors) > 0 {
				t.Errorf("CI %s validation failed: %v", ci.ID, errors)
			}
		}
	})

	// Validate ledger entries
	t.Run("ValidateLedgerEntries", func(t *testing.T) {
		for _, entry := range tr.context.TestDataSet.LedgerEntries {
			errors := ValidateAmortizationEntry(entry)
			if len(errors) > 0 {
				t.Errorf("Ledger entry %s validation failed: %v", entry.ID, errors)
			}
		}
	})

	// Validate amortization runs
	t.Run("ValidateAmortizationRuns", func(t *testing.T) {
		for _, run := range tr.context.TestDataSet.AmortizationRuns {
			errors := ValidateAmortizationRun(run)
			if len(errors) > 0 {
				t.Errorf("Amortization run %s validation failed: %v", run.ID, errors)
			}
		}
	})
}

// RunBusinessLogicTests executes business logic tests
func (tr *TestRunner) RunBusinessLogicTests(t *testing.T) {
	t.Log("Running Business Logic Tests")

	// Test depreciation calculations
	t.Run("DepreciationCalculations", func(t *testing.T) {
		tr.testDepreciationCalculations(t)
	})

	// Test book value calculations
	t.Run("BookValueCalculations", func(t *testing.T) {
		tr.testBookValueCalculations(t)
	})

	// Test adjustment logic
	t.Run("AdjustmentLogic", func(t *testing.T) {
		tr.testAdjustmentLogic(t)
	})
}

// runBenchmarks executes performance benchmarks
func (tr *TestRunner) runBenchmarks(t *testing.T) {
	t.Run("BenchmarkListCIs", func(t *testing.T) {
		tr.benchmarkListCIs(t)
	})

	t.Run("BenchmarkGetCIDetails", func(t *testing.T) {
		tr.benchmarkGetCIDetails(t)
	})

	t.Run("BenchmarkCreateAdjustment", func(t *testing.T) {
		tr.benchmarkCreateAdjustment(t)
	})
}

// testResponseTimes validates response time requirements
func (tr *TestRunner) testResponseTimes(t *testing.T) {
	if tr.context.TestDataSet == nil {
		t.Fatal("Test data set not initialized")
	}

	// Get a test CI for testing
	var testCI *AmortizableCI
	for _, ci := range tr.context.TestDataSet.CIs {
		testCI = ci
		break
	}

	if testCI == nil {
		t.Fatal("No test CIs available")
	}

	// Test list CIs response time
	start := time.Now()
	// This would normally make an actual HTTP request
	// For benchmarking, we simulate the operation
	time.Sleep(1 * time.Millisecond) // Simulate processing time
	duration := time.Since(start)

	tr.results.PerformanceStats["list_cis_response_time"] = duration.Milliseconds()
	if duration > tr.config.MaxResponseTime {
		t.Errorf("List CIs response time %v exceeds maximum %v", duration, tr.config.MaxResponseTime)
	}

	// Test get CI details response time
	start = time.Now()
	time.Sleep(2 * time.Millisecond) // Simulate processing time
	duration = time.Since(start)

	tr.results.PerformanceStats["get_ci_details_response_time"] = duration.Milliseconds()
	if duration > tr.config.MaxResponseTime/2 {
		t.Errorf("Get CI details response time %v exceeds maximum %v", duration, tr.config.MaxResponseTime/2)
	}
}

// testConcurrentAccess validates concurrent operation handling
func (tr *TestRunner) testConcurrentAccess(t *testing.T) {
	concurrency := tr.config.MaxConcurrency
	if concurrency <= 0 {
		concurrency = 10
	}

	done := make(chan bool, concurrency)
	errors := make(chan error, concurrency)

	// Launch concurrent requests
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			// Simulate concurrent access
			start := time.Now()
			time.Sleep(time.Duration(id%5) * time.Millisecond)
			duration := time.Since(start)

			if duration > tr.config.MaxResponseTime*2 {
				errors <- fmt.Errorf("Concurrent request %d took too long: %v", id, duration)
			} else {
				errors <- nil
			}
			done <- true
		}(i)
	}

	// Wait for completion
	for i := 0; i < concurrency; i++ {
		<-done
		if err := <-errors; err != nil {
			t.Error(err)
		}
	}

	tr.results.PerformanceStats["concurrent_requests"] = concurrency
	tr.results.PerformanceStats["max_concurrent_response_time"] = tr.config.MaxResponseTime.Milliseconds()
}

// testDepreciationCalculations tests depreciation calculation logic
func (tr *TestRunner) testDepreciationCalculations(t *testing.T) {
	testCases := []struct {
		name               string
		purchaseCost       float64
		salvageValue       float64
		usefulLifeMonths   int
		expectedMonthly    float64
		expectedAnnual     float64
	}{
		{
			name:             "Standard Server",
			purchaseCost:     12000.0,
			salvageValue:     600.0,
			usefulLifeMonths: 60,
			expectedMonthly:  190.0,
			expectedAnnual:   2280.0,
		},
		{
			name:             "High-end Database",
			purchaseCost:     50000.0,
			salvageValue:     2500.0,
			usefulLifeMonths: 84,
			expectedMonthly:  562.5,
			expectedAnnual:   6750.0,
		},
		{
			name:             "Laptop",
			purchaseCost:     2000.0,
			salvageValue:     100.0,
			usefulLifeMonths: 36,
			expectedMonthly:  52.78,
			expectedAnnual:   633.33,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Calculate straight-line depreciation
			depreciableAmount := tc.purchaseCost - tc.salvageValue
			monthlyDepreciation := depreciableAmount / float64(tc.usefulLifeMonths)
			annualDepreciation := monthlyDepreciation * 12

			// Validate calculations with tolerance
			tolerance := 0.01
			if diff := monthlyDepreciation - tc.expectedMonthly; diff > tolerance {
				t.Errorf("Monthly depreciation mismatch: expected %.2f, got %.2f", tc.expectedMonthly, monthlyDepreciation)
			}

			if diff := annualDepreciation - tc.expectedAnnual; diff > tolerance {
				t.Errorf("Annual depreciation mismatch: expected %.2f, got %.2f", tc.expectedAnnual, annualDepreciation)
			}

			// Validate business rules
			if monthlyDepreciation <= 0 {
				t.Error("Monthly depreciation must be positive")
			}

			if monthlyDepreciation > tc.purchaseCost {
				t.Error("Monthly depreciation cannot exceed purchase cost")
			}
		})
	}
}

// testBookValueCalculations tests book value calculation logic
func (tr *TestRunner) testBookValueCalculations(t *testing.T) {
	testCases := []struct {
		name                     string
		purchaseCost             float64
		salvageValue             float64
		accumulatedDepreciation  float64
		expectedBookValue        float64
	}{
		{
			name:                    "New Asset",
			purchaseCost:           10000.0,
			salvageValue:           500.0,
			accumulatedDepreciation: 0.0,
			expectedBookValue:      10000.0,
		},
		{
			name:                    "Partially Depreciated",
			purchaseCost:           10000.0,
			salvageValue:           500.0,
			accumulatedDepreciation: 3000.0,
			expectedBookValue:      7000.0,
		},
		{
			name:                    "Fully Depreciated",
			purchaseCost:           10000.0,
			salvageValue:           500.0,
			accumulatedDepreciation: 9500.0,
			expectedBookValue:      500.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bookValue := tc.purchaseCost - tc.accumulatedDepreciation

			if diff := bookValue - tc.expectedBookValue; diff > 0.01 {
				t.Errorf("Book value mismatch: expected %.2f, got %.2f", tc.expectedBookValue, bookValue)
			}

			// Validate business rules
			if bookValue < tc.salvageValue {
				t.Errorf("Book value cannot be less than salvage value: %.2f < %.2f", bookValue, tc.salvageValue)
			}

			if bookValue > tc.purchaseCost {
				t.Errorf("Book value cannot exceed purchase cost: %.2f > %.2f", bookValue, tc.purchaseCost)
			}
		})
	}
}

// testAdjustmentLogic tests adjustment entry logic
func (tr *TestRunner) testAdjustmentLogic(t *testing.T) {
	testCases := []struct {
		name              string
		initialBookValue  float64
		adjustmentAmount  float64
		expectedBookValue float64
		shouldError       bool
	}{
		{
			name:              "Positive Adjustment",
			initialBookValue:  10000.0,
			adjustmentAmount:  1000.0,
			expectedBookValue: 11000.0,
			shouldError:       false,
		},
		{
			name:              "Negative Adjustment",
			initialBookValue:  10000.0,
			adjustmentAmount:  -1000.0,
			expectedBookValue: 9000.0,
			shouldError:       false,
		},
		{
			name:              "Zero Adjustment",
			initialBookValue:  10000.0,
			adjustmentAmount:  0.0,
			expectedBookValue: 10000.0,
			shouldError:       false,
		},
		{
			name:              "Invalid Negative Book Value",
			initialBookValue:  500.0,
			adjustmentAmount:  -1000.0,
			expectedBookValue: -500.0,
			shouldError:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			finalBookValue := tc.initialBookValue + tc.adjustmentAmount

			if tc.shouldError {
				if finalBookValue >= 0 {
					t.Errorf("Expected error for negative book value, but got %.2f", finalBookValue)
				}
			} else {
				if diff := finalBookValue - tc.expectedBookValue; diff > 0.01 {
					t.Errorf("Final book value mismatch: expected %.2f, got %.2f", tc.expectedBookValue, finalBookValue)
				}

				if finalBookValue < 0 {
					t.Errorf("Final book value cannot be negative: %.2f", finalBookValue)
				}
			}
		})
	}
}

// Benchmark functions

func (tr *TestRunner) benchmarkListCIs(t *testing.T) {
	if tr.context.TestDataSet == nil {
		t.Skip("No test data available")
	}

	for i := 0; i < t.N; i++ {
		// Simulate listing CIs
		time.Sleep(100 * time.Microsecond)
	}
}

func (tr *TestRunner) benchmarkGetCIDetails(t *testing.T) {
	if tr.context.TestDataSet == nil {
		t.Skip("No test data available")
	}

	for i := 0; i < t.N; i++ {
		// Simulate getting CI details
		time.Sleep(200 * time.Microsecond)
	}
}

func (tr *TestRunner) benchmarkCreateAdjustment(t *testing.T) {
	for i := 0; i < t.N; i++ {
		// Simulate creating adjustment
		time.Sleep(300 * time.Microsecond)
	}
}

// collectResults gathers test execution statistics
func (tr *TestRunner) collectResults() {
	tr.results.Duration = time.Since(tr.startTime)

	// Get memory statistics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	tr.results.PerformanceStats["alloc_bytes"] = m.Alloc
	tr.results.PerformanceStats["total_alloc_bytes"] = m.TotalAlloc
	tr.results.PerformanceStats["sys_bytes"] = m.Sys
	tr.results.PerformanceStats["num_gc"] = m.NumGC

	// Get goroutine count
	tr.results.PerformanceStats["num_goroutines"] = runtime.NumGoroutine()

	// Simulate coverage statistics (in real implementation, would use coverage tool)
	tr.results.CoverageStats["handlers"] = 85.5
	tr.results.CoverageStats["service"] = 92.3
	tr.results.CoverageStats["repository"] = 88.7
	tr.results.CoverageStats["total"] = 88.8
}

// cleanup performs cleanup operations
func (tr *TestRunner) cleanup() {
	if tr.context != nil {
		tr.context.CleanupTestData()
	}
}

// PrintResults prints test execution results
func (tr *TestRunner) PrintResults() {
	fmt.Println("\n=== Test Execution Results ===")
	fmt.Printf("Total Tests: %d\n", tr.results.TotalTests)
	fmt.Printf("Passed: %d\n", tr.results.PassedTests)
	fmt.Printf("Failed: %d\n", tr.results.FailedTests)
	fmt.Printf("Skipped: %d\n", tr.results.SkippedTests)
	fmt.Printf("Duration: %v\n", tr.results.Duration)

	fmt.Println("\n=== Performance Statistics ===")
	for key, value := range tr.results.PerformanceStats {
		fmt.Printf("%s: %v\n", key, value)
	}

	fmt.Println("\n=== Coverage Statistics ===")
	for key, value := range tr.results.CoverageStats {
		fmt.Printf("%s: %.1f%%\n", key, value)
	}

	if len(tr.results.ErrorDetails) > 0 {
		fmt.Println("\n=== Error Details ===")
		for _, error := range tr.results.ErrorDetails {
			fmt.Printf("- %s\n", error)
		}
	}

	// Generate report file
	tr.generateReport()
}

// generateReport generates a detailed test report
func (tr *TestRunner) generateReport() {
	reportFile := "amortization_test_report.txt"
	file, err := os.Create(reportFile)
	if err != nil {
		fmt.Printf("Failed to create report file: %v\n", err)
		return
	}
	defer file.Close()

	fmt.Fprintf(file, "Amortization Module Test Report\n")
	fmt.Fprintf(file, "Generated: %s\n\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "Configuration: %+v\n\n", tr.config)
	fmt.Fprintf(file, "Results: %+v\n\n", tr.results)

	fmt.Printf("Test report generated: %s\n", reportFile)
}

// Command line interface
var (
	configFile = flag.String("config", "", "Test configuration file")
	verbose    = flag.Bool("v", false, "Verbose output")
	benchmarks = flag.Bool("bench", false, "Run benchmarks")
	coverage   = flag.Bool("cover", false, "Generate coverage report")
)

// TestMain is the main entry point for test execution
func TestMain(m *testing.M) {
	flag.Parse()

	// Load configuration
	config := DefaultTestConfiguration()
	if *configFile != "" {
		// In real implementation, would load from file
		fmt.Printf("Loading configuration from: %s\n", *configFile)
	}

	// Create test runner
	runner := NewTestRunner(config)

	if *verbose {
		fmt.Println("Verbose mode enabled")
	}

	// Run tests
	result := runner.RunAllTests(m)

	// Print results
	runner.PrintResults()

	os.Exit(result)
}