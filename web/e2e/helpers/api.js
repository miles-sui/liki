// API helpers for E2E test data setup.
// Request shapes are derived from the actual JS source and API docs (web/docs/*.md).

const API = 'http://localhost:8080/api';

async function req(method, path, body) {
  const url = `${API}${path}`;
  const opts = { method, headers: { 'Content-Type': 'application/json' } };
  if (body && method !== 'GET') opts.body = JSON.stringify(body);
  const resp = await fetch(url, opts);
  const data = await resp.json();
  if (!resp.ok) throw new Error(data.error?.message || `HTTP ${resp.status}`);
  return data.data;
}

// ---- Free API (from docs) ----

/** POST /api/bazi/chart — free full BaZi chart (docs/bazi.md) */
export async function baziChart(params = {}) {
  return req('POST', '/bazi/chart', {
    year: params.year || 2000,
    month: params.month || 6,
    day: params.day || 15,
    hour: params.hour ?? 12,
    minute: params.minute ?? 0,
    longitude: params.longitude ?? 120,
    timezone: params.timezone ?? 8,
    gender: params.gender || 'male',
  });
}

/** POST /api/bazi/bond — free bond analysis (docs/bazi.md) */
export async function baziBond(aParams = {}, bParams = {}) {
  return req('POST', '/bazi/bond', {
    a: { year: 2000, month: 6, day: 15, hour: 12, minute: 0, longitude: 120, timezone: 8, gender: 'male', ...aParams },
    b: { year: 1999, month: 3, day: 20, hour: 8, minute: 0, longitude: 120, timezone: 8, gender: 'female', ...bParams },
  });
}

/** GET /api/health */
export async function health() {
  return req('GET', '/health');
}

// ---- Paid API ----

/** POST /api/payments/checkout — create payment */
export async function checkout(orderID, email = '') {
  return req('POST', '/payments/checkout', { order_id: orderID, email });
}

/** GET /api/report/:orderID — fetch report (report.js:65) */
export async function getReport(orderID) {
  return req('GET', `/report/${orderID}`);
}
