// payment.js — Checkout flow with QR code support for desktop.
// Uses apiPost from api.js and i18next for translations.

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
      '<img class="qr-modal-img" src="' + escapeHTML(qrcodeUrl) + '" alt="QR Code"' +
        ' onerror="this.onerror=null; location.href=document.querySelector(\'#purchase-btn\').dataset.checkoutUrl||\'\';">' +
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
  var data = await apiPost('/payments/checkout', { order_id: orderID });
  if (!data) throw new Error(i18next.t('error.noCheckoutUrl'));

  // Desktop with QR code available: show modal
  if (data.qrcode_url && !isMobileDevice()) {
    // Store checkout_url as fallback for image load failures
    var btn = document.getElementById('purchase-btn');
    if (btn && data.checkout_url) btn.dataset.checkoutUrl = data.checkout_url;
    showQRModal(data.qrcode_url);
    return;
  }

  // Mobile or no QR: redirect
  var url = data.checkout_url;
  if (!url) throw new Error(i18next.t('error.noCheckoutUrl'));
  window.location.href = url;
}
