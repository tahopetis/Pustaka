# Code Concerns & Technical Debt

**Analysis Date:** 2026-02-20

## High Priority Issues

### Unimplemented Amortization Module
**Status:** Planned but incomplete
**Impact:** High - Dashboard errors, missing financial tracking
**Location:** `/internal/amortization/`, `/docs/amortization/`
**Details:**
- Frontend components reference amortization API that returns 404
- Dashboard gracefully handles errors but feature is non-functional
- PRD and implementation plan exist in `/docs/amortization/`
- Phase 1 database migrations complete but backend service not integrated
- Recent commits show graceful error handling added to prevent crashes

**Remediation:** Implement amortization module per documented PRD, or remove references if postponed

### Large File Complexity
**Status:** Active technical debt
**Impact:** Medium - Maintenance difficulty
**Files:**
- `internal/amortization/repository.go` (1,994 lines)
- `internal/api/handlers/amortization_tests/amortization_contract_tests.go` (1,990 lines)
- `internal/ci/repository.go` (1,961 lines)

**Concern:** These files exceed maintainability thresholds and should be refactored into smaller modules

**Remediation:** Split repositories by domain (e.g., amortization queries vs. mutations), extract test helpers

## Medium Priority Issues

### Duplicate Migration Files
**Status:** Confusion risk
**Impact:** Low - Schema management
**Location:** `cmd/migrations/`
**Details:**
- `006_add_amortization.sql` (3,210 bytes)
- `006_add_amortization_module.sql` (18,030 bytes)

**Concern:** Two files with similar names and purposes may cause migration confusion

**Remediation:** Consolidate or rename to clarify purpose and ordering

### Test Coverage Gaps
**Status:** Incomplete testing
**Impact:** Medium - Regression risk
**Details:**
- Only 9 test files found in Go codebase
- Frontend has 18 test files (unit + E2E)
- No formal coverage targets enforced
- Large repository files with minimal test coverage

**Remediation:** Increase test coverage, especially for complex business logic in amortization and CI services

### Inconsistent Error Handling
**Status:** Partially addressed
**Impact:** Medium - User experience
**Details:**
- Recent fixes added graceful handling for 404/network errors
- Some components still show console errors for unimplemented features
- Not all error paths provide user-friendly messages

**Remediation:** Audit all error handling paths, ensure consistent UX

## Low Priority Issues

### Documentation Maintenance
**Status:** Documentation drift possible
**Impact:** Low - Onboarding overhead
**Details:**
- Extensive documentation in `/docs/` but may not reflect recent changes
- Recent cleanup removed 63,461 lines of code/docs
- Some CLAUDE.md files may be outdated

**Remediation:** Review and update documentation to match current codebase state

### Neo4j Configuration
**Status:** Recently optimized
**Impact:** Low - Performance on resource-constrained systems
**Location:** `internal/ci/neo4j_service.go`
**Details:**
- Memory configuration adjusted for 8GB systems
- Default settings may still be high for smaller deployments

**Remediation:** Document configuration requirements and tuning guidelines

### Mixed Test Frameworks
**Status:** Technical debt
**Impact:** Low - Maintenance complexity
**Details:**
- Frontend uses both Cypress and Playwright for E2E testing
- Duplicate E2E test suites increase maintenance burden

**Remediation:** Standardize on one E2E framework (Playwright recommended)

## Security Considerations

### Configuration Security
**Status:** Good practices
**Details:**
- Strong JWT secrets enforced in production
- Environment variables for sensitive data
- Default secrets flagged in validation

### Input Validation
**Status:** Implemented
**Details:**
- RBAC with granular permissions
- Audit logging for all changes
- Password hashing with Argon2ID

## Performance Concerns

### Database Query Complexity
**Status:** Monitored
**Location:** `internal/ci/repository.go`, `internal/amortization/repository.go`
**Details:**
- Complex amortization calculations in SQL
- Large repository files may indicate N+1 query risks
- Redis caching implemented for frequently accessed data

**Remediation:** Profile query performance, add indexes as needed

### Concurrent Operations
**Status:** Good practices
**Details:**
- Goroutines used for concurrent data fetching
- Connection pooling configured
- No obvious race conditions in recent code

## Architectural Debt

### Module Boundaries
**Status:** Generally clean
**Details:**
- Clear layering (API → Service → Repository)
- Some circular dependencies possible in complex workflows
- Amortization module separation is good

### Frontend State Management
**Status:** Acceptable
**Details:**
- Pinia stores well organized
- Some components may have excessive prop drilling
- Consider additional composables for shared logic

## Recommendations

**Immediate Actions:**
1. Complete amortization module implementation or remove frontend references
2. Refactor large repository files (>1,500 lines)
3. Consolidate duplicate migration files

**Short-term (1-2 weeks):**
4. Increase test coverage to 60%+ for critical paths
5. Standardize on single E2E framework
6. Update documentation to match codebase state

**Long-term (1-3 months):**
7. Implement formal coverage targets
8. Performance profiling and optimization
9. Security audit and penetration testing

---

*Concerns analysis: 2026-02-20*
