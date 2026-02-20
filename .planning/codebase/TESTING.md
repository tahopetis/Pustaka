# Testing Patterns

**Analysis Date:** 2026-02-20

## Test Framework

### Backend (Go)
**Runner:**
- Framework: `go test` - Standard Go testing framework
- Config: Integrated with Makefile (`make test`)
- Runner: `go test -v ./...` for verbose output

**Assertion Library:**
- Package: `github.com/stretchr/testify`
- Key components: `assert`, `require`, `mock`

**Test Commands:**
```bash
make test              # Run all tests
go test ./...          # Run tests directly
go test -v ./internal/ci/  # Run specific package tests
go test -cover ./...  # Run with coverage report
```

**Coverage:**
- Tool: Built-in Go coverage
- Target: Not enforced at package level
- View: `go test -cover ./...`

### Frontend (Vue.js)
**Runner:**
- Framework: Vitest - Modern testing framework for Vite
- Config: No explicit config file detected (uses defaults)
- Runner: `npm test` or `vitest`

**Assertion Library:**
- Package: `@vue/test-utils` with built-in assertions
- Additional: `vi` from vitest for mocking

**Test Commands:**
```bash
npm test                 # Run all tests
npm run test:ui         # Run tests with UI
npm run test:e2e        # Run Cypress E2E tests
npm run test:e2e:playwright # Run Playwright E2E tests
```

## Test File Organization

### Backend
**Location:** Co-located with source files (`*_test.go`)
**Naming:** `*_test.go` for unit tests, `*_integration_test.go` for integration tests
**Structure:**
```
internal/
├── api/
│   ├── handlers/
│   │   ├── auth_test.go
│   │   └── ci_handlers_test.go
│   ├── api_test.go
│   └── relationship_type_integration_test.go
├── ci/
│   ├── service_test.go
│   └── relationship_types_service_test.go
└── auth/
    └── rbac_test.go
```

### Frontend
**Location:** `web/tests/` directory
**Naming:** `*.test.ts` pattern
**Structure:**
```
web/tests/
├── auth.test.ts
├── ci.test.ts
├── AttributeEditor.test.ts
├── BaseSelect.test.ts
└── components/
    └── CIListView.test.ts
```

## Test Structure

### Backend Test Patterns
**Suite Organization:**
```go
package auth

import (
    "testing"
    "context"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/suite"
)

type RBACSuite struct {
    suite.Suite
    db   *pgxpool.Pool
    rbac *RBACService
    cleanup func()
}

func (suite *RBACSuite) SetupSuite() {
    suite.db, suite.cleanup = testutils.SetupTestDB(suite.T())
    suite.rbac = NewRBACService(suite.db)
}

func (suite *RBACSuite) SetupTest() {
    testutils.CleanupDB(suite.T(), suite.db)
    testutils.InsertTestData(suite.T(), suite.db)
}

func (suite *RBACSuite) TearDownSuite() {
    suite.cleanup()
}

func TestRBACSuite(t *testing.T) {
    suite.Run(t, new(RBACSuite))
}
```

**Test Patterns:**
- Setup/TearDown methods for test isolation
- `require.NoError()` for critical failures
- `assert.Equal()` for value comparisons
- Context usage for timeouts/cancellation

### Frontend Test Patterns
**Suite Organization:**
```typescript
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAuthStore } from '@/stores/auth'
import axios from 'axios'

// Mock axios
vi.mock('axios')

describe('Auth Store', () => {
  let authStore: any

  beforeEach(() => {
    setActivePinia(createPinia())
    authStore = useAuthStore()
    localStorage.clear()
    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('should have correct initial state', () => {
      expect(authStore.user).toBeNull()
      expect(authStore.accessToken).toBeNull()
      expect(authStore.isAuthenticated).toBe(false)
    })
  })
})
```

**Patterns:**
- `describe()` for test grouping
- `beforeEach()` for setup
- `vi.mock()` for external dependencies
- `vi.clearAllMocks()` for cleanup

## Mocking

### Backend Mocking
**Framework:** `github.com/stretchr/testify/mock`
**Pattern:**
```go
type MockRBACService struct {
    mock.Mock
}

func (m *MockRBACService) CreateUser(ctx context.Context, username, email, hashedPassword string, roles []string) (*User, error) {
    args := m.Called(ctx, username, email, hashedPassword, roles)
    return args.Get(0).(*User), args.Error(1)
}

// In test:
mockService := new(MockRBACService)
mockService.On("CreateUser", mock.Anything, "testuser", "test@example.com", "hashed", []string{"viewer"}).
    Return(&User{ID: uuid.New(), Username: "testuser"}, nil)
```

