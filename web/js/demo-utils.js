// Shared utilities for demo pages (chart.html, compatibility.html, naming.html)

function shareDemo() {
  const t = window.Liki.t;
  if (navigator.share) {
    navigator.share({ title: document.title, url: location.href }).catch(function() {});
  } else {
    navigator.clipboard.writeText(location.href).then(function() {
      window.Liki.showToast(t('demo.copied'), 'success');
    }).catch(function() {});
  }
}

(function() {
  var el = document.getElementById('print-date');
  if (el) { el.textContent = new Date().toISOString().slice(0, 10); }
})();

// Set page title and meta description from i18n.
(function() {
  var t = window.Liki.t;
  var m = location.pathname.match(/\/(chart|compatibility|naming)\.html/i);
  var product = m ? (m[1] === 'compatibility' ? 'bond' : m[1]) : 'chart';
  var titleKey = 'demo.title.' + product;
  var descKey = 'demo.metaDesc.' + product;
  document.title = t(titleKey);
  var descEl = document.querySelector('meta[name="description"]');
  if (descEl) descEl.content = t(descKey);
  var coverSub = document.getElementById('print-cover-sub');
  if (coverSub) coverSub.textContent = t('demo.printCover.' + product);
})();
