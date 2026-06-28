const API_BASE = '/api';
const DEFAULT_TIMEOUT = 30000;
const TOAST_DURATION = 4000;
const MAX_RETRIES = 2;
const RETRY_DELAY = 1000;

// ── connectivity ──

let online = navigator.onLine;
let offlineToast = null;

window.addEventListener('online', () => {
  online = true;
  if (offlineToast) { offlineToast.remove(); offlineToast = null; }
  showToast(i18next.t('error.backOnline'), 'info');
});

window.addEventListener('offline', () => {
  online = false;
  offlineToast = showToast(i18next.t('error.offline'), 'error');
});

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

function isRetryable(err) {
  return err.name === 'TypeError' || err.name === 'AbortError' && err.message !== 'TimeoutError';
}

// ── public API ──

async function apiGet(path, opts = {}) {
  const timeout = opts.timeout ?? DEFAULT_TIMEOUT;
  const retries = opts.retries ?? MAX_RETRIES;
  let lastErr;
  for (let i = 0; i <= retries; i++) {
    try {
      const init = {};
      if (timeout > 0) init.signal = AbortSignal.timeout(timeout);
      return handleResponse(await fetch(API_BASE + path, init));
    } catch (err) {
      lastErr = err;
      if (!isRetryable(err) || i >= retries) throw err;
      await new Promise(r => setTimeout(r, RETRY_DELAY * (i + 1)));
    }
  }
  throw lastErr;
}

async function apiPost(path, body, opts = {}) {
  const timeout = opts.timeout ?? DEFAULT_TIMEOUT;
  const retries = opts.retries ?? 0; // POST not idempotent by default
  let lastErr;
  for (let i = 0; i <= retries; i++) {
    try {
      const init = {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      };
      if (timeout > 0) init.signal = AbortSignal.timeout(timeout);
      return handleResponse(await fetch(API_BASE + path, init));
    } catch (err) {
      lastErr = err;
      if (!isRetryable(err) || i >= retries) throw err;
      await new Promise(r => setTimeout(r, RETRY_DELAY * (i + 1)));
    }
  }
  throw lastErr;
}

function isMobileDevice() {
  if (/Mobi|Android|iPhone|iPad|iPod/i.test(navigator.userAgent)) return true;
  if (navigator.maxTouchPoints > 0 && window.innerWidth < 1024) return true;
  return false;
}

function showQRModal(qrcodeUrl) {
  var existing = document.querySelector('.qr-modal-overlay');
  if (existing) existing.remove();

  var overlay = document.createElement('div');
  overlay.className = 'qr-modal-overlay';
  overlay.innerHTML =
    '<div class="qr-modal" role="dialog" aria-modal="true" aria-label="' + escapeHTML(i18next.t('payment.scanQR')) + '">' +
      '<button class="qr-modal-close" aria-label="' + escapeHTML(i18next.t('payment.qrClose')) + '">&times;</button>' +
      '<p class="qr-modal-title">' + escapeHTML(i18next.t('payment.scanQR')) + '</p>' +
      '<img class="qr-modal-img" src="' + escapeHTML(qrcodeUrl) + '" alt="QR Code">' +
      '<p class="qr-modal-hint">' + escapeHTML(i18next.t('payment.qrHint')) + '</p>' +
    '</div>';

  var prevFocus = document.activeElement;
  var closeBtn = overlay.querySelector('.qr-modal-close');

  var close = function() {
    overlay.remove();
    document.removeEventListener('keydown', trapFocus);
    if (prevFocus && typeof prevFocus.focus === 'function') {
      try { prevFocus.focus(); } catch (_) {}
    }
  };

  var trapFocus = function(e) {
    if (e.key !== 'Tab') return;
    var modal = overlay.querySelector('.qr-modal');
    var focusable = modal.querySelectorAll('button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])');
    if (focusable.length === 0) return;
    var first = focusable[0];
    var last = focusable[focusable.length - 1];
    if (e.shiftKey) {
      if (document.activeElement === first) { e.preventDefault(); last.focus(); }
    } else {
      if (document.activeElement === last) { e.preventDefault(); first.focus(); }
    }
  };

  overlay.addEventListener('click', function(e) { if (e.target === overlay) close(); });
  closeBtn.addEventListener('click', close);
  document.addEventListener('keydown', trapFocus);

  document.body.appendChild(overlay);
  closeBtn.focus();
}

async function goPay(orderID) {
  const data = await apiPost('/payments/checkout', { order_id: orderID });

  // Desktop + xunhu (qrcode_url present): show QR code for scanning
  if (data.qrcode_url && !isMobileDevice()) {
    showQRModal(data.qrcode_url);
    return;
  }

  // Mobile or dodo: redirect directly
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
  return el;
}

// ── namespace ──
const t = (key) => i18next.t(key);
window.Liki = { apiGet, apiPost, goPay, showQRModal, isMobileDevice, escapeHTML, renderMD, showToast, t, get isOnline() { return online; } };
