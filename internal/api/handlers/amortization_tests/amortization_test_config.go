package amortization

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TestConfiguration provides configuration for amortization tests
type TestConfiguration struct {
	// Database settings
	DatabaseURL string `json:"database_url"`
	UseMemoryDB bool   `json:"use_memory_db"`

	// Authentication settings
	JWTSecret     string `json:"jwt_secret"`
	TestUserToken string `json:"test_user_token"`
	TestAdminToken string `json:"test_admin_token"`

	// Test data settings
	NumTestCIs           int `json:"num_test_cis"`
	NumTestLedgerEntries int `json:"num_test_ledger_entries"`
	NumTestRuns          int `json:"num_test_runs"`

	// Performance settings
	MaxResponseTime       time.Duration `json:"max_response_time"`
	MaxConcurrency        int           `json:"max_concurrency"`
	EnableProfiling       bool          `json:"enable_profiling"`

	// Validation settings
	StrictValidation      bool `json:"strict_validation"`
	ValidateCalculations  bool `json:"validate_calculations"`
	ValidateConsistency   bool `json:"validate_consistency"`
}

// DefaultTestConfiguration returns default test configuration
func DefaultTestConfiguration() *TestConfiguration {
	return &TestConfiguration{
		DatabaseURL:    "memory://test",
		UseMemoryDB:    true,
		JWTSecret:      "test-secret-key-for-testing-only",
		TestUserToken:  "test-user-token",
		TestAdminToken: "test-admin-token",

		NumTestCIs:           50,
		NumTestLedgerEntries: 200,
		NumTestRuns:          10,

		MaxResponseTime: 100 * time.Millisecond,
		MaxConcurrency:  10,
		EnableProfiling: false,

		StrictValidation:     true,
		ValidateCalculations: true,
		ValidateConsistency:  true,
	}
}

// TestContext provides context for test execution
type TestContext struct {
	Config          *TestConfiguration
	DataFactory     *TestDataFactory
	TestDataSet     *TestDataSet
	Authentication  *TestAuthentication
	Database        TestDatabase
	Logger          TestLogger
}

// TestAuthentication provides test authentication utilities
type TestAuthentication struct {
	AdminUser    *TestUser
	RegularUser  *TestUser
	ReadOnlyUser *TestUser
	Tokens       map[string]string
}

// TestUser represents a test user
type TestUser struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	Permissions []string  `json:"permissions"`
	Token       string    `json:"token"`
}

// TestDatabase provides test database interface
type TestDatabase interface {
	Setup() error
	Cleanup() error
	BeginTx(ctx context.Context) (TestTransaction, error)
	SeedData(ctx context.Context, dataSet *TestDataSet) error
	ValidateData(ctx context.Context) error
}

// TestTransaction provides test transaction interface
type TestTransaction interface {
	Commit() error
	Rollback() error
	Exec(query string, args ...interface{}) error
	Query(query string, args ...interface{}) (TestRows, error)
}

// TestRows provides test rows interface
type TestRows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close() error
}

// TestLogger provides test logging interface
type TestLogger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	GetLogs() []TestLogEntry
	ClearLogs()
}

