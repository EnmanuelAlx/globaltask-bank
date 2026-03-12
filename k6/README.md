# GlobalTask Bank - k6 Stress Tests

This directory contains load and stress testing scripts for the GlobalTask Bank API using [k6](https://k6.io/).

## Available Scripts

### 1. `scale-test.js` ⭐ (Recommended for 1M users)
Simulates a **full day** of traffic based on daily users:

| Scenario | VUs | Simulation |
|-----------|-----|------------|
| Night | 2% peak | Minimum users |
| Dawn | 10% | Starting to wake up |
| Morning | 30% | Growth |
| Noon | 60% | Peak hour #1 |
| **Maximum peak** | 100% | Maximum load |
| Evening | 40% | Decreasing |
| Night | 10% | End of day |

**Automatic projection**: At the end it tells you how many users/day your system can handle.

**Usage:**
```bash
# 1M users/day (default), 200 VUs
docker compose run --rm k6 run k6/scale-test.js

# Scale to more VUs
docker compose run --rm k6 run k6/scale-test.js -e MAX_VUS=500

# Specify daily users
docker compose run --rm k6 run k6/scale-test.js -e DAILY_USERS=2000000 -e MAX_VUS=400
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `ADMIN_EMAIL` | Admin email | `admin@globaltask.com` |
| `ADMIN_PASSWORD` | Admin password | `password123` |
| `SUPABASE_URL` | Supabase URL | `http://localhost:54321` |
| `API_URL` | API URL | `http://localhost` |
| `COUNTRY_ID` | Country ID for loans | `1` |
| `DAILY_USERS` | Daily users (scale-test) | `1000000` |
| `MAX_VUS` | Maximum VUs (scale-test) | `200` |
| `SUSTAIN_MINUTES` | Minutes at peak | `3` |

## Scenario Configuration

Scripts include configurable scenarios in `export const options`:

### Ramp Up (stress-test.js)
```
10s → 5 users
20s → 10 users
30s → 20 users
30s → maintain 20 users
10s → 0 users
```

## Monitored Metrics

- **login_failures**: Login failure rate
- **loan_create_failures**: Loan creation failure rate
- **login_time**: Login response time (p95 < 2s)
- **loan_create_time**: Creation time (p95 < 5s)
- **http_req_failed**: Global HTTP error rate

## Expected Results

With default configuration (20 VUs):
- ~120-180 loans created per minute
- p95 creation time < 5s
- Success rate > 90%
