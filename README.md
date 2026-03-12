# 🏦 GlobalTask Bank - Sistema de Gestión de Crédito

[![Go Version](https://img.shields.io/badge/Go-1.22-blue)](https://go.dev/)
[![Vue.js 3](https://img.shields.io/badge/Vue.js-3-%234FC08D)](https://vuejs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Supabase-336791)](https://supabase.com/)

> Sistema de backend y frontend diseñado para gestionar solicitudes de crédito a escala global. Construido sobre una arquitectura limpia (**Clean Architecture**) y diseño orientado al dominio (**DDD**), el sistema soporta flujos de trabajo asíncronos distribuidos, procesamiento concurrente y actualizaciones en tiempo real mediante **WebSockets**.

---

## 🚀 ¿Por qué Go?

A pesar de no tener experiencia previa significativa con Go, elegí este lenguaje por sus **beneficios fundamentales**:

- **Concurrencia nativa**: Las **Goroutines** permiten manejar miles de solicitudes simultáneas con mínimo overhead de memoria.
- **Velocidad**: El rendimiento nativo de Go es ideal para procesar múltiples créditos en paralelo.
- **WebSockets + Worker**: La capacidad de crear un websocket y un worker en de manera sencilla gracias a las go routines

> El código fue casi en su totalidad generado con agentes de IA, específicamente usando [OpenCode](https://opencode.ai/) y el modelo **Gemini 3.0 Flash**, siguiendo la metodología de desarrollo **Spec-Driven Development (SDD)**.

---

## 📋 Metodología de Desarrollo

Este proyecto fue desarrollado utilizando **[OpenSpec](https://openspec.dev/)** (Spec-Driven Development):

1. **Exploración** → Investigación del codebase y contexto técnico
2. **Propuesta** → Análisis de impacto y plan de rollback
3. **Specs** → Especificaciones formales usando Given/When/Then
4. **Diseño** → Arquitectura técnica, data flow y contratos
5. **Tareas** → Desglose en tareas implementables
6. **Implementación** → Código siguiendo specs y diseño
7. **Verificación** → Validación contra especificaciones
8. **Archivo** → Documentación final

---

## 🏗️ Arquitectura del Sistema

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

### Componentes

| Componente | Tecnología | Descripción |
|------------|------------|-------------|
| **Frontend** | Vue 3 + Vite | Dashboard interactivo para gestión de créditos |
| **API** | Go + Gin | REST API con validación JWT y WebSockets |
| **Worker** | Go | Procesador de eventos asíncronos |
| **Mock Bank** | Go | Simulador de proveedor bancario externo |
| **K6** | Grafana k6 | Pruebas de carga y stress testing |
| **Nginx** | Nginx | Balanceador de cargas y reverse proxy |
| **Base de Datos** | PostgreSQL (Supabase) | Persistencia + Message Broker |

---

## 🗄️ Modelo de Datos

### Diagrama Entidad-Relación

```dbml
// GlobalTask Bank - Database Schema

// Tabla de países soportados
Table countries as c {
  id serial [pk]
  iso_code varchar(3) [not null, unique]
  name varchar(255) [not null]
  currency varchar(10) [not null]
}

// Perfiles de usuarios (sincronizado desde Supabase Auth)
Table profiles as p {
  id uuid [pk]
  full_name varchar(255)
  identity_document text [note: 'Almacena el documento cifrado (AES-256-GCM)']
  identity_document_bidx text [note: 'Blind index para búsquedas seguras']
  country_id int [ref: > c.id]
  role varchar(50) [not null, default: 'USER', note: 'ADMIN o USER']
  created_at timestamptz
}

// Proveedores bancarios por país
Table bank_providers as bp {
  id serial [pk]
  country_id int [not null, ref: > c.id]
  provider_name varchar(255) [not null]
  api_config jsonb [default: '{}']
  base_url varchar(255) [not null]
  created_at timestamptz
}

// Solicitudes de crédito
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

// Cola de eventos (Event Outbox Pattern)
Table event_outbox as eo {
  id uuid [pk, default: 'uuid_generate_v4()']
  event_type varchar(100) [not null]
  payload jsonb [not null, default: '{}']
  status varchar(20) [not null, default: 'PENDING', note: 'PENDING, PROCESSING, DONE, FAILED']
  created_at timestamptz
  locked_at timestamptz
}

// Mapeo de workflows a proveedores
Table workflow_providers as wp {
  id uuid [pk, default: 'uuid_generate_v4()']
  workflow_name varchar(10) [not null, note: 'Código de país: PT, CO, MX, etc.']
  provider_id int [not null, ref: > bp.id]
  event_step varchar(100) [not null, note: 'FETCH_BANK_DATA, VALIDATE_USER_IDENTITY, etc.']
  endpoint_path varchar(255) [not null]
  priority int [not null, default: 0, note: 'Menor número = mayor prioridad']
  is_active boolean [default: true]
  created_at timestamptz
  updated_at timestamptz
}

// Índices
Index idx_loan_applications_user_id on la(user_id)
Index idx_loan_applications_status on la(status)
Index idx_event_outbox_status on eo(status) where status = 'PENDING'
Index idx_workflow_providers_lookup on wp(workflow_name, event_step, is_active)
Index idx_profiles_identity_document_bidx on p(identity_document_bidx)
```

---

## 🔧 Triggers y Funcionalidades

### 1. `handle_new_user()` - Sincronización de Usuarios

```sql
-- Trigger: on_auth_user_created
-- Evento: AFTER INSERT ON auth.users
```

**Propósito**: Automatiza la creación de perfiles cuando un usuario se registra en Supabase Auth.

**Flujo**:
1. Usuario se registra en Supabase (vía frontend o Admin API)
2. Trigger `on_auth_user_created` ejecuta `handle_new_user()`
3. Se crea un registro en `public.profiles` con los metadatos del usuario
4. El rol se determina por el email:
   - Emails con `@globaltask` → **ADMIN**
   - Otros emails → **USER**

**Nota de seguridad**: El patrón de email no es seguro para producción, pero se implementó por practicidad en desarrollo.

---

### 2. `notify_loan_application_update()` - Notificaciones en Tiempo Real

```sql
-- Trigger: on_loan_application_update
-- Evento: AFTER INSERT OR UPDATE ON public.loan_applications
```

**Propósito**: Notifica a los clientes via WebSocket cuando el estado de su solicitud cambia.

**Flujo**:
1. Worker procesa un evento y actualiza el estado del crédito
2. Trigger envía notificación a través de `pg_notify('loan_application_updates', ...)`
3. API recibe la notificación y transmite a clientes conectados vía WebSocket
4. Frontend actualiza la UI en tiempo real

---

### 3. `update_updated_at_column()` - Auto-actualización de Timestamps

```sql
-- Trigger: update_loan_applications_updated_at
-- Trigger: update_workflow_providers_updated_at
-- Evento: BEFORE UPDATE ON [tabla]
```

**Propósito**: Mantiene automáticamente la columna `updated_at` con la fecha/hora actual en cada actualización.

---

### 4. Políticas de Seguridad (RLS - Row Level Security)

| Tabla | Política | Condición |
|-------|----------|-----------|
| `loan_applications` | ADMIN: gestionar todos | `is_admin() = true` |
| `loan_applications` | USER: ver propios | `user_id = auth.uid()` |
| `profiles` | ADMIN: gestionar todos | `is_admin() = true` |
| `profiles` | USER: ver propio | `id = auth.uid()` |
| `countries` | Authenticated: listar | `true` |
| `bank_providers` | ADMIN: gestionar todos | `is_admin() = true` |
| `workflow_providers` | ADMIN: gestionar todos | `is_admin() = true` |
| `event_outbox` | ADMIN: gestionar todos | `is_admin() = true` |

---

## 📦 Instalación y Ejecución

### Requisitos Previos

- **Docker** y **Docker Compose** instalados
- **Git** para clonar el repositorio
- **Supabase CLI** instalado globalmente - si no lo tienes, `make init` lo instala automáticamente
- **npm** para instalar supabase cli

### Pasos de Instalación

```bash
# 1. Clonar el repositorio
git clone <repo-url>
cd globaltask-bank

# 2. Inicializar Supabase (instala CLI si no existe y genera keys)
make init

# 3. Copiar y configurar variables de entorno
cp .env.example .env
# ⚠️ Edita el archivo .env con las keys generadas en el paso anterior

# 4. Levantar el proyecto
make up
```

### Servicios Disponibles

| Servicio | URL | Descripción |
|----------|-----|-------------|
| **Frontend** | [http://localhost](http://localhost) | Dashboard de gestión |
| **API** | [http://localhost/health](http://localhost/health) | Health check |
| **WebSocket** | `ws://localhost/api/v1/ws` | Endpoint de WebSockets |
| **Mock Bank** | [http://localhost:8081](http://localhost:8081) | Simulador bancario |

---

## ⚙️ Variables de Entorno

```bash
# ==========================================
# Supabase Configuration (LOCAL)
# ==========================================

# URL de Supabase para comunicación interna entre contenedores
SUPABASE_URL=http://host.docker.internal:54321

# Claves de Supabase (generadas por make init)
SUPABASE_SERVICE_ROLE_KEY=eyJhbGciOiJFUzI1NiIs...
DEFAULT_USER_PASSWORD=password123

# Configuración JWT
SUPABASE_JWKS_URL=http://host.docker.internal:54321/auth/v1/.well-known/jwks.json

# Conexión a la base de datos
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

> **⚠️ NOTA DE SEGURIDAD**: La variable `DEFAULT_USER_PASSWORD` es una práctica **no recomendada** para producción. En un entorno real, debería implementarse un flujo de invitación o establecimiento de contraseña por el usuario.

---

## 🧠 Decisiones Técnicas

### 1. Supabase para Autenticación y Base de Datos

Se eligió **Supabase** por:
- **API de autenticación** sencilla y segura con JWT de firma asimétrica
- **PostgreSQL gestionado** con Row Level Security (RLS) incluido
- El JWT se valida en el middleware de Go, extrayendo `user_id` y `role` para la sesión

### 2. Nginx como Balanceador de Cargas

Originalmente se planificó una arquitectura de microservicios con múltiples réplicas, pero se optó por una arquitectura **Event-Driven**. Nginx quedó como reverse proxy y balanceador de cargas para futuras escalabilidades.

### 3. Arquitectura Orientada por Eventos (Event-Driven)

Para evitar que la solicitud de un crédito y las comprobaciones con proveedores sea un proceso bloqueante:

1. La API recibe la solicitud y la guarda en la base de datos
2. Crea un registro en `event_outbox` (dentro de la misma transacción - **Atomicidad UoW**)
3. El **Worker** procesa eventos de forma asíncrona:
   - Usa `SELECT FOR UPDATE SKIP LOCKED` para evitar procesamiento duplicado
   - Patrón **Unit of Work** para garantizar transacciones atómicas
4. Si el evento requiere consultar un proveedor, usa `workflow_providers` para determinar cuál llamar
5. Si falla, puede reintentar con el siguiente proveedor según la prioridad configurada

### 4. Tabla `workflow_providers`

Mapea eventos a proveedores específicos por país:
- `workflow_name`: Código del país (PT, CO, MX)
- `event_step`: Tipo de evento (FETCH_BANK_DATA, VALIDATE_USER_IDENTITY)
- `priority`: Si falla un proveedor, se intenta el siguiente

### 5. PII Safe Queries (Cifrado de Documentos)

Para cumplir con regulaciones de privacidad:
- `identity_document` se almacena **cifrado** (Base64 por practicidad)
- `identity_document_bidx` es un **blind index** que permite búsquedas exactas sin exponer el documento real

---

## 🔒 Consideraciones de Seguridad

### Implementado ✅
- **Validación JWT**: El backend valida criptográficamente cada token
- **RLS (Row Level Security)**: A nivel de base de datos
- **Principio de Mínimo Privilegio**: Backend usa usuario con permisos restringidos

### Pendiente ⚠️
- **Throttling**: Limitar requests por IP/usuario a nivel de aplicación
- **Conexiones optimizadas para RLS**: Mejorar pooling de conexiones para que RLS funcione correctamente en todos los escenarios
- **Condiguración de kubernetes**, debido a que no tengo experiencia en el area

---

## 📈 Escalabilidad

- **Workers horizontales**: Pueden replicarse infinitamente gracias a `FOR UPDATE SKIP LOCKED`
- **Particionamiento**: La tabla `loan_applications` puede particionarse por país o fecha
- **Goroutines**: El modelo CSP de Go permite miles de tareas ligeras

---

## 🛠️ Comandos Útiles

```bash
make init        # Inicializar Supabase y generar keys
make up          # Levantar servicios
make down        # Detener servicios
make logs        # Ver logs en tiempo real
make test        # Ejecutar tests de la API
make create-user EMAIL=user@example.com PASS=password123  # Crear usuario admin
```

---

## 📁 Estructura del Proyecto

```
globaltask-bank/
├── backend/              # API y Worker (Go)
│   ├── cmd/              # Puntos de entrada
│   ├── internal/         # Código domain
│   └── ...
├── frontend/             # Dashboard (Vue 3)
│   ├── src/
│   └── ...
├── infra/
│   ├── supabase/         # Migraciones y configuración
│   │   └── migrations/   # Scripts SQL
│   └── nginx/           # Configuración de Nginx
├── k6/                   # Pruebas de carga (K6)
│   └── scale-test.js    # Test de escala
├── docker-compose.yml   # Orquestación de servicios
├── Makefile            # Comandos de desarrollo
└── README.md           # Este archivo
```

---

## 🧪 Pruebas de Carga con K6

El proyecto incluye **K6** (Grafana k6) para ejecutar pruebas de carga y stress testing.

### ¿Qué hace el test?

El archivo `k6/scale-test.js` simula un día completo de tráfico:

- **Cálculos base:**
  - 1M usuarios/día ≈ 11.5 RPS promedio
  - Hora pico (3x) ≈ 35 RPS
  - Pico absoluto (10x) ≈ 115 RPS

- **Comportamiento simulado:**
  - 70% reads (listar préstamos)
  - 25% writes (crear préstamos)
  - 5% detail reads (ver detalle)

### Configuración de Stages

| Etapa | Duración | Usuarios (VUs) |
|-------|----------|----------------|
| Noche (mínimo) | 30s | 2% del pico |
| Amanecer | 30s | 10% del pico |
| Mañana | 1m | 30% del pico |
| Mediodía | 1m | 60% del pico |
| **Pico máximo** | 3m (configurable) | 100% del pico |
| Atardecer | 1m | 40% del pico |
| Noche | 1m | 10% del pico |

### Umbrales de Rendimiento

| Métrica | Umbral | Descripción |
|---------|--------|-------------|
| Login failures | < 5% | 95% de éxito |
| Loan failures | < 10% | 90% de éxito |
| Read failures | < 2% | 98% de éxito |
| Login time p(95) | < 3s | Percentil 95 |
| Loan time p(95) | < 5s | Percentil 95 |
| Read time p(95) | < 1s | Percentil 95 |

### Ejecución

```bash
# Ejecutar test con configuración por defecto (200 VUs, 1M usuarios/día)
docker compose run --rm k6 run /scripts/scale-test.js
O
make test-scale

# Simular 1M usuarios/día con pico de 500 VUs
docker compose run --rm k6 run /scripts/scale-test.js -e MAX_VUS=500

# Simular escenario extremo (2M usuarios/día)
docker compose run --rm k6 run /scripts/scale-test.js -e MAX_VUS=1000 -e DAILY_USERS=2000000

# Escenario de pico sostenido por 10 minutos
docker compose run --rm k6 run /scripts/scale-test.js -e SUSTAIN_MINUTES=10
```

### Variables de Entorno

```bash
ADMIN_EMAIL=admin@globaltask.com      # Email del usuario admin
ADMIN_PASSWORD=password123            # Contraseña
SUPABASE_URL=http://host.docker.internal:54321
API_URL=http://host.docker.internal
MAX_VUS=200                           # Máx usuarios virtuales
DAILY_USERS=1000000                   # Usuarios diarios objetivo
SUSTAIN_MINUTES=3                     # Minutos en pico
COUNTRY_ID=1                          # País para los tests
```

> **Nota**: K6 está configurado en docker-compose.yml pero comentado por defecto. Para activarlo, descomenta la línea `command` en el servicio.

---

## 🧪 Stack Tecnológico

| Capa | Tecnología |
|------|------------|
| **Backend** | Go 1.22, Gin, PostgreSQL (pgx/v5), JWT, WebSockets |
| **Frontend** | Vue 3, Vite 5, Pinia, Vue Router, Axios, Supabase JS, Tailwind CSS |
| **Testing** | K6 (Grafana) - Pruebas de carga |
| **Infra** | Docker, Docker Compose, Nginx, Supabase |

---

**Desarrollado con ❤️**
*Powered by AI + OpenSpec Methodology*