// TestLogEntry represents a test log entry
type TestLogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// NewTestContext creates a new test context
func NewTestContext(config *TestConfiguration) (*TestContext, error) {
	if config == nil {
		config = DefaultTestConfiguration()
	}

	ctx := &TestContext{
		Config:      config,
		DataFactory: NewTestDataFactory(),
		TestDataSet: nil,
	}

	// Setup authentication
	ctx.Authentication = &TestAuthentication{
		AdminUser: &TestUser{
			ID:       uuid.New(),
			Username: "test-admin",
			Email:    "admin@test.com",
			Role:     "admin",
			Permissions: []string{
				"amortization:read", "amortization:write",
				"amortization:adjust", "amortization:admin",
				"system:admin",
			},
			Token: config.TestAdminToken,
		},
		RegularUser: &TestUser{
			ID:       uuid.New(),
			Username: "test-user",
			Email:    "user@test.com",
			Role:     "user",
			Permissions: []string{
				"amortization:read", "amortization:write",
				"amortization:adjust",
			},
			Token: config.TestUserToken,
		},
		ReadOnlyUser: &TestUser{
			ID:       uuid.New(),
			Username: "readonly-user",
			Email:    "readonly@test.com",
			Role:     "readonly",
			Permissions: []string{"amortization:read"},
			Token: "readonly-token",
		},
		Tokens: map[string]string{
			"admin":    config.TestAdminToken,
			"user":     config.TestUserToken,
			"readonly": "readonly-token",
		},
	}

	// Initialize database (memory implementation for testing)
	ctx.Database = NewMemoryTestDatabase()

	// Initialize logger
	ctx.Logger = NewMemoryTestLogger()

	return ctx, nil
}

// SetupTestData creates and sets up test data
func (ctx *TestContext) SetupTestData() error {
	ctx.TestDataSet = ctx.DataFactory.CreateTestDataSet(
		ctx.Config.NumTestCIs,
		ctx.Config.NumTestLedgerEntries,
		ctx.Config.NumTestRuns,
	)

	return nil
}

// CleanupTestData cleans up test data
func (ctx *TestContext) CleanupTestData() error {
	ctx.TestDataSet = nil
	if ctx.Logger != nil {
		ctx.Logger.ClearLogs()
	}
	return nil
}

// ValidateTestResults validates test results against expectations
func (ctx *TestContext) ValidateTestResults() error {
	if !ctx.Config.ValidateConsistency {
		return nil
	}

	if ctx.TestDataSet == nil {
		return fmt.Errorf("no test data to validate")
	}

	// Validate CI data consistency
	for _, ci := range ctx.TestDataSet.CIs {
		errors := ValidateAmortizableCI(ci)
		if len(errors) > 0 {
			return fmt.Errorf("CI validation failed for %s: %v", ci.ID, errors)
		}
	}

	// Validate ledger entry consistency
	for _, entry := range ctx.TestDataSet.LedgerEntries {
		errors := ValidateAmortizationEntry(entry)
		if len(errors) > 0 {
			return fmt.Errorf("Ledger entry validation failed for %s: %v", entry.ID, errors)
		}

		// Validate that CI exists for the entry
		if _, exists := ctx.TestDataSet.CIs[entry.CIID]; !exists {
			return fmt.Errorf("Ledger entry references non-existent CI: %s", entry.CIID)
		}
	}

	// Validate amortization run consistency
	for _, run := range ctx.TestDataSet.AmortizationRuns {
		errors := ValidateAmortizationRun(run)
		if len(errors) > 0 {
			return fmt.Errorf("Amortization run validation failed for %s: %v", run.ID, errors)
		}
	}

	return nil
}

// MemoryTestDatabase provides an in-memory test database implementation
type MemoryTestDatabase struct {
	data map[string]interface{}
	tx   *MemoryTestTransaction
}

// NewMemoryTestDatabase creates a new memory test database
func NewMemoryTestDatabase() *MemoryTestDatabase {
	return &MemoryTestDatabase{
		data: make(map[string]interface{}),
	}
}

func (db *MemoryTestDatabase) Setup() error {
	// Initialize memory database
	return nil
}

func (db *MemoryTestDatabase) Cleanup() error {
	db.data = make(map[string]interface{})
	db.tx = nil
	return nil
}

func (db *MemoryTestDatabase) BeginTx(ctx context.Context) (TestTransaction, error) {
	db.tx = &MemoryTestTransaction{
		db:  db,
		data: make(map[string]interface{}),
	}
	return db.tx, nil
}

