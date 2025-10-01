# 🧪 Phase-by-Phase Test Plan (Revised)

**Project**: Pustaka CMDB
**Version**: 2.0
**Date**: 2025-10-01
**Author**: Tahopetis  

---

## 🎯 Purpose
This test plan defines testing scope for each implementation phase.  
- Certain tests (**unit, integration, regression**) run in *every phase*.  
- Other tests are **phase-specific**, aligned with the Implementation Plan.  
- A **full suite** (including performance, chaos, penetration, DR) runs before final release.  

---

## 🔹 Tests Always Run in Every Phase
- **Unit tests**: For all new code introduced in the phase.  
- **Integration tests**: For APIs, DB, and services touched in the phase.  
- **Regression tests**: Re-run earlier test cases to ensure nothing breaks.  

---

## 🔹 Phase 1 – Foundation & FSD-Compliant Core Features (Weeks 1-6)
- Always-run tests (unit, integration, regression).
- **FSD Compliance Tests**:
  - JSONB schema validation (user-defined CI type schemas)
  - CI type schema management CRUD operations
  - Flexible attributes storage and retrieval
  - Schema validation error handling
- **Functional Tests**:
  - CI CRUD APIs with flexible attributes
  - Basic RBAC rules (Admin, User roles)
  - Relationship sync between PostgreSQL and Neo4j
- **Performance Baseline Tests**:
  - JSONB attribute query performance (< 200ms)
  - Basic graph performance (< 500 nodes)
- **Audit Tests**: Basic audit logging for all CRUD operations

---

## 🔹 Phase 2 – Enhanced Features & Performance (Weeks 7-12)
- Always-run tests (unit, integration, regression).
- **Advanced RBAC Tests**:
  - Granular permissions (ci:create, ci:read, ci:update, ci:delete)
  - Role hierarchy and permission inheritance
  - CI type-specific permissions
- **Security Tests**:
  - JWT authentication and token refresh
  - Role-based access enforcement
  - SQL injection prevention
  - Input validation with flexible schemas
- **Import/Export Tests**:
  - CSV file processing and validation
  - Column mapping with flexible attributes
  - Bulk operations performance
- **Graph Performance Tests**:
  - Graph visualization up to 5k nodes (< 3s load time)
  - Multiple layout algorithms
  - Graph clustering performance

---

## 🔹 Phase 3 – Enterprise Features & Optimization (Weeks 13-20)
- Always-run tests (unit, integration, regression).
- **Comprehensive Audit Tests**:
  - Complete audit trail for all operations
  - Audit log search and filtering
  - Audit log immutability
  - Compliance reporting accuracy
- **Integration Tests**:
  - External connectors (HR, monitoring tools)
  - Data synchronization across systems
  - Error handling and retry mechanisms
- **Large-Scale Performance Tests**:
  - 50k+ CIs performance validation
  - Neo4j query optimization
  - PostgreSQL JSONB optimization at scale
  - Cache performance and consistency
- **Monitoring & Health Tests**:
  - Health check endpoints
  - Metrics collection accuracy
  - Alert system functionality

---

## 🔹 Phase 4 – Production Readiness & Launch (Weeks 21-24)
- Always-run tests (unit, integration, regression).
- **Load & Stress Tests**:
  - Production-scale load testing (500+ concurrent users)
  - API throughput validation
  - Database connection pooling under load
- **Chaos Engineering Tests**:
  - Database node failures
  - Network partitions between services
  - Cache failures and recovery
- **Security Hardening Tests**:
  - Penetration testing
  - RBAC bypass attempts
  - API fuzzing and vulnerability scanning
- **Disaster Recovery Tests**:
  - Backup and restore procedures
  - Failover scenarios
  - Data consistency after recovery
- **User Acceptance Tests**:
  - End-to-end business workflows
  - Cross-browser compatibility
  - User experience validation

---

## 🔹 Final Release Readiness (Pre-Launch)
Run **full suite end-to-end**, including:  
- All functional, compliance, and integration tests.  
- Load and stress tests at expected production scale.  
- Security penetration testing.  
- Business continuity and DR scenario testing.  

---

## ✅ Summary
- **Unit, integration, regression = always** for every phase.
- **FSD compliance testing starts in Phase 1** - flexible schemas are foundational.
- **Phase-specific tests** focus on features delivered in that phase.
- **Performance testing is progressive** - baseline in Phase 1, scale in Phase 3, production in Phase 4.
- **Security testing builds incrementally** - basic auth in Phase 1, comprehensive in Phase 2, hardening in Phase 4.
- **Full exhaustive testing** is performed before production readiness.
- This ensures each phase delivers safe increments without unnecessary overhead while maintaining alignment with the 4-phase implementation plan.  

