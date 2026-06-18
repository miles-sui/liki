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
    solar_time: params.solar_time || '2000-06-15T12:00:00+08:00',
    gender: params.gender || 'male',
  });
}

/** POST /api/bazi/bond — free bond analysis (docs/bazi.md) */
export async function baziBond(aParams = {}, bParams = {}) {
  return req('POST', '/bazi/bond', {
    a: { solar_time: '2000-06-15T12:00:00+08:00', gender: 'male', ...aParams },
    b: { solar_time: '1999-03-20T08:00:00+08:00', gender: 'female', ...bParams },
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
