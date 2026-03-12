# GlobalTask Bank - k6 Stress Tests

Este directorio contiene scripts de pruebas de carga y estrés para la API de GlobalTask Bank usando [k6](https://k6.io/).

## Scripts Disponibles

### 1. `scale-test.js` ⭐ (Recomendado para 1M usuarios)
Simula un **día completo** de tráfico basándose en usuarios diarios:

| Escenario | VUs | Simulación |
|-----------|-----|------------|
| Noche | 2% pico | Usuarios mínimos |
| Amanecer | 10% | Comienzan a despertar |
| Mañana | 30% | Crecimiento |
| Mediodía | 60% | Hora pico #1 |
| **Pico máximo** | 100% | Máxima carga |
| Atardecer | 40% | Decreciente |
| Noche | 10% | Fin del día |

**Proyección automática**: Al final dice cuántos usuarios/día puede manejar tu sistema.

**Uso:**
```bash
# 1M usuarios/día (default), 200 VUs
docker compose run --rm k6 run k6/scale-test.js

# Escalar a más VUs
docker compose run --rm k6 run k6/scale-test.js -e MAX_VUS=500

# Especificar usuarios diarios
docker compose run --rm k6 run k6/scale-test.js -e DAILY_USERS=2000000 -e MAX_VUS=400
```

## Variables de Entorno

| Variable | Descripción | Default |
|----------|-------------|---------|
| `ADMIN_EMAIL` | Email del administrador | `admin@globaltask.com` |
| `ADMIN_PASSWORD` | Password del administrador | `password123` |
| `SUPABASE_URL` | URL de Supabase | `http://localhost:54321` |
| `API_URL` | URL de la API | `http://localhost` |
| `COUNTRY_ID` | ID del país para préstamos | `1` |
| `DAILY_USERS` | Usuarios diarios (scale-test) | `1000000` |
| `MAX_VUS` | Máximo VUs (scale-test) | `200` |
| `SUSTAIN_MINUTES` | Minutos en pico | `3` |

## Configuración de Escenarios

Los scripts incluyen escenarios configurables en `export const options`:

### Ramp Up (stress-test.js)
```
10s → 5 usuarios
20s → 10 usuarios
30s → 20 usuarios
30s → mantener 20 usuarios
10s → 0 usuarios
```

## Métricas Monitoreadas

- **login_failures**: Tasa de fallos en login
- **loan_create_failures**: Tasa de fallos en creación de préstamos
- **login_time**: Tiempo de respuesta del login (p95 < 2s)
- **loan_create_time**: Tiempo de creación (p95 < 5s)
- **http_req_failed**: Tasa de errores HTTP global

## Resultados Esperados

Con la configuración por defecto (20 VUs):
- ~120-180 préstamos creados por minuto
- Tiempo p95 de creación < 5s
- Tasa de éxito > 90%
