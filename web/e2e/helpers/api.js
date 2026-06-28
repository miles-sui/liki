// API helpers for E2E test data setup.
// All engine endpoints use JSON-RPC (POST /api/jsonrpc).
// Business endpoints use REST paths.

const API = 'http://localhost:8080/api';
const RPC = `${API}/jsonrpc`;
let rpcID = 0;

async function req(method, path, body) {
  const url = `${API}${path}`;
  const opts = { method, headers: { 'Content-Type': 'application/json' } };
  if (body && method !== 'GET') opts.body = JSON.stringify(body);
  const resp = await fetch(url, opts);
  const data = await resp.json();
  if (!resp.ok) throw new Error(data.error?.message || `HTTP ${resp.status}`);
  return data.data;
}

async function rpc(method, params = {}) {
  rpcID++;
  const resp = await fetch(RPC, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ jsonrpc: '2.0', id: rpcID, method, params }),
  });
  const data = await resp.json();
  if (data.error) throw new Error(data.error.message || `RPC error ${data.error.code}`);
  return data.result.data;
}

// ---- JSON-RPC engine endpoints ----

/** rpc bazi.chart — free full BaZi chart */
export async function baziChart(params = {}) {
  return rpc('bazi.chart', {
    birth: { time: params.solar_time || '2000-06-15T12:00:00+08:00', longitude: params.longitude || 116.4 },
    gender: params.gender || 'male',
  });
}

/** rpc bazi.bond — free bond analysis */
export async function baziBond(aParams = {}, bParams = {}) {
  return rpc('bazi.bond', {
    a: { birth: { time: '2000-06-15T12:00:00+08:00', longitude: 116.4, ...aParams.birth }, gender: 'male', ...aParams },
    b: { birth: { time: '1999-03-20T08:00:00+08:00', longitude: 116.4, ...bParams.birth }, gender: 'female', ...bParams },
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
