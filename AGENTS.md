# GlobalTask Bank - Agent Skills & Stack

## 🏗️ Stack Tecnológico

### Backend
- **Go 1.22** - Lenguaje principal
- **Gin** - Framework HTTP
- **PostgreSQL (pgx/v5)** - Base de datos
- **JWT (golang-jwt/v5)** - Autenticación
- **WebSockets (gorilla)** - Real-time updates
- **YAML** - Configuración

### Frontend
- **Vue 3** - Framework UI
- **Vite 5** - Build tool
- **Pinia** - State management
- **Vue Router 4** - Routing
- **Axios** - HTTP client
- **Supabase JS** - Auth & DB client
- **Tailwind CSS** - Styling
- **Lucide Icons** - Iconografía

### Infraestructura
- **Docker & Docker Compose** - Containerización
- **Nginx** - Reverse proxy & rate limiting
- **Supabase** - Auth & PostgreSQL hosting

### Arquitectura
- **Clean Architecture** - Separación de capas
- **Domain-Driven Design (DDD)** - Modelado del dominio
- **Event-Driven** - Webhooks & Event Outbox
- **Servicios** - API, Worker, MockBank

---

## 🤖 Skills del Proyecto (`.agents/skills/`)

| Skill | Para qué sirve |
|-------|----------------|
| **golang-pro** | Patrones concurrentes (goroutines, channels), microservicios, testing, error handling idiomático en Go |
| **api-design** | Diseño de APIs REST y GraphQL, OpenAPI, versionado, paginación, autenticación |
| **supabase-postgres-best-practices** | Optimización de schemas PostgreSQL, Row Level Security (RLS), queries, integración con Supabase Auth |
| **web-design-guidelines** | Diseño de interfaces, UX/UI, accesibilidad, responsive design |
| **bank-provider-scaffold** ⭐ | Crear integraciones con nuevos proveedores bancarios usando el patrón Strategy (`ProviderClient`) |

**⭐ = Skills personalizadas del proyecto**

---

## 🌍 Skills Globales (`~/.config/opencode/skills/`)

Estas skills están disponibles en **todos** tus proyectos:

| Skill | Para qué sirve |
|-------|----------------|
| **sdd-init** | Inicializar Spec-Driven Development: detecta stack, crea estructura `openspec/` o usa Engram |
| **sdd-explore** | Explorar el codebase y documentar contexto técnico antes de cambios |
| **sdd-propose** | Crear propuestas de cambio con análisis de impacto y rollback plan |
| **sdd-spec** | Escribir especificaciones formales usando Given/When/Then y RFC 2119 |
| **sdd-design** | Diseñar arquitectura técnica: decisiones, data flow, contratos, testing strategy |
| **sdd-tasks** | Desglosar el diseño en tareas implementables y priorizadas |
| **sdd-apply** | Implementar tareas siguiendo specs y design (soporta TDD si está configurado) |
| **sdd-verify** | Verificar que la implementación cumple todas las especificaciones |
| **sdd-archive** | Archivar cambios completados y actualizar la documentación del proyecto |

---

## 📝 Cómo Usar las Skills

### Skills del Proyecto
Se activan automáticamente según el contexto:
- Trabajando con Go → `golang-pro`
- Diseñando endpoints → `api-design`
- Queries/schemas → `supabase-postgres-best-practices`
- "Agrega proveedor bancario BBVA" → `bank-provider-scaffold`

### Skills Globales (SDD)
Se invocan explícitamente o por el orquestador SDD:

**Flujo típico:**
1. `sdd-init` → Inicializa el contexto SDD
2. `sdd-explore` → Investiga el área a cambiar
3. `sdd-propose` → Crea la propuesta de cambio
4. `sdd-spec` + `sdd-design` → Especifica comportamiento y diseño técnico (pueden correr en paralelo)
5. `sdd-tasks` → Desglosa en tareas
6. `sdd-apply` → Implementa las tareas
7. `sdd-verify` → Valida contra specs
8. `sdd-archive` → Archiva el cambio

**Modos de persistencia:**
- `engram`: Usa el sistema de memoria distribuida
---

## 🎯 Ejemplo Práctico

**User:** "Necesito agregar soporte para Chile en el sistema"

**Agent:**
1. Detecta → Usa `add-country-provider`
2. Lee → `golang-pro`, `api-design`
3. Pregunta → País, moneda, proveedores, reglas de riesgo
4. Implementa → SQL migrations, YAML config, Go code, Vue UI

---

**Última actualización:** 12 Mar 2026
