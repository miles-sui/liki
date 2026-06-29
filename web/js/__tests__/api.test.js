import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// ── minimal DOM mock for Node ──

function encodeText(s) {
  return String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

function createMockDocument() {
  const allElements = [];
  const body = {
    _id: 0, _tag: 'body', _text: '', _html: '', _className: '', _children: [],
    appendChild(c) { this._children.push(c); },
  };
  allElements.push(body);

  const doc = {
    body,
    createElement(tag) {
      const el = {
        _id: allElements.length, _tag: tag,
        _text: '', _html: '', _className: '',
        _children: [],
        set textContent(v) { this._text = String(v); this._html = encodeText(v); },
        get textContent() { return this._text; },
        set innerHTML(v) { this._html = v; },
        get innerHTML() { return this._html; },
        set className(v) { this._className = v; },
        get className() { return this._className; },
        appendChild(c) { this._children.push(c); },
        remove() {
          for (const e of allElements) {
            const idx = (e._children || []).indexOf(this);
            if (idx !== -1) { e._children.splice(idx, 1); }
          }
          const ai = allElements.indexOf(this);
          if (ai !== -1) allElements.splice(ai, 1);
        },
      };
      allElements.push(el);
      return el;
    },
    querySelector(sel) {
      for (const e of allElements) {
        if (sel === '.toast' && e._className && e._className.includes('toast')) return e;
        if (e._className && e._className.split(' ').every(c => sel.includes(c))) return e;
      }
      // also check children
      for (const e of allElements) {
        for (const c of (e._children || [])) {
          if (sel === '.toast' && c._className && c._className.includes('toast')) return c;
          if (c._className && c._className.split(' ').every(cls => sel.includes(cls))) return c;
        }
      }
      return null;
    },
  };
  return doc;
}

// simplified AbortSignal.timeout mock
function fakeAbortSignal(ms) {
  const ctrl = new AbortController();
  setTimeout(() => ctrl.abort(new DOMException('Timeout', 'TimeoutError')), ms);
  return ctrl.signal;
}

let mockDoc;
beforeEach(() => {
  mockDoc = createMockDocument();
  vi.stubGlobal('document', mockDoc);
  vi.stubGlobal('i18next', { t: (k) => k });
  vi.stubGlobal('marked', { parse: (text) => `<h2>${text}</h2>` });
  vi.stubGlobal('DOMPurify', { sanitize: (html) => html });
  vi.stubGlobal('AbortSignal', { timeout: fakeAbortSignal });
});

afterEach(() => {
  vi.unstubAllGlobals();
});

// ── functions under test (copied from api.js) ──

const API_BASE = '/api';
const DEFAULT_TIMEOUT = 30000;
const MAX_RETRIES = 2;
const RETRY_DELAY = 1000;

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
  // iPadOS 13+ spoofs as desktop Safari — check touch points
  if (navigator.maxTouchPoints > 1) return true;
  return false;
}

function showQRModal(qrcodeUrl, fallbackUrl) {
  var existing = document.querySelector('.qr-modal-overlay');
  if (existing) existing.remove();

  var overlay = document.createElement('div');
  overlay.className = 'qr-modal-overlay';
  overlay.innerHTML =
    '<div class="qr-modal" role="dialog" aria-modal="true" aria-label="' + escapeHTML(i18next.t('payment.scanQR')) + '">' +
      '<button class="qr-modal-close" aria-label="' + escapeHTML(i18next.t('payment.qrClose')) + '">&times;</button>' +
      '<p class="qr-modal-title">' + escapeHTML(i18next.t('payment.scanQR')) + '</p>' +
      '<img class="qr-modal-img" alt="QR Code">' +
      '<p class="qr-modal-hint">' + escapeHTML(i18next.t('payment.qrHint')) + '</p>' +
    '</div>';

  var img = overlay.querySelector('.qr-modal-img');
  img.src = qrcodeUrl;
  if (fallbackUrl) {
    img.onerror = function() { this.onerror = null; location.href = fallbackUrl; };
  }

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
  var data = await apiPost('/payments/checkout', { order_id: orderID });
  if (!data) throw new Error(i18next.t('error.noCheckoutUrl'));

  if (data.qrcode_url && !isMobileDevice()) {
    showQRModal(data.qrcode_url, data.checkout_url);
    return;
  }

  var url = data.checkout_url;
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
  setTimeout(() => el.remove(), 4000);
  return el;
}

// ── helpers ──

function mockFetch(status, data, ok = true) {
  return Promise.resolve({ ok, status, json: () => Promise.resolve(data) });
}

// ── tests ──

