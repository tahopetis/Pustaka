# Amortization Module Testing Guide

## Overview

This document provides comprehensive guidance for testing the IT Asset Amortization Module API endpoints. The test suite includes contract validation, integration testing, performance testing, and business logic validation.

## Test Architecture

### Test Components

The amortization test suite consists of several key components:

1. **Contract Tests** (`amortization_contract_test.go`)
   - Validates API compliance with OpenAPI specification
   - Tests request/response schemas
   - Verifies HTTP status codes
   - Validates error handling

2. **Integration Tests** (`amortization_integration_test.go`)
   - End-to-end workflow testing
   - Multi-endpoint scenario testing
   - Data consistency validation
   - Realistic usage patterns

3. **Mock Services** (`mock_amortization_service.go`)
   - Isolated unit testing
   - Controlled test data
   - Simulated error conditions
   - Performance benchmarking

4. **Test Utilities** (`amortization_test_utils.go`)
   - Test data generation
   - Validation helpers
   - Business rule checking
   - Data factory patterns

5. **Test Configuration** (`amortization_test_config.go`)
   - Test environment setup
   - Configuration management
   - Authentication simulation
   - Database abstraction

## Running Tests

### Prerequisites

- Go 1.19 or higher
- Access to the Pustaka codebase
- Required Go modules:
  ```bash
  go mod tidy
  ```

### Test Execution Commands

#### Run All Tests
```bash
# Run all amortization tests
go test ./internal/api/handlers -run Amortization

# Run with verbose output
go test ./internal/api/handlers -v -run Amortization

# Run with coverage
go test ./internal/api/handlers -cover -run Amortization
```

#### Specific Test Suites
```bash
# Contract tests only
go test ./internal/api/handlers -run AmortizationContractTestSuite

# Integration tests only
go test ./internal/api/handlers -run AmortizationIntegrationTestSuite

# Performance benchmarks
go test ./internal/api/handlers -bench=BenchmarkAmortization
```

#### Test Configuration
```bash
# Run with custom configuration
go test ./internal/api/handlers -run Amortization -config=test-config.json

# Run with verbose output and benchmarks
go test ./internal/api/handlers -v -bench=. -run Amortization
```

### Makefile Integration

Add to your `Makefile`:

```makefile
# Amortization tests
test-amortization:
	@echo "Running amortization module tests..."
	go test ./internal/api/handlers -v -run Amortization -cover

test-amortization-contract:
	@echo "Running amortization contract tests..."
	go test ./internal/api/handlers -v -run AmortizationContractTestSuite

test-amortization-integration:
	@echo "Running amortization integration tests..."
	go test ./internal/api/handlers -v -run AmortizationIntegrationTestSuite

test-amortization-performance:
	@echo "Running amortization performance tests..."
	go test ./internal/api/handlers -bench=BenchmarkAmortization -run Amortization

test-amortization-coverage:
	@echo "Running amortization tests with coverage..."
	go test ./internal/api/handlers -coverprofile=amortization-coverage.out -run Amortization
	go tool cover -html=amortization-coverage.out -o amortization-coverage.html
```

## Test Coverage

### API Endpoints Covered

| Endpoint | Method | Contract Tests | Integration Tests |
|----------|--------|----------------|-------------------|
| `/amortization/configuration-items` | GET | ✅ | ✅ |
| `/amortization/configuration-items/{ciId}` | GET | ✅ | ✅ |
| `/amortization/configuration-items/{ciId}` | PUT | ✅ | ✅ |
| `/amortization/ledger` | GET | ✅ | ✅ |
| `/amortization/ledger/{entryId}` | GET | ✅ | ✅ |
| `/amortization/adjustments` | POST | ✅ | ✅ |
| `/amortization/runs` | GET | ✅ | ✅ |
| `/amortization/runs` | POST | ✅ | ✅ |
| `/amortization/runs/{runId}` | GET | ✅ | ✅ |
| `/amortization/summaries` | GET | ✅ | ✅ |
| `/amortization/reports/depreciation-schedule` | GET | ✅ | ✅ |

### Test Scenarios

#### 1. Happy Path Scenarios
- Valid request/response handling
- Successful data creation and retrieval
- Proper pagination and filtering
- Correct sorting and searching

