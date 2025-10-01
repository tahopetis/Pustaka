# 🔐 Security & Compliance Checklist

**Project**: Pustaka CMDB
**Version**: 1.0
**Date**: 2025-10-01
**Author**: Tahopetis  

---

## 🎯 Purpose  
This checklist defines the minimum security and compliance requirements for all code and infrastructure in the Pustaka CMDB project. It ensures that AI-generated code consistently applies security best practices and meets compliance standards (PCI-DSS, GDPR, banking IT governance).  

---

## 🔹 Authentication & Authorization  
- ✅ All endpoints must enforce **JWT authentication** middleware.  
- ✅ JWTs must use **HS256 or RS256** with configurable secret/keys via env variables.  
- ✅ Access control must be enforced with **RBAC** at the service layer.  
- ✅ Passwords must be hashed using **Argon2id** (no MD5, SHA1, or plain bcrypt).  
- ✅ Implement account lockout after **5 failed login attempts**.  
- ✅ Refresh tokens must be supported, stored securely, and revocable.  

---

## 🔹 Input Validation & Data Protection  
- ✅ All incoming requests must validate **payload schema** (Pydantic/Go structs).  
- ✅ No direct string concatenation for queries → **parameterized queries only** (Postgres, Neo4j).  
- ✅ File uploads must enforce **MIME type & size checks**.  
- ✅ Sensitive data (PII, credentials, tokens) must never be logged.  
- ✅ All secrets (DB, API keys, JWT keys) must come from **environment variables / Vault**, never hardcoded.  

---

## 🔹 API Security  
- ✅ All APIs must use **HTTPS** (TLS 1.2+).  
- ✅ CORS must be explicitly configured (allowed origins, methods).  
- ✅ Rate limiting must be applied to login and write-heavy endpoints.  
- ✅ Error messages must be sanitized → never expose stack traces or SQL errors.  
- ✅ APIs must follow **least privilege principle** (users can only access permitted CIs).  

---

## 🔹 Database & Storage Security  
- ✅ PostgreSQL:  
  - Enforce **role-based access** (app user has only required privileges).  
  - Enable **encryption at rest** (Postgres TDE or storage-level).  
- ✅ Neo4j:  
  - Require **TLS for driver connections**.  
  - Use separate users for read/write operations.  
- ✅ Redis:  
  - Protected by **password authentication**.  
  - No open access to the internet.  
- ✅ Backups must be **encrypted** before storage.  

---

## 🔹 Logging & Auditing  
- ✅ Use **structured logs** (zerolog for Go, JSON format).  
- ✅ Each request log must include: timestamp, user ID, correlation ID, endpoint, status.  
- ✅ All **user actions on CIs and relationships must be audit-logged**.  
- ✅ Audit logs must be immutable (append-only).  
- ✅ Retain logs according to compliance rules (banking: 7 years).  

---

## 🔹 Compliance Requirements  
- ✅ GDPR: support **user data export & deletion** requests.  
- ✅ PCI-DSS: encrypt all sensitive data in transit & at rest.  
- ✅ Maintain audit trail for **all authentication & CI modification events**.  
- ✅ Ensure **RBAC roles** map to organizational policies.  
- ✅ Provide **access reports** (who accessed which CI, when).  

---

## 🔹 Testing & Verification  
- ✅ Security tests must run in CI/CD:  
  - Static analysis (gosec, bandit).  
  - Dependency scanning (Snyk, pip-audit, npm audit).  
  - Dynamic tests (OWASP ZAP, k6 security checks).  
- ✅ Negative test cases required:  
  - Invalid JWT.  
  - Expired JWT.  
  - Unauthorized access.  
  - SQL injection attempts.  
- ✅ Regular **penetration tests** must be part of release cycles.  

---

## ✅ Summary  
- **Enforce JWT + RBAC** everywhere.  
- **Argon2id** for password hashing.  
- **No secrets in code** → use env variables / Vault.  
- **Schema validation** for every input.  
- **Audit log** for all changes.  
- **Compliance with GDPR + PCI-DSS** in data handling.  
- **Security scans & negative tests** integrated in CI/CD.  