describe('handleResponse', () => {
  it('returns data.data on success', async () => {
    const resp = await mockFetch(200, { data: { hello: 'world' } });
    expect(await handleResponse(resp)).toEqual({ hello: 'world' });
  });

  it('throws with error message from response body on non-ok', async () => {
    const resp = await mockFetch(422, { error: { message: 'bad input' } }, false);
    await expect(handleResponse(resp)).rejects.toThrow('bad input');
  });

  it('throws with fallback when body has no error message', async () => {
    const resp = await mockFetch(500, {}, false);
    await expect(handleResponse(resp)).rejects.toThrow('error.requestFailed');
  });

  it('throws with fallback when body is not JSON', async () => {
    const resp = { ok: false, json: () => Promise.reject(new Error('parse')) };
    await expect(handleResponse(resp)).rejects.toThrow('error.requestFailed');
  });

  it('throws when response has no data envelope', async () => {
    const resp = await mockFetch(200, { not_data: true });
    await expect(handleResponse(resp)).rejects.toThrow('Unexpected API response');
  });

  it('throws when response body is null', async () => {
    const resp = await mockFetch(200, null);
    await expect(handleResponse(resp)).rejects.toThrow('Unexpected API response');
  });
});

describe('apiGet', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn());
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it('calls fetch with API_BASE + path and returns data', async () => {
    fetch.mockResolvedValue({ ok: true, json: () => Promise.resolve({ data: { result: 'ok' } }) });
    expect(await apiGet('/test')).toEqual({ result: 'ok' });
    expect(fetch).toHaveBeenCalledWith('/api/test', expect.objectContaining({}));
  });

  it('sets AbortSignal.timeout from custom timeout option', async () => {
    fetch.mockResolvedValue({ ok: true, json: () => Promise.resolve({ data: {} }) });
    await apiGet('/test', { timeout: 5000 });
    expect(fetch).toHaveBeenCalledWith('/api/test', expect.objectContaining({
      signal: expect.any(Object),
    }));
  });

  it('uses DEFAULT_TIMEOUT when no timeout specified', async () => {
    fetch.mockResolvedValue({ ok: true, json: () => Promise.resolve({ data: {} }) });
    await apiGet('/test');
    expect(fetch.mock.calls[0][1].signal).toBeDefined();
  });

  it('skips signal when timeout is 0', async () => {
    fetch.mockResolvedValue({ ok: true, json: () => Promise.resolve({ data: {} }) });
    await apiGet('/test', { timeout: 0 });
    expect(fetch.mock.calls[0][1].signal).toBeUndefined();
  });
});

describe('apiPost', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn());
  });

  it('sends POST with JSON body and Content-Type header', async () => {
    fetch.mockResolvedValue({ ok: true, json: () => Promise.resolve({ data: { id: 1 } }) });
    const result = await apiPost('/submit', { name: 'test' });
    const [url, init] = fetch.mock.calls[0];
    expect(url).toBe('/api/submit');
    expect(init.method).toBe('POST');
    expect(init.headers['Content-Type']).toBe('application/json');
    expect(JSON.parse(init.body)).toEqual({ name: 'test' });
    expect(result).toEqual({ id: 1 });
  });

  it('propagates fetch errors', async () => {
    fetch.mockRejectedValue(new Error('network error'));
    await expect(apiPost('/submit', {})).rejects.toThrow('network error');
  });
});

describe('goPay', () => {
  let qrUrls = [];
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn());
    vi.stubGlobal('window', { location: { href: '' } });
    // Default: desktop (non-mobile) for QR tests
    vi.stubGlobal('navigator', { userAgent: '', maxTouchPoints: 0 });
    vi.stubGlobal('innerWidth', 1920);
    qrUrls = [];
  });

  it('throws when checkout_url is missing from response', async () => {
    fetch.mockResolvedValue({ ok: true, json: () => Promise.resolve({ data: {} }) });
    await expect(goPay('order-1')).rejects.toThrow('error.noCheckoutUrl');
  });

  it('redirects to checkout_url on success', async () => {
    fetch.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ data: { checkout_url: 'https://pay.example.com/checkout/xyz' } }),
    });
    await goPay('order-1');
    expect(window.location.href).toBe('https://pay.example.com/checkout/xyz');
  });

  it('shows QR modal on desktop when qrcode_url is present', async () => {
    var calledUrl = null, calledFallback = null;
    var orig = showQRModal;
    showQRModal = function(url, fb) { calledUrl = url; calledFallback = fb; };
    try {
      fetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ data: { checkout_url: 'https://pay.example.com/checkout/xyz', qrcode_url: 'https://qr.example.com/qr.png' } }),
      });
      await goPay('order-1');
      expect(window.location.href).toBe('');
      expect(calledUrl).toBe('https://qr.example.com/qr.png');
      expect(calledFallback).toBe('https://pay.example.com/checkout/xyz');
    } finally {
      showQRModal = orig;
    }
  });

  it('redirects on mobile even when qrcode_url is present', async () => {
    vi.stubGlobal('navigator', { userAgent: 'iPhone', maxTouchPoints: 1 });
    vi.stubGlobal('innerWidth', 375);
    fetch.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ data: { checkout_url: 'https://pay.example.com/checkout/xyz', qrcode_url: 'https://qr.example.com/qr.png' } }),
    });
    await goPay('order-1');
    expect(window.location.href).toBe('https://pay.example.com/checkout/xyz');
  });

  it('redirects on desktop when qrcode_url is absent (omitempty)', async () => {
    fetch.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ data: { checkout_url: 'https://pay.example.com/checkout/xyz' } }),
    });
    await goPay('order-1');
    expect(window.location.href).toBe('https://pay.example.com/checkout/xyz');
  });

  it('redirects on desktop when qrcode_url is empty string', async () => {
    fetch.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ data: { checkout_url: 'https://pay.example.com/checkout/xyz', qrcode_url: '' } }),
    });
    await goPay('order-1');
    expect(window.location.href).toBe('https://pay.example.com/checkout/xyz');
  });

  it('propagates network error from apiPost', async () => {
    fetch.mockRejectedValue(new TypeError('Failed to fetch'));
    await expect(goPay('order-1')).rejects.toThrow('Failed to fetch');
  });
});

