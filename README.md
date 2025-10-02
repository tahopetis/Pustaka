# Pustaka CMDB

> A modern, open-source **Configuration Management Database (CMDB)** designed for managing IT assets, their relationships, and governance in enterprise environments.

---

## 🚀 Overview

**Pustaka** provides a structured and extensible way to store, manage, and visualize configuration items (CIs) across IT domains.  
It supports hierarchical taxonomies, relationship mapping, access control, and auditing — built to align with enterprise architecture and IT service management practices.

**Phase 1 (✅ Complete):**

- Core CMDB data model (Domains → Categories → Subcategories → CIs)  
- Relationship management between CIs  
- Role-Based Access Control (RBAC)  
- Audit logging for traceability  
- Search and filtering of assets  
- Graph visualization of relationships  
- Docker-based deployment  

---

## 🏗 Architecture

### Backend (Go)
- HTTP framework: Chi v5  
- PostgreSQL for structured CMDB data  
- Neo4j for storing and querying CI relationships  
- Redis for caching and session handling  
- JWT-based authentication & RBAC  

### Frontend (Vue 3 + TypeScript)
- Modern Vue.js UI with Pinia for state management  
- vis-network for graph visualization of CI relationships  
- Responsive layout for asset browsing and management  

---

## 📦 Features (Phase 1)

- **CMDB Taxonomy:** Domain → Category → Subcategory → Configuration Item  
- **Relationship Mapping:** Define and query dependencies across assets  
- **RBAC:** Admin, User, and Viewer roles with scoped permissions  
- **Audit Logging:** Record every change for compliance and governance  
- **Graph Explorer:** Interactive visualization of CI relationships  
- **Search & Filters:** Find CIs by type, attributes, or relations  

---

## 🔧 Prerequisites

- Docker & Docker Compose  
- Go 1.22+ (if running locally)  
- Node.js 18+ (frontend development)  

---

## 🎯 Getting Started

### Quick Start (Docker Compose)

```sh
git clone https://github.com/tahopetis/Pustaka.git
cd Pustaka
cp .env.example .env   # configure DB, JWT secret, etc.
make dev
````

* Frontend: [http://localhost:3000](http://localhost:3000)
* API: [http://localhost:8080](http://localhost:8080)
* Neo4j Browser: [http://localhost:7474](http://localhost:7474)

Create an initial admin user:

```sh
make create-admin
```

---

## 🛠 Project Structure

```
pustaka/
├── cmd/                 # Entrypoints (API, migrations, setup)
├── internal/            # Backend application logic
│   ├── cmdb/            # CMDB models & services
│   ├── auth/            # Auth & RBAC
│   ├── audit/           # Audit logging
│   └── graph/           # Neo4j integration
├── web/                 # Vue.js frontend
├── docs/                # FSD, TSD, implementation plan, etc.
├── docker/              # Deployment configs
├── Makefile
└── README.md
```

---

## 📋 Documentation

* [FSD.md](./docs/FSD.md) — Functional Specification
* [TSD.md](./docs/TSD.md) — Technical Specification
* [IMPLEMENTATION_PLAN.md](./docs/IMPLEMENTATION_PLAN.md) — Development phases & tasks

---

## 🧭 Roadmap

* **Phase 1 (Complete):** Core CMDB, RBAC, Audit logging, Graph visualization
* **Phase 2 (Planned):** API integrations, advanced search, reporting dashboards
* **Phase 3 (Planned):** Workflow automation, external system sync, AI-assisted insights

---

## 🙌 Contributing

1. Fork the repo and create a branch
2. Implement changes and add tests
3. Run `make test`
4. Submit a Pull Request

---

## ✅ License

Licensed under the **MIT License**. See [LICENSE](./LICENSE).

---

## 📚 Related

* `FSD.md` (Functional Specification)
* `TSD.md` (Technical Specification)
* `IMPLEMENTATION_PLAN.md`

---

Pustaka aims to provide enterprises with a robust and extensible **CMDB** foundation for IT governance, compliance, and architecture alignment.


