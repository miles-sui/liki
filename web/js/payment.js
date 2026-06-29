// payment.js — Checkout flow with QR code support for desktop.
// Uses apiPost from api.js and i18next for translations.

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

  // Desktop with QR code available: show modal
  if (data.qrcode_url && !isMobileDevice()) {
    showQRModal(data.qrcode_url, data.checkout_url);
    return;
  }

  // Mobile or no QR: redirect
  var url = data.checkout_url;
  if (!url) throw new Error(i18next.t('error.noCheckoutUrl'));
  window.location.href = url;
}
