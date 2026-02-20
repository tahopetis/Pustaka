# Coding Conventions

**Analysis Date:** 2026-02-20

## Naming Patterns

### Go Files
**Files:**
- `snake_case.go` - All Go source files use snake_case (e.g., `ci_handlers.go`, `auth_service.go`)
- `*_test.go` - Test files follow Go standard naming convention
- `*_test_*.go` - Integration tests use this pattern (e.g., `amortization_integration_test.go`)

**Functions:**
- `ExportedFunction()` - Exported functions use PascalCase for external visibility
- `privateFunction()` - Private/internal functions use snake_case
- Handler methods: `Create()`, `Update()`, `Delete()` - Simple, action-oriented

**Variables:**
- `exportedVariable` - Exported variables use PascalCase
- `privateVariable` - Private variables use snake_case
- Function parameters: Descriptive names like `ctx context.Context`, `userID uuid.UUID`

**Types:**
- `StructName` - All structs use PascalCase (e.g., `ConfigurationItem`, `AuthHandler`)
- Interface names: `Service`, `Repository` - Suffix indicates purpose
- Request/response types: `Request`, `Response` - Clear distinction

### Frontend (Vue/TypeScript)
**Files:**
- `PascalCase.vue` - Component files use PascalCase (e.g., `DashboardWidget.vue`)
- `camelCase.ts` - Store and service files use camelCase (e.g., `auth.ts`, `api.ts`)
- `kebab-case` - Route files use kebab-case (e.g., `auth-form.ts`)

**Functions:**
- `useStore()` - Composables prefixed with `use`
- `componentMethod()` - Component methods use camelCase
- Handler functions: `handleSubmit()`, `fetchData()`

**Variables:**
- `camelCase` - All variables and properties use camelCase
- Store variables: `user`, `isAuthenticated`, `accessToken`
- Props: `title`, `isVisible`, `items`

**Interfaces/Types:**
- `interface InterfaceName` - TypeScript interfaces use PascalCase
- `TypeName` - Custom types use PascalCase
- Generic types: `T`, `K` - Single letters for simple generics

## Code Style

### Go
**Formatting:**
- Tool: `go fmt ./...` - Standard Go formatter enforced via Makefile
- Indentation: Tabs for indentation (Go default)
- Brace placement: K&R style for structs, else style for control flow
- Line length: Generally kept under 100 characters

**Linting:**
- Tool: `golangci-lint` - Comprehensive linter via Makefile
- Key rules enforced:
  - golint: Naming conventions
  - govet: Static analysis
  - gosec: Security checks
  - ineffassign: Unused variable detection

**Error Handling:**
- Pattern: `if err != nil { return err }` - Consistent error checking
- Wrapping: `fmt.Errorf("context: %w", err)` - Error wrapping with context
- Logging: `h.logger.Error().Err(err).Msg("operation failed")` - Structured logging with zerolog

### Frontend
**Formatting:**
- Tool: Prettier - Standard code formatter via npm scripts
- Indentation: 2 spaces (consistent with Vue standards)
- Line length: 80-100 characters (configurable)

**Linting:**
- Tool: ESLint - Configured with Vue and TypeScript rules
- Key rules:
  - No unused variables
  - Component prop validation
  - TypeScript type checking
  - Vue component style rules

**Vue Style Guide:**
- Single file components: `<template>`, `<script>`, `<style>` in logical order
- Component name: Multi-word PascalCase (e.g., `LoginButton`)
- Template: `v-bind:prop` → `:prop` shorthand
- Props: Default values and validation

## Import Organization

### Go
**Order:**
1. Standard library packages (alphabetical)
2. External packages (alphabetical by import path)
3. Internal packages (alphabetical by package path)

**Pattern:**
```go
import (
    "context"
    "fmt"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"
    "github.com/rs/zerolog"

    "github.com/pustaka/pustaka/internal/auth"
    "github.com/pustaka/pustaka/internal/api/middleware"
    pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)
```

**Internal imports:**
- Aliases used for clarity (e.g., `pustakaLogger`)
- `internal/` prefix for all internal packages

### Frontend
**Order:**
1. Vue core imports
2. External library imports
3. Internal component/store imports
4. Type imports

**Pattern:**
```typescript
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import type { User } from '@/types/user'
```

**Path aliases:**
- `@/` - Points to `web/src/` directory
- Relative imports avoided for better tree-shaking

## Error Handling

### Go
**Patterns:**
```go
// Structured error responses
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.logger.Error().Err(err).Msg("Invalid request body")
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Business logic with error handling
    user, err := h.authService.Authenticate(ctx, req.Username, req.Password)
    if err != nil {
        h.logger.Error().Err(err).Msg("Authentication failed")
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }
}
```

**Logging:**
- Structured logging with zerolog
- Contextual information included
- No sensitive data in logs

### Frontend
**Patterns:**
```typescript
// Error handling with try/catch
try {
    const response = await api.post('/auth/login', credentials)
    if (response.data.error) {
      throw new Error(response.data.error)
    }
    return response.data
  } catch (error) {
    toast.error(error.message || 'Login failed')
    throw error
  }

// Error boundaries for Vue components
onErrorCaptured(() => {
  toast.error('An error occurred')
  return false // Prevent error from propagating
})
```

**User feedback:**
- Toast notifications for errors
- Loading states during async operations
- Graceful degradation for API failures

## Comments

### Go
**When to Comment:**
- Public APIs (documentation required)
- Complex business logic
- Algorithm explanations
- Important gotchas or limitations

**JSDoc/TSDoc:**
```go
// CreateCI creates a new configuration item with validation
// Returns the created CI or error if validation fails
func (s *CIService) CreateCI(ctx context.Context, req CreateCIRequest) (*ConfigurationItem, error) {
    // Implementation
}
```

### Frontend
**JSDoc/TSDoc:**
```typescript
/**
 * Authenticates user with provided credentials
 * @param credentials - User login credentials
 * @returns Promise resolving to auth response
 * @throws Error if authentication fails
 */
export const authenticateUser = async (credentials: LoginCredentials): Promise<AuthResponse> => {
    // Implementation
}
```

## Function Design

### Go
**Size:** Single-purpose functions, generally < 50 lines
**Parameters:** 3-5 parameters maximum, use structs for complex data
**Return Values:** Explicit error returns, nil for success

```go
// Good: Single responsibility
func (s *CIService) ValidateCI(req CreateCIRequest) error {
    // Validation logic
}

// Good: Structured return
func (s *CIService) GetCIByID(ctx context.Context, id uuid.UUID) (*ConfigurationItem, error) {
    // Business logic
}
```

### Frontend
**Size:** Component methods kept concise, < 30 lines
**Parameters:** Destructured props, default values
**Composition:** Prefer composables for shared logic

```typescript
// Good: Composable for shared logic
export const useAuthStore = defineStore('auth', () => {
  // Store implementation
})

// Good: Component method
const handleSubmit = async () => {
  await authStore.login(credentials)
}
```

## Module Design

### Go
**Exports:** Minimal exports, use unexported types for internal implementation
**Packages:** Clear boundaries, internal/ for implementation details
**Dependencies:** Explicit dependency injection

**File organization:**
```go
// package handlers
// package services
// package repositories
// Clear separation of concerns
```

### Frontend
**Exports:** Named exports for functions, default for components
**Composition:** Prefer composables over mixins
**Tree-shaking:** Keep exports minimal

```typescript
// Good: Named exports
export const useAuthStore = defineStore('auth', () => { ... })
export type { User, LoginCredentials }

// Good: Default component export
export default LoginForm
```

---

*Convention analysis: 2026-02-20*