#### 2. Error Handling Scenarios
- Invalid input validation
- Missing required fields
- Resource not found
- Unauthorized access
- Permission denied
- Rate limiting
- Server errors

#### 3. Edge Cases
- Maximum/minimum values
- Unicode characters
- Special characters in data
- Empty datasets
- Concurrent access
- Network timeouts

#### 4. Business Logic Validation
- Depreciation calculations
- Book value consistency
- Adjustment entry validation
- Financial rounding rules
- Date-based calculations

## Test Data Management

### Test Data Factory

The `TestDataFactory` provides methods for generating realistic test data:

```go
factory := NewTestDataFactory()

// Create test CI
ci := factory.CreateTestCI()

// Create test ledger entry
entry := factory.CreateTestLedgerEntry(ciID)

// Create test amortization run
run := factory.CreateTestAmortizationRun()

// Create complete test data set
dataSet := factory.CreateTestDataSet(50, 200, 10)
```

### Data Validation

Built-in validation functions ensure data integrity:

```go
// Validate CI business rules
errors := ValidateAmortizableCI(ci)

// Validate ledger entry consistency
errors := ValidateAmortizationEntry(entry)

// Validate amortization run status
errors := ValidateAmortizationRun(run)
```

## Performance Testing

### Benchmarks

Performance benchmarks validate response time requirements:

```bash
go test -bench=BenchmarkAmortization -benchmem
```

Key performance metrics:
- List amortizable CIs: < 100ms
- Get CI details: < 50ms
- Create adjustment: < 75ms
- Generate report: < 500ms

### Concurrent Testing

Tests validate concurrent access patterns:

```go
func TestConcurrentAccess(t *testing.T) {
    concurrency := 50
    done := make(chan bool, concurrency)

    for i := 0; i < concurrency; i++ {
        go func(id int) {
            // Make concurrent requests
            done <- true
        }(i)
    }

    // Wait for completion
    for i := 0; i < concurrency; i++ {
        <-done
    }
}
```

## Authentication & Authorization Testing

### Test Users

The test suite includes predefined test users:

```go
// Admin user with full permissions
adminUser := &TestUser{
    Role: "admin",
    Permissions: []string{
        "amortization:read", "amortization:write",
        "amortization:adjust", "amortization:admin",
    },
}

// Regular user with limited permissions
regularUser := &TestUser{
    Role: "user",
    Permissions: []string{
        "amortization:read", "amortization:write",
        "amortization:adjust",
    },
}

// Read-only user
readOnlyUser := &TestUser{
    Role: "readonly",
    Permissions: []string{"amortization:read"},
}
```

### Permission Testing

Tests validate role-based access control:

```go
func TestPermissions(t *testing.T) {
    tests := []struct {
        user         *TestUser
        endpoint     string
        method       string
        expectedCode int
    }{
        {
            user:         adminUser,
            endpoint:     "/amortization/runs",
            method:       "POST",
            expectedCode: http.StatusAccepted,
        },
        {
            user:         readOnlyUser,
            endpoint:     "/amortization/runs",
            method:       "POST",
            expectedCode: http.StatusForbidden,
        },
    }

    for _, tt := range tests {
        // Test permission-based access
    }
}
```

## Contract Validation

### OpenAPI Compliance

Tests validate API compliance with the OpenAPI specification:

- Request/response schema validation
- Required field enforcement
- Data type validation
- HTTP status code correctness
- Error response format

### Schema Validation Examples

```go
func TestAmortizationCISchema(t *testing.T) {
    response := suite.makeRequest("GET", "/amortization/configuration-items")

    // Validate response structure
    expectedFields := []string{
        "cis", "page", "limit", "total", "total_pages",
    }

    for _, field := range expectedFields {
        assert.Contains(t, response, field)
    }

    // Validate CI structure
    cis := response["cis"].([]interface{})
    for _, ci := range cis {
        suite.validateAmortizableCIFields(ci.(map[string]interface{}))
    }
}
```

## Integration Testing

### Workflow Testing

Integration tests validate complete user workflows:

