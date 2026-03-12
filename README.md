# 🏦 GlobalTask Bank - Credit Management System

[![Go Version](https://img.shields.io/badge/Go-1.22-blue)](https://go.dev/)
[![Vue.js 3](https://img.shields.io/badge/Vue.js-3-%234FC08D)](https://vuejs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Supabase-336791)](https://supabase.com/)

> Backend and frontend system designed to manage credit applications at global scale. Built on **Clean Architecture** and **Domain-Driven Design (DDD)**, the system supports distributed asynchronous workflows, concurrent processing, and real-time updates via **WebSockets**.

---

## 🚀 Why Go?

Despite having no significant prior experience with Go, I chose this language for its **fundamental benefits**:

- **Native concurrency**: **Goroutines** allow handling thousands of simultaneous requests with minimal memory overhead.
- **Speed**: Go's native performance is ideal for processing multiple credit applications in parallel.
- **WebSockets + Worker**: The ability to create a websocket and worker easily thanks to goroutines

> The code was almost entirely generated with AI agents, specifically using [OpenCode](https://opencode.ai/) and the **Gemini 3.0 Flash** model, following the **Spec-Driven Development (SDD)** methodology.

---

## 📋 Development Methodology

This project was developed using **[OpenSpec](https://openspec.dev/)** (Spec-Driven Development):

1. **Exploration** → Codebase investigation and technical context
2. **Proposal** → Impact analysis and rollback plan
3. **Specs** → Formal specifications using Given/When/Then
4. **Design** → Technical architecture, data flow and contracts
5. **Tasks** → Breakdown into actionable tasks
6. **Implementation** → Code following specs and design
7. **Verification** → Validation against specifications
8. **Archive** → Final documentation

---

## 🏗️ System Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              NGINX (Reverse Proxy & Load Balancer)          │
│                                    Puerto 80                                 │
└────────────────────────────────┬────────────────────────────────────────────┘
                                 │
              ┌──────────────────┼──────────────────┐
              │                  │                  │
              ▼                  ▼                  ▼
       ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
       │  Frontend  │    │     API     │    │  Mock Bank  │
       │  (Vue 3)   │    │    (Go)     │    │   (Go)      │
       │  Puerto 80 │    │  Puerto 8080│    │  Puerto 8081│
       └─────────────┘    └──────┬──────┘    └─────────────┘
                                 │
                                 ▼
                    ┌────────────────────────┐
                    │   PostgreSQL           │
                    │   (Supabase)          │
                    │   + Event Outbox      │
                    └───────────┬────────────┘
                                │
                                ▼
                    ┌────────────────────────┐
                    │      Worker            │
                    │   (Go - Background)   │
                    └────────────────────────┘
```

### Components

| Component | Technology | Description |
|------------|------------|-------------|
| **Frontend** | Vue 3 + Vite | Interactive dashboard for credit management |
| **API** | Go + Gin | REST API with JWT validation and WebSockets |
| **Worker** | Go | Asynchronous event processor |
| **Mock Bank** | Go | External bank provider simulator |
| **K6** | Grafana k6 | Load testing and stress testing |
| **Nginx** | Nginx | Load balancer and reverse proxy |
| **Database** | PostgreSQL (Supabase) | Persistence + Message Broker |

---

## 🗄️ Data Model

### Entity-Relationship Diagram

```dbml
// GlobalTask Bank - Database Schema

// Supported countries table
Table countries as c {
  id serial [pk]
  iso_code varchar(3) [not null, unique]
  name varchar(255) [not null]
  currency varchar(10) [not null]
}

// User profiles (synchronized from Supabase Auth)
Table profiles as p {
  id uuid [pk]
  full_name varchar(255)
  identity_document text [note: 'Stores encrypted document (AES-256-GCM)']
  identity_document_bidx text [note: 'Blind index for secure searches']
  country_id int [ref: > c.id]
  role varchar(50) [not null, default: 'USER', note: 'ADMIN or USER']
  created_at timestamptz
}

// Bank providers by country
Table bank_providers as bp {
  id serial [pk]
  country_id int [not null, ref: > c.id]
  provider_name varchar(255) [not null]
  api_config jsonb [default: '{}']
  base_url varchar(255) [not null]
  created_at timestamptz
}

// Credit applications
Table loan_applications as la {
  id uuid [pk, default: 'uuid_generate_v4()']
  user_id uuid [not null, ref: > p.id]
  requested_amount numeric(15, 2) [not null]
  monthly_income numeric(15, 2) [not null]
  status varchar(50) [not null, default: 'DRAFT', note: 'DRAFT, PENDING_VALIDATION, AWAITING_BANK_DATA, ANALYZING_RISK, APPROVED, REJECTED']
  bank_information jsonb [default: '{}']
  requested_at timestamptz [not null]
  created_at timestamptz
  updated_at timestamptz
}

// Event queue (Event Outbox Pattern)
Table event_outbox as eo {
  id uuid [pk, default: 'uuid_generate_v4()']
  event_type varchar(100) [not null]
  payload jsonb [not null, default: '{}']
  status varchar(20) [not null, default: 'PENDING', note: 'PENDING, PROCESSING, DONE, FAILED']
  created_at timestamptz
  locked_at timestamptz
}

// Workflow to provider mapping
Table workflow_providers as wp {
  id uuid [pk, default: 'uuid_generate_v4()']
  workflow_name varchar(10) [not null, note: 'Country code: PT, CO, MX, etc.']
  provider_id int [not null, ref: > bp.id]
  event_step varchar(100) [not null, note: 'FETCH_BANK_DATA, VALIDATE_USER_IDENTITY, etc.']
  endpoint_path varchar(255) [not null]
  priority int [not null, default: 0, note: 'Lower number = higher priority']
  is_active boolean [default: true]
  created_at timestamptz
  updated_at timestamptz
}

// Indexes
Index idx_loan_applications_user_id on la(user_id)
Index idx_loan_applications_status on la(status)
Index idx_event_outbox_status on eo(status) where status = 'PENDING'
Index idx_workflow_providers_lookup on wp(workflow_name, event_step, is_active)
Index idx_profiles_identity_document_bidx on p(identity_document_bidx)
```

---

## 🔧 Triggers and Functionality

### 1. `handle_new_user()` - User Synchronization

```sql
-- Trigger: on_auth_user_created
-- Event: AFTER INSERT ON auth.users
```

**Purpose**: Automates the creation of profiles when a user registers in Supabase Auth.

**Flow**:
1. User registers in Supabase (via frontend or Admin API)
2. Trigger `on_auth_user_created` executes `handle_new_user()`
3. A record is created in `public.profiles` with the user's metadata
4. The role is determined by the email:
   - Emails with `@globaltask` → **ADMIN**
   - Other emails → **USER**

**Security note**: The email pattern is not secure for production, but was implemented for development convenience.

---

### 2. `notify_loan_application_update()` - Real-Time Notifications

```sql
-- Trigger: on_loan_application_update
-- Event: AFTER INSERT OR UPDATE ON public.loan_applications
```

**Purpose**: Notifies clients via WebSocket when their application status changes.

**Flow**:
1. Worker processes an event and updates the credit status
2. Trigger sends notification via `pg_notify('loan_application_updates', ...)`
3. API receives the notification and transmits to connected clients via WebSocket
4. Frontend updates the UI in real-time

---

### 3. `update_updated_at_column()` - Auto-update Timestamps

```sql
-- Trigger: update_loan_applications_updated_at
-- Trigger: update_workflow_providers_updated_at
-- Event: BEFORE UPDATE ON [table]
```

**Purpose**: Automatically maintains the `updated_at` column with the current date/time on each update.

---

### 4. Security Policies (RLS - Row Level Security)

| Table | Policy | Condition |
|-------|----------|-----------|
| `loan_applications` | ADMIN: manage all | `is_admin() = true` |
| `loan_applications` | USER: view own | `user_id = auth.uid()` |
| `profiles` | ADMIN: manage all | `is_admin() = true` |
| `profiles` | USER: view own | `id = auth.uid()` |
| `countries` | Authenticated: list | `true` |
| `bank_providers` | ADMIN: manage all | `is_admin() = true` |
| `workflow_providers` | ADMIN: manage all | `is_admin() = true` |
| `event_outbox` | ADMIN: manage all | `is_admin() = true` |

---

## 📦 Installation and Execution

### Prerequisites

- **Docker** and **Docker Compose** installed
- **Git** to clone the repository
- **Supabase CLI** installed globally - if you don't have it, `make init` installs it automatically
- **npm** to install supabase CLI

### Installation Steps

```bash
# 1. Clone the repository
git clone <repo-url>
cd globaltask-bank

# 2. Initialize Supabase (installs CLI if not exists and generates keys)
make init

# 3. Copy and configure environment variables
cp .env.example .env
# ⚠️ Edit the .env file with the keys generated in the previous step

# 4. Start the project
make up
```

### Available Services

| Service | URL | Description |
|----------|-----|-------------|
| **Frontend** | [http://localhost](http://localhost) | Management dashboard |
| **API** | [http://localhost/health](http://localhost/health) | Health check |
| **WebSocket** | `ws://localhost/api/v1/ws` | WebSockets endpoint |
| **Mock Bank** | [http://localhost:8081](http://localhost:8081) | Bank simulator |

---

## ⚙️ Environment Variables

```bash
# ==========================================
# Supabase Configuration (LOCAL)
# ==========================================

# Supabase URL for internal container communication
SUPABASE_URL=http://host.docker.internal:54321

# Supabase keys (generated by make init)
SUPABASE_SERVICE_ROLE_KEY=eyJhbGciOiJFUzI1NiIs...
DEFAULT_USER_PASSWORD=password123

# JWT configuration
SUPABASE_JWKS_URL=http://host.docker.internal:54321/auth/v1/.well-known/jwks.json

# Database connection
DATABASE_URL=postgresql://postgres:postgres@db:5432/postgres

# ==========================================
# Application Configuration
# ==========================================
API_PORT=8080
LOG_LEVEL=debug
WORKER_CONCURRENCY=5

# ==========================================
# External Services
# ==========================================
MOCK_BANK_URL=http://mock-bank:8081/validate
BACKEND_WEBHOOK_URL=http://api:8080/webhook/bank-update

# ==========================================
# Frontend Configuration
# ==========================================
VITE_SUPABASE_URL=http://localhost/supabase
VITE_SUPABASE_ANON_KEY=SUPABASE_ANON_KEY
```

> **⚠️ SECURITY NOTE**: The `DEFAULT_USER_PASSWORD` variable is **not recommended** for production. In a real environment, an invitation flow or user-defined password setup should be implemented.

---

## 🧠 Technical Decisions

### 1. Supabase for Authentication and Database

**Supabase** was chosen for:
- Simple and secure **authentication API** with asymmetric JWT signing
- **Managed PostgreSQL** with Row Level Security (RLS) included
- JWT is validated in Go middleware, extracting `user_id` and `role` for the session

### 2. Nginx as Load Balancer

Originally a microservices architecture with multiple replicas was planned, but an **Event-Driven** architecture was chosen instead. Nginx remained as reverse proxy and load balancer for future scalability.

### 3. Event-Driven Architecture

To avoid credit application requests and provider checks being a blocking process:

1. API receives the request and saves it to the database
2. Creates a record in `event_outbox` (within the same transaction - **UoW Atomicity**)
3. The **Worker** processes events asynchronously:
   - Uses `SELECT FOR UPDATE SKIP LOCKED` to avoid duplicate processing
   - **Unit of Work** pattern to guarantee atomic transactions
4. If the event requires querying a provider, uses `workflow_providers` to determine which one to call
5. If it fails, can retry with the next provider according to configured priority

### 4. `workflow_providers` Table

Maps events to specific providers by country:
- `workflow_name`: Country code (PT, CO, MX)
- `event_step`: Event type (FETCH_BANK_DATA, VALIDATE_USER_IDENTITY)
- `priority`: If a provider fails, the next one is tried

### 5. PII Safe Queries (Document Encryption)

To comply with privacy regulations:
- `identity_document` is stored **encrypted** (Base64 for convenience)
- `identity_document_bidx` is a **blind index** that allows exact searches without exposing the real document

---

## 🔒 Security Considerations

### Implemented ✅
- **JWT Validation**: Backend cryptographically validates each token
- **RLS (Row Level Security)**: At database level
- **Principle of Least Privilege**: Backend uses user with restricted permissions

### Pending ⚠️
- **Throttling**: Limit requests per IP/user at application level
- **Optimized connections for RLS**: Improve connection pooling so RLS works correctly in all scenarios
- **Kubernetes configuration**, due to lack of experience in the area

---

## 📈 Scalability

- **Horizontal workers**: Can be replicated infinitely thanks to `FOR UPDATE SKIP LOCKED`
- **Partitioning**: The `loan_applications` table can be partitioned by country or date
- **Goroutines**: Go's CSP model allows thousands of lightweight tasks

---

## 🛠️ Useful Commands

```bash
make init        # Initialize Supabase and generate keys
make up          # Start services
make down        # Stop services
make logs        # View logs in real-time
make test        # Run API tests
make create-user EMAIL=user@example.com PASS=password123  # Create admin user
```

---

## 📁 Project Structure

```
globaltask-bank/
├── backend/              # API and Worker (Go)
│   ├── cmd/              # Entry points
│   ├── internal/         # Domain code
│   └── ...
├── frontend/             # Dashboard (Vue 3)
│   ├── src/
│   └── ...
├── infra/
│   ├── supabase/         # Migrations and configuration
│   │   └── migrations/   # SQL scripts
│   └── nginx/           # Nginx configuration
├── k6/                   # Load testing (K6)
│   └── scale-test.js    # Scale test
├── docker-compose.yml   # Service orchestration
├── Makefile            # Development commands
└── README.md           # This file
```

---

## 🧪 Load Testing with K6

The project includes **K6** (Grafana k6) for load testing and stress testing.

### What does the test do?

The `k6/scale-test.js` file simulates a full day of traffic:

- **Base calculations:**
  - 1M users/day ≈ 11.5 RPS average
  - Peak hour (3x) ≈ 35 RPS
  - Absolute peak (10x) ≈ 115 RPS

- **Simulated behavior:**
  - 70% reads (list loans)
  - 25% writes (create loans)
  - 5% detail reads (view detail)

### Stage Configuration

| Stage | Duration | Users (VUs) |
|-------|----------|----------------|
| Night (minimum) | 30s | 2% of peak |
| Dawn | 30s | 10% of peak |
| Morning | 1m | 30% of peak |
| Noon | 1m | 60% of peak |
| **Maximum peak** | 3m (configurable) | 100% of peak |
| Evening | 1m | 40% of peak |
| Night | 1m | 10% of peak |

### Performance Thresholds

| Metric | Threshold | Description |
|---------|--------|-------------|
| Login failures | < 5% | 95% success |
| Loan failures | < 10% | 90% success |
| Read failures | < 2% | 98% success |
| Login time p(95) | < 3s | 95th percentile |
| Loan time p(95) | < 5s | 95th percentile |
| Read time p(95) | < 1s | 95th percentile |

### Execution

```bash
# Run test with default configuration (200 VUs, 1M users/day)
docker compose run --rm k6 run /scripts/scale-test.js
OR
make test-scale

# Simulate 1M users/day with 500 VUs peak
docker compose run --rm k6 run /scripts/scale-test.js -e MAX_VUS=500

# Simulate extreme scenario (2M users/day)
docker compose run --rm k6 run /scripts/scale-test.js -e MAX_VUS=1000 -e DAILY_USERS=2000000

# Sustained peak scenario for 10 minutes
docker compose run --rm k6 run /scripts/scale-test.js -e SUSTAIN_MINUTES=10
```

### Environment Variables

```bash
ADMIN_EMAIL=admin@globaltask.com      # Admin user email
ADMIN_PASSWORD=password123            # Password
SUPABASE_URL=http://host.docker.internal:54321
API_URL=http://host.docker.internal
MAX_VUS=200                           # Max virtual users
DAILY_USERS=1000000                   # Target daily users
SUSTAIN_MINUTES=3                     # Minutes at peak
COUNTRY_ID=1                          # Country for tests
```

> **Note**: K6 is configured in docker-compose.yml but commented by default. To activate it, uncomment the `command` line in the service.

---

## 🧪 Technology Stack

| Layer | Technology |
|------|------------|
| **Backend** | Go 1.22, Gin, PostgreSQL (pgx/v5), JWT, WebSockets |
| **Frontend** | Vue 3, Vite 5, Pinia, Vue Router, Axios, Supabase JS, Tailwind CSS |
| **Testing** | K6 (Grafana) - Load testing |
| **Infra** | Docker, Docker Compose, Nginx, Supabase |

---

**Developed with ❤️**
*Powered by AI + OpenSpec Methodology*