**What to Mock:**
- External services (databases, APIs)
- Time-dependent operations
- File system operations
- Network calls

**What NOT to Mock:**
- Business logic (test it directly)
- Standard library functions
- Simple utilities

### Frontend Mocking
**Framework:** Viest mocking utilities
**Pattern:**
```typescript
// Mock axios
vi.mock('axios')
const mockedAxios = axios as any

// In test:
mockedAxios.post.mockResolvedValue({
  data: {
    access_token: 'mock-token',
    user: { id: '1', username: 'test' }
  }
})

// Mock localStorage
Object.defineProperty(window, 'localStorage', {
  value: {
    getItem: vi.fn(),
    setItem: vi.fn(),
    clear: vi.fn()
  }
})
```

## Fixtures and Factories

### Backend Test Data
**Location:** `internal/testutils/database.go`
**Pattern:**
```go
// TestDB holds database connection and cleanup function
type TestDB struct {
    Pool   *pgxpool.Pool
    Cleanup func()
}

// SetupTestDB creates a test database using Docker
func SetupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
    // Docker setup for test database
}

// CleanupDB cleans up test data
func CleanupDB(t *testing.T, db *pgxpool.Pool) {
    // Truncate tables for clean state
}

// InsertTestData inserts test data for tests
func InsertTestData(t *testing.T, db *pgxpool.Pool) {
    // Insert default test data
}
```

### Frontend Test Data
**Pattern:**
```typescript
const mockUser = {
  id: '550e8400-e29b-41d4-a716-446655440000',
  username: 'testuser',
  email: 'test@example.com',
  is_active: true,
  roles: [],
  permissions: []
}

const mockAuthResponse = {
  data: {
    access_token: 'mock-access-token',
    token_type: 'Bearer',
    expires_in: 3600,
    user: mockUser
  }
}
```

## Coverage

### Backend
**Requirements:** No formal coverage targets enforced
**View Coverage:**
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # Generate HTML report
```

**Typical Coverage Areas:**
- Handler validation logic
- Service layer business rules
- Repository CRUD operations
- Error handling paths

### Frontend
**Requirements:** Not enforced at package level
**View Coverage:**
```bash
npm run test -- --coverage  # If configured
```

**Test Types:**
**Unit Tests:**
- Component rendering and props
- Store state mutations
- Utility functions
- API layer (with mocks)

**Integration Tests:**
- Component interaction
- Store integration
- API integration (Docker services)

**E2E Tests:**
- Cypress: Browser automation
- Playwright: Modern browser automation
- Full user workflows

## Common Patterns

### Backend Async Testing
```go
func (s *CIService) TestAsyncOperation() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    result, err := s.AsyncOperation(ctx)
    require.NoError(s.T(), err)
    assert.NotNil(s.T(), result)
}
```

### Frontend Async Testing
```typescript
it('should login successfully and set tokens', async () => {
  await authStore.login(credentials)

  expect(authStore.user).toEqual(mockUser)
  expect(authStore.accessToken).toBe('mock-access-token')
})
```

### Error Testing
```go
func (s *CIService) TestErrorCases() {
    // Test invalid input
    _, err := s.CreateCI(ctx, CreateCIRequest{Name: ""})
    assert.Error(s.T(), err)

    // Test not found
    _, err = s.GetCIByID(ctx, uuid.Nil)
    assert.Error(s.T(), err)
    assert.Contains(s.T(), err.Error(), "not found")
}
```

### Component Testing
```typescript
import { mount } from '@vue/test-utils'
import Component from '@/components/Component.vue'

describe('Component', () => {
  it('renders correctly', () => {
    const wrapper = mount(Component, {
      props: { title: 'Test Title' }
    })
    expect(wrapper.text()).toContain('Test Title')
  })

  it('emits event on click', async () => {
    const wrapper = mount(Component)
    await wrapper.find('button').trigger('click')
    expect(wrapper.emitted()).toHaveProperty('click')
  })
})
```

---

*Testing analysis: 2026-02-20*