func (db *MemoryTestDatabase) SeedData(ctx context.Context, dataSet *TestDataSet) error {
	if dataSet == nil {
		return nil
	}

	// Store test data in memory
	db.data["cis"] = dataSet.CIs
	db.data["ledger_entries"] = dataSet.LedgerEntries
	db.data["amortization_runs"] = dataSet.AmortizationRuns
	db.data["summaries"] = dataSet.Summaries

	return nil
}

func (db *MemoryTestDatabase) ValidateData(ctx context.Context) error {
	// Basic data integrity checks
	if db.data == nil {
		return fmt.Errorf("database not initialized")
	}

	return nil
}

// MemoryTestTransaction provides an in-memory test transaction
type MemoryTestTransaction struct {
	db  *MemoryTestDatabase
	data map[string]interface{}
}

func (tx *MemoryTestTransaction) Commit() error {
	// Copy transaction data to main database
	for key, value := range tx.data {
		tx.db.data[key] = value
	}
	tx.db.tx = nil
	return nil
}

func (tx *MemoryTestTransaction) Rollback() error {
	tx.data = make(map[string]interface{})
	tx.db.tx = nil
	return nil
}

func (tx *MemoryTestTransaction) Exec(query string, args ...interface{}) error {
	// Mock execution - would parse SQL in real implementation
	return nil
}

func (tx *MemoryTestTransaction) Query(query string, args ...interface{}) (TestRows, error) {
	// Mock query - would parse SQL in real implementation
	return &MemoryTestRows{}, nil
}

// MemoryTestRows provides mock database rows
type MemoryTestRows struct {
	rows [][]interface{}
	pos  int
}

func (r *MemoryTestRows) Next() bool {
	r.pos++
	return r.pos <= len(r.rows)
}

func (r *MemoryTestRows) Scan(dest ...interface{}) error {
	if r.pos > len(r.rows) {
		return fmt.Errorf("no more rows")
	}

	row := r.rows[r.pos-1]
	if len(row) != len(dest) {
		return fmt.Errorf("column count mismatch")
	}

	for i, val := range row {
		switch d := dest[i].(type) {
		case *string:
			*d = val.(string)
		case *int:
			*d = val.(int)
		case *float64:
			*d = val.(float64)
		case *bool:
			*d = val.(bool)
		case *[]byte:
			*d = val.([]byte)
		default:
			return fmt.Errorf("unsupported scan type")
		}
	}

	return nil
}

func (r *MemoryTestRows) Close() error {
	r.pos = 0
	return nil
}

// MemoryTestLogger provides an in-memory test logger
type MemoryTestLogger struct {
	logs []TestLogEntry
}

// NewMemoryTestLogger creates a new memory test logger
func NewMemoryTestLogger() *MemoryTestLogger {
	return &MemoryTestLogger{
		logs: make([]TestLogEntry, 0),
	}
}

func (l *MemoryTestLogger) Info(msg string, fields ...interface{}) {
	l.logs = append(l.logs, TestLogEntry{
		Timestamp: time.Now(),
		Level:     "INFO",
		Message:   msg,
	})
}

func (l *MemoryTestLogger) Error(msg string, fields ...interface{}) {
	l.logs = append(l.logs, TestLogEntry{
		Timestamp: time.Now(),
		Level:     "ERROR",
		Message:   msg,
	})
}

func (l *MemoryTestLogger) Debug(msg string, fields ...interface{}) {
	l.logs = append(l.logs, TestLogEntry{
		Timestamp: time.Now(),
		Level:     "DEBUG",
		Message:   msg,
	})
}

func (l *MemoryTestLogger) Warn(msg string, fields ...interface{}) {
	l.logs = append(l.logs, TestLogEntry{
		Timestamp: time.Now(),
		Level:     "WARN",
		Message:   msg,
	})
}

func (l *MemoryTestLogger) GetLogs() []TestLogEntry {
	return l.logs
}

func (l *MemoryTestLogger) ClearLogs() {
	l.logs = make([]TestLogEntry, 0)
}