```go
func TestCompleteAmortizationWorkflow(t *testing.T) {
    // 1. Create amortizable CI
    ci := suite.createTestAmortizableCI()

    // 2. Update financial configuration
    suite.updateAmortizationConfig(ci.ID, financialConfig)

    // 3. Create manual adjustment
    adjustment := suite.createAdjustment(ci.ID, amount, reason)

    // 4. Trigger amortization run
    run := suite.triggerManualRun([]uuid.UUID{ci.ID})

    // 5. Verify ledger entries
    ledger := suite.getLedgerEntries(ci.ID)

    // 6. Generate reports
    report := suite.generateDepreciationReport(ci.ID)

    // Validate end-to-end consistency
    suite.validateWorkflowConsistency(ci, adjustment, run, ledger, report)
}
```

## Continuous Integration

### CI/CD Pipeline Integration

Add to your CI pipeline:

```yaml
# .github/workflows/test.yml
name: Amortization Module Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test-amortization:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v3

    - name: Setup Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.19

    - name: Run Amortization Tests
      run: |
        make test-amortization
        make test-amortization-coverage

    - name: Upload Coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./amortization-coverage.out
```

### Quality Gates

Set quality gates for test metrics:

- **Test Coverage**: > 85%
- **Test Success Rate**: 100%
- **Performance**: All benchmarks under threshold
- **Contract Compliance**: 100% OpenAPI validation

## Test Configuration

### Environment Configuration

Create test configuration file:

```json
{
  "database_url": "memory://test",
  "use_memory_db": true,
  "jwt_secret": "test-secret-key",
  "num_test_cis": 50,
  "num_test_ledger_entries": 200,
  "num_test_runs": 10,
  "max_response_time": "100ms",
  "strict_validation": true,
  "validate_calculations": true,
  "validate_consistency": true
}
```

### Database Configuration

For integration tests with real database:

```go
func setupTestDB() *sql.DB {
    db, err := sql.Open("postgres", "postgres://test:test@localhost/pustaka_test?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }

    // Run migrations
    migrate(db)

    return db
}
```

## Troubleshooting

### Common Issues

1. **Test Data Inconsistency**
   ```bash
   # Reset test data
   go test ./internal/api/handlers -run Amortization -test.v -test.reset
   ```

2. **Performance Test Failures**
   ```bash
   # Run with increased timeout
   go test ./internal/api/handlers -run Amortization -timeout 30m
   ```

3. **Authentication Failures**
   ```bash
   # Check test user configuration
   go test ./internal/api/handlers -run Amortization -test.auth.debug
   ```

### Debug Tips

1. **Enable Verbose Logging**
   ```bash
   go test ./internal/api/handlers -v -run Amortization -test.log.level=debug
   ```

2. **Generate Test Report**
   ```bash
   go test ./internal/api/handlers -run Amortization -test.report.output=test-report.html
   ```

3. **Run Single Test**
   ```bash
   go test ./internal/api/handlers -run TestAmortizationDetails_Contract/Valid_CI_ID -v
   ```

## Best Practices

### Test Organization

1. **Use Test Suites**: Group related tests in suite structures
2. **Shared Setup**: Use `SetupSuite` and `TearDownSuite` for expensive operations
3. **Test Isolation**: Ensure tests don't interfere with each other
4. **Descriptive Names**: Use clear, descriptive test names

### Data Management

1. **Deterministic Data**: Use predictable test data
2. **Cleanup Resources**: Clean up database connections and files
3. **Mock External Services**: Isolate tests from external dependencies
4. **Validate Business Rules**: Ensure data follows business constraints

### Performance Considerations

1. **Benchmark Critical Paths**: Focus on performance-critical operations
2. **Concurrent Testing**: Test concurrent access patterns
3. **Resource Limits**: Validate behavior under load
4. **Memory Usage**: Monitor memory consumption in tests

## Maintenance

### Regular Updates

1. **Update Test Data**: Refresh test data to reflect production patterns
2. **Review Coverage**: Maintain or improve test coverage
3. **Performance Baselines**: Update performance expectations
4. **Security Testing**: Add security-related test scenarios

### Documentation

1. **Test Documentation**: Keep test documentation current
2. **API Changes**: Update tests for API changes
3. **Business Rules**: Document business rule validations
4. **Test Examples**: Provide clear test examples for new developers

## Conclusion

This comprehensive test suite ensures the reliability and correctness of the IT Asset Amortization Module API. Regular execution of these tests helps maintain code quality, prevents regressions, and ensures compliance with the OpenAPI specification.

For questions or contributions to the test suite, please refer to the project's contribution guidelines or contact the development team.