describe('apiGet retry', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn());
  });

  it('does not retry on HTTP errors', async () => {
    fetch.mockResolvedValue({ ok: false, status: 422, json: () => Promise.resolve({ error: { message: 'bad input' } }) });
    await expect(apiGet('/test')).rejects.toThrow('bad input');
    expect(fetch).toHaveBeenCalledTimes(1);
  });

  it('accepts retries option and propagates error when all fail', async () => {
    // With retries:0, no retry
    fetch.mockRejectedValue(new TypeError('Failed'));
    await expect(apiGet('/test', { retries: 0 })).rejects.toThrow('Failed');
    expect(fetch).toHaveBeenCalledTimes(1);
  });
});

describe('apiPost retry', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn());
  });

  it('does not retry by default', async () => {
    fetch.mockRejectedValue(new TypeError('Failed to fetch'));
    await expect(apiPost('/submit', {})).rejects.toThrow('Failed to fetch');
    expect(fetch).toHaveBeenCalledTimes(1);
  });

  it('accepts retries option', async () => {
    fetch.mockRejectedValue(new TypeError('Failed'));
    await expect(apiPost('/submit', {}, { retries: 1 })).rejects.toThrow('Failed');
    expect(fetch).toHaveBeenCalledTimes(2); // initial + 1 retry
  });
});

describe('isRetryable', () => {
  it('returns true for TypeError', () => {
    expect(isRetryable(new TypeError('Failed to fetch'))).toBe(true);
  });

  it('returns true for AbortError (not TimeoutError)', () => {
    const e = new DOMException('The operation was aborted', 'AbortError');
    expect(isRetryable(e)).toBe(true);
  });

  it('returns false for regular Error', () => {
    expect(isRetryable(new Error('something'))).toBe(false);
  });

  it('returns false for SyntaxError', () => {
    expect(isRetryable(new SyntaxError('bad JSON'))).toBe(false);
  });
});

describe('escapeHTML', () => {
  it('escapes angle brackets and ampersand', () => {
    expect(escapeHTML('<script>alert("xss")</script>'))
      .toBe('&lt;script&gt;alert(&quot;xss&quot;)&lt;/script&gt;');
  });

  it('returns plain text unchanged', () => {
    expect(escapeHTML('hello')).toBe('hello');
  });

  it('returns empty string for empty input', () => {
    expect(escapeHTML('')).toBe('');
  });
});

describe('renderMD', () => {
  it('returns sanitized HTML from marked.parse', () => {
    expect(renderMD('# Hello')).toBe('<h2># Hello</h2>');
  });

  it('falls back to escaped HTML when marked.parse throws', () => {
    vi.stubGlobal('marked', { parse: () => { throw new Error('parse error'); } });
    expect(renderMD('<b>bold</b>')).toBe('&lt;b&gt;bold&lt;/b&gt;');
  });
});

describe('showToast', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it('creates toast element with correct class and appends to body', () => {
    showToast('error message', 'error');
    const toast = document.querySelector('.toast.toast-error');
    expect(toast).not.toBeNull();
    expect(toast.textContent).toBe('error message');
  });

  it('defaults type to "error"', () => {
    showToast('msg');
    expect(document.querySelector('.toast.toast-error')).not.toBeNull();
  });

  it('returns the toast element', () => {
    const el = showToast('msg');
    expect(el).not.toBeNull();
    expect(el._tag).toBe('div');
    expect(el._className).toBe('toast toast-error');
  });

  it('removes toast after 4000ms', () => {
    showToast('msg');
    expect(document.querySelector('.toast')).not.toBeNull();
    vi.advanceTimersByTime(4000);
    expect(document.querySelector('.toast')).toBeNull();
  });
});
