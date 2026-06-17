const API_BASE = '/api';
const DEFAULT_TIMEOUT = 30000;
const TOAST_DURATION = 4000;

// ── response helpers ──

async function handleResponse(resp) {
  if (!resp.ok) {
    let msg = i18next.t('error.requestFailed');
    try { const body = await resp.json(); msg = body.error?.message || msg; } catch (_) {}
    throw new Error(msg);
  }
  const data = await resp.json();
  if (!data || !('data' in data)) throw new Error('Unexpected API response');
  return data.data;
}

// ── public API ──

async function apiGet(path, opts = {}) {
  const timeout = opts.timeout ?? DEFAULT_TIMEOUT;
  const init = {};
  if (timeout > 0) init.signal = AbortSignal.timeout(timeout);
  return handleResponse(await fetch(API_BASE + path, init));
}

async function apiPost(path, body, opts = {}) {
  const timeout = opts.timeout ?? DEFAULT_TIMEOUT;
  const init = {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  };
  if (timeout > 0) init.signal = AbortSignal.timeout(timeout);
  return handleResponse(await fetch(API_BASE + path, init));
}

async function goPay(orderID) {
  const data = await apiPost('/payments/checkout', { order_id: orderID });
  const url = data.checkout_url;
  if (!url) throw new Error(i18next.t('error.noCheckoutUrl'));
  window.location.href = url;
}

function escapeHTML(text) {
  const d = document.createElement('div');
  d.textContent = text;
  return d.innerHTML;
}

function renderMD(text) {
  try {
    const raw = marked.parse(text);
    return DOMPurify.sanitize(raw, {
      ALLOWED_TAGS: ['h2', 'h3', 'h4', 'p', 'br', 'strong', 'em', 'ul', 'ol', 'li', 'code', 'pre', 'blockquote', 'table', 'thead', 'tbody', 'tr', 'th', 'td', 'hr'],
      ALLOWED_ATTR: [],
    });
  } catch (_) { return escapeHTML(text); }
}

function showToast(msg, type = 'error') {
  const el = document.createElement('div');
  el.className = 'toast toast-' + type;
  el.textContent = msg;
  document.body.appendChild(el);
  setTimeout(() => el.remove(), TOAST_DURATION);
}

// ── namespace ──
window.Liki = { apiGet, apiPost, goPay, escapeHTML, renderMD, showToast };
