/**
 * GlobalTask Bank - Production Scale Test
 *
 * Simula escenarios de producción basados en usuarios reales diarios.
 *
 * CÁLCULOS:
 * - 1M usuarios/día ≈ 11.5 RPS promedio
 * - Hora pico (3x) ≈ 35 RPS
 * - Pico absoluto (10x) ≈ 115 RPS
 *
 * Para testing, usamos VUs para simular usuarios concurrentes.
 * Un VU hace ~1 request cada 1-3 segundos, así que:
 * - 100 VUs ≈ 30-100 RPS
 * - 500 VUs ≈ 150-500 RPS
 *
 * Uso:
 *   k6 run k6/scale-test.js
 *
 *   # Simular 1M usuarios/día con pico de 500 VUs
 *   k6 run k6/scale-test.js -e MAX_VUS=500
 *
 *   # Simular escenario extremo (2M usuarios/día)
 *   k6 run k6/scale-test.js -e MAX_VUS=1000 -e DAILY_USERS=2000000
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter, Gauge } from 'k6/metrics';

// ==========================================
// Configuración
// ==========================================

// Parámetros del negocio
const DAILY_USERS = parseInt(__ENV.DAILY_USERS) || 1000000;  // Usuarios por día
const MAX_VUS = parseInt(__ENV.MAX_VUS) || 200;               // Máx VUs a testear
const SUSTAIN_MINUTES = parseInt(__ENV.SUSTAIN_MINUTES) || 3;  // Minutos en pico

// URLs
const ADMIN_EMAIL = __ENV.ADMIN_EMAIL || 'admin@globaltask.com';
const ADMIN_PASSWORD = __ENV.ADMIN_PASSWORD || 'password123';
const SUPABASE_URL = __ENV.SUPABASE_URL || 'http://localhost:54321';
const API_URL = __ENV.API_URL || 'http://localhost';
const COUNTRY_ID = __ENV.COUNTRY_ID || '1';

// ==========================================
// Métricas
// ==========================================

const loginFailRate = new Rate('login_failures');
const loanFailRate = new Rate('loan_failures');
const readFailRate = new Rate('read_failures');
const loginTime = new Trend('login_time');
const loanTime = new Trend('loan_time');
const readTime = new Trend('read_time');
const loansCreated = new Counter('loans_created');
const requestsTotal = new Counter('requests_total');

// ==========================================
// Calculadora de Escala
// ==========================================

function calculateScale() {
  // Métricas derivadas
  const rps = DAILY_USERS / 86400;  // Requests por segundo promedio
  const peakRps = rps * 10;          // Pico (10x promedio)

  // VUs necesarios para lograr cierto RPS
  // Asumiendo 1 VU hace ~1 request cada 2 segundos = 0.5 RPS/VU
  const peakVUs = Math.ceil(peakRps / 0.5);

  // Pero el sistema real tiene think time variable
  // Un VU puede hacer 0.3-1 request/segundo dependiendo del escenario
  const targetVUs = Math.min(MAX_VUS, peakVUs * 2); // Margen de seguridad

  return {
    dailyUsers: DAILY_USERS,
    avgRps: Math.round(rps * 100) / 100,
    peakRps: Math.round(peakRps * 100) / 100,
    targetVUs: targetVUs,
    expectedRps: targetVUs * 0.5,
  };
}

// ==========================================
// Login con Cache
// ==========================================

let cachedToken = null;
let tokenExpiry = 0;

function login() {
  const url = `${SUPABASE_URL}/auth/v1/token?grant_type=password`;
  const payload = JSON.stringify({
    email: ADMIN_EMAIL,
    password: ADMIN_PASSWORD,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const startTime = new Date();
  const response = http.post(url, payload, params);
  const duration = new Date() - startTime;

  loginTime.add(duration);

  const success = check(response, { 'login status 200': (r) => r.status === 200 });
  loginFailRate.add(!success);
  if (success) {
    try {
      const body = JSON.parse(response.body);
      tokenExpiry = Date.now() + 5 * 60 * 1000;
      return body.access_token;
    } catch (e) {}
  }

}

// ==========================================
// Operaciones CRUD
// ==========================================

function createLoan(token) {
  const url = `${API_URL}/api/v1/applications`;

  const randomId = Math.floor(Math.random() * 100000000);
  const payload = JSON.stringify({
    country_id: parseInt(COUNTRY_ID),
    borrower_name: `User ${randomId}`,
    identity_document: `ID${randomId}`,
    requested_amount: Math.floor(Math.random() * 20000) + 1000,
    monthly_income: Math.floor(Math.random() * 8000) + 1000,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    timeout: '15s',
  };

  const startTime = new Date();
  const response = http.post(url, payload, params);
  const duration = new Date() - startTime;

  loanTime.add(duration);
  requestsTotal.add(1);

  const success = check(response, {
    'loan created': (r) => r.status === 200 || r.status === 201
  });

  loanFailRate.add(!success);

  if (success) {
    loansCreated.add(1);
  }

  return success;
}

function listLoans(token, page = 1) {
  const url = `${API_URL}/api/v1/applications?limit=20&offset=${(page - 1) * 20}`;

  const params = {
    headers: {
      'Authorization': `Bearer ${token}`,
    },
    timeout: '10s',
  };

  const startTime = new Date();
  const response = http.get(url, params);
  const duration = new Date() - startTime;

  readTime.add(duration);
  requestsTotal.add(1);

  const success = check(response, { 'list ok': (r) => r.status === 200 });
  readFailRate.add(!success);

  return success;
}

function getLoan(token, id) {
  const url = `${API_URL}/api/v1/applications/${id}`;

  const params = {
    headers: {
      'Authorization': `Bearer ${token}`,
    },
    timeout: '5s',
  };

  const startTime = new Date();
  const response = http.get(url, params);
  const duration = new Date() - startTime;

  readTime.add(duration);
  requestsTotal.add(1);

  return response.status === 200;
}

// ==========================================
// Configuración del Test
// ==========================================

const scale = calculateScale();

export const options = {
  // Mejor manejo de conexiones para alta carga
  noVUConnReuse: false, // Reuse conexiones para mejor performance

  scenarios: {
    // Escenario: Simular día completo en tiempo comprimido
    day_simulation: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        // Noche: muy pocos usuarios (2% de pico)
        { duration: '30s', target: Math.round(scale.targetVUs * 0.02) },

        // Amanecer: usuarios comienzan a despertar (10% de pico)
        { duration: '30s', target: Math.round(scale.targetVUs * 0.10) },

        // Mañana: crecimiento gradual (30% de pico)
        { duration: '1m', target: Math.round(scale.targetVUs * 0.30) },

        // Mediodía: hora pico #1 (60% de pico)
        { duration: '1m', target: Math.round(scale.targetVUs * 0.60) },

        // Tarde: hora pico #2 (100% de pico - PICO MÁXIMO)
        { duration: `${SUSTAIN_MINUTES}m`, target: scale.targetVUs },

        // Atardecer: disminución (40% de pico)
        { duration: '1m', target: Math.round(scale.targetVUs * 0.40) },

        // Noche: vuelta a mínimos (10% de pico)
        { duration: '1m', target: Math.round(scale.targetVUs * 0.10) },

        // Fin del día (5% de pico)
        { duration: '30s', target: Math.round(scale.targetVUs * 0.05) },
      ],
      gracefulRampDown: '30s',
    },
  },

  // Umbrales basados en expectativas de producción
  thresholds: {
    // Login: crítico, debe ser confiable
    'login_failures': ['rate<0.05'],      // 95% éxito

    // Writes: pueden tener más fallos por limitaciones de BD
    'loan_failures': ['rate<0.10'],        // 90% éxito

    // Reads: deben ser muy confiables
    'read_failures': ['rate<0.02'],        // 98% éxito

    // Tiempos - metas ambiciosas para producción
    'login_time': ['p(95)<3000'],          // 3s p95
    'loan_time': ['p(95)<5000'],           // 5s p95
    'read_time': ['p(95)<1000'],           // 1s p95

    //throughput
    http_req_duration: ['p(95)<5000'],     // 5s p95 global
  },
};

// ==========================================
// Setup
// ==========================================

export function setup() {
  console.log('╔═══════════════════════════════════════════════════════════╗');
  console.log('║       PRODUCTION SCALE TEST - CONFIGURATION              ║');
  console.log('╠═══════════════════════════════════════════════════════════╣');
  console.log(`║  Usuarios diarios:     ${DAILY_USERS.toString().padEnd(30)}║`);
  console.log(`║  RPS promedio:        ${scale.avgRps.toString().padEnd(30)}║`);
  console.log(`║  RPS pico estimado:   ${scale.peakRps.toString().padEnd(30)}║`);
  console.log(`║  VUs objetivo:        ${scale.targetVUs.toString().padEnd(30)}║`);
  console.log(`║  RPS esperado:        ${scale.expectedRps.toString().padEnd(30)}║`);
  console.log('╠═══════════════════════════════════════════════════════════╣');
  console.log(`║  API URL:             ${API_URL.padEnd(38)}║`);
  console.log(`║  Supabase URL:        ${SUPABASE_URL.padEnd(38)}║`);
  console.log('╚═══════════════════════════════════════════════════════════╝');
  console.log('');

  const token = login();

  return {
    token: token || login(), // Retry si falla
    valid: !!token,
    scale,
  };
}

// ==========================================
// Test - Simula usuario real
// ==========================================

export default function(data) {
  if (!data || !data.valid) {
    sleep(1);
    return;
  }

  const token = login();
  if (!token) {
    sleep(1);
    return;
  }

  // Simular comportamiento de usuario real
  // Un usuario típico: ~70% reads, ~30% writes
  const rand = Math.random();

  if (rand < 0.70) {
    // 70%: Ver lista de préstamos
    listLoans(token, Math.floor(Math.random() * 50) + 1);
  } else if (rand < 0.95) {
    // 25%: Crear nuevo préstamo
    createLoan(token);
  } else {
    // 5%: Ver detalle de un préstamo específico
    getLoan(token, '00000000-0000-0000-0000-000000000001');
  }

  // Think time típico: 0.5 - 3 segundos entre acciones
  sleep(Math.random() * 2.5 + 0.5);
}

// ==========================================
// Results
// ==========================================

export function handleSummary(data) {
  const duration = data.state.testRunDurationMs / 1000;
  const totalReqs = data.metrics.http_reqs.values.count;
  const rps = totalReqs / duration;

  const loans = data.metrics.loans_created?.values?.count || 0;
  const loginFails = data.metrics.login_failures?.values?.rate || 0;
  const loanFails = data.metrics.loan_failures?.values?.rate || 0;
  const readFails = data.metrics.read_failures?.values?.rate || 0;

  console.log('\n╔═══════════════════════════════════════════════════════════╗');
  console.log('║                    TEST RESULTS                           ║');
  console.log('╠═══════════════════════════════════════════════════════════╣');
  console.log(`║  Duración:            ${duration.toFixed(1)}s`.padEnd(50) + '║');
  console.log(`║  Requests totales:    ${totalReqs}`.padEnd(50) + '║');
  console.log(`║  Throughput promedio: ${rps.toFixed(1)} req/s`.padEnd(50) + '║');
  console.log(`║  Préstamos creados:   ${loans}`.padEnd(50) + '║');
  console.log('╠═══════════════════════════════════════════════════════════╣');
  console.log('║  TASAS DE ERROR:');
  console.log(`║    Login failures:     ${(loginFails * 100).toFixed(1)}%`.padEnd(50) + '║');
  console.log(`║    Loan failures:      ${(loanFails * 100).toFixed(1)}%`.padEnd(50) + '║');
  console.log(`║    Read failures:      ${(readFails * 100).toFixed(1)}%`.padEnd(50) + '║');
  console.log('╠═══════════════════════════════════════════════════════════╣');

  // Proyección
  const projectedDaily = Math.round(rps * 86400);
  console.log('║  PROYECCIÓN A 24H:');
  console.log(`║    Con ${rps.toFixed(1)} req/s → ${projectedDaily.toLocaleString()} usuarios/día`.padEnd(50) + '║');
  console.log('╚═══════════════════════════════════════════════════════════╝');

  return {
    stdout: 'Test completado',
    'results.json': JSON.stringify({
      config: {
        dailyUsers: DAILY_USERS,
        targetVUs: MAX_VUS,
      },
      metrics: {
        duration,
        totalRequests: totalReqs,
        rps,
        loansCreated: loans,
      },
      rates: {
        login: loginFails,
        loan: loanFails,
        read: readFails,
      },
      projection: {
        rps,
        dailyUsers: projectedDaily,
      },
    }, null, 2),
  };
}
