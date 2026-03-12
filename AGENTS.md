# GlobalTask Bank - Agent Skills & Stack

## 🏗️ Technology Stack

### Backend
- **Go 1.22** - Primary language
- **Gin** - HTTP framework
- **PostgreSQL (pgx/v5)** - Database
- **JWT (golang-jwt/v5)** - Authentication
- **WebSockets (gorilla)** - Real-time updates
- **YAML** - Configuration

### Frontend
- **Vue 3** - UI framework
- **Vite 5** - Build tool
- **Pinia** - State management
- **Vue Router 4** - Routing
- **Axios** - HTTP client
- **Supabase JS** - Auth & DB client
- **Tailwind CSS** - Styling
- **Lucide Icons** - Iconography

### Infrastructure
- **Docker & Docker Compose** - Containerization
- **Nginx** - Reverse proxy & rate limiting
- **Supabase** - Auth & PostgreSQL hosting

### Architecture
- **Clean Architecture** - Layer separation
- **Domain-Driven Design (DDD)** - Domain modeling
- **Event-Driven** - Webhooks & Event Outbox
- **Services** - API, Worker, MockBank

---

## 🤖 Project Skills (`.agents/skills/`)

| Skill | Purpose |
|-------|----------------|
| **golang-pro** | Concurrent patterns (goroutines, channels), microservices, testing, idiomatic Go error handling |
| **api-design** | REST and GraphQL API design, OpenAPI, versioning, pagination, authentication |
| **supabase-postgres-best-practices** | PostgreSQL schema optimization, Row Level Security (RLS), queries, Supabase Auth integration |
| **web-design-guidelines** | Interface design, UX/UI, accessibility, responsive design |
| **bank-provider-scaffold** ⭐ | Create integrations with new bank providers using the Strategy pattern (`ProviderClient`) |

**⭐ = Custom project skills**

---

## 🌍 Global Skills (`~/.config/opencode/skills/`)

These skills are available in **all** your projects:

| Skill | Purpose |
|-------|----------------|
| **sdd-init** | Initialize Spec-Driven Development: detects stack, creates `openspec/` structure or uses Engram |
| **sdd-explore** | Explore codebase and document technical context before changes |
| **sdd-propose** | Create change proposals with impact analysis and rollback plan |
| **sdd-spec** | Write formal specifications using Given/When/Then and RFC 2119 |
| **sdd-design** | Design technical architecture: decisions, data flow, contracts, testing strategy |
| **sdd-tasks** | Break down design into actionable and prioritized tasks |
| **sdd-apply** | Implement tasks following specs and design (supports TDD if configured) |
| **sdd-verify** | Verify that implementation meets all specifications |
| **sdd-archive** | Archive completed changes and update project documentation |

---

## 📝 How to Use Skills

### Project Skills
Activated automatically based on context:
- Working with Go → `golang-pro`
- Designing endpoints → `api-design`
- Queries/schemas → `supabase-postgres-best-practices`
- "Add BBVA bank provider" → `bank-provider-scaffold`

### Global Skills (SDD)
Invoked explicitly or by the SDD orchestrator:

**Typical flow:**
1. `sdd-init` → Initialize SDD context
2. `sdd-explore` → Investigate area to change
3. `sdd-propose` → Create change proposal
4. `sdd-spec` + `sdd-design` → Specify behavior and technical design (can run in parallel)
5. `sdd-tasks` → Break down into tasks
6. `sdd-apply` → Implement tasks
7. `sdd-verify` → Validate against specs
8. `sdd-archive` → Archive the change

**Persistence modes:**
- `engram`: Uses distributed memory system
---

## 🎯 Practical Example

**User:** "I need to add support for Chile in the system"

**Agent:**
1. Detects → Uses `add-country-provider`
2. Reads → `golang-pro`, `api-design`
3. Asks → Country, currency, providers, risk rules
4. Implements → SQL migrations, YAML config, Go code, Vue UI

---

**Last updated:** Mar 12, 2026
