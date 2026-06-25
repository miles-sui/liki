// Liki i18n — path-based: /zh-Hans/ (简) /zh-Hant/ (繁) /en/
// i18next + http-backend + browser-languagedetector.
(function(){
  // FOUC guard — hidden until locale data arrives or 1.5s timeout
  var style = document.createElement('style');
  style.textContent = 'html{visibility:hidden}';
  document.head.appendChild(style);
  var foucDone = false;
  var foucTimer = setTimeout(function(){ foucDone = true; style.remove(); }, 1500);

  var LOCALES = ['zh-Hans', 'zh-Hant', 'en'];

  i18next
    .use(i18nextHttpBackend)
    .use(i18nextBrowserLanguageDetector)
    .init({
      fallbackLng: 'zh-Hant',
      keySeparator: false,
      nsSeparator: false,
      load: 'languageOnly',
      backend: { loadPath: '/i18n/{{lng}}.json' },
      detection: { order: ['path', 'navigator'], lookupFromPathIndex: 0, caches: [] }
    });

  document.documentElement.lang = i18next.language;

  var l = i18next.language || 'zh-Hant';

  // ── DOM localization ──
  function localizeDOM() {
    document.querySelectorAll('[data-i18n]').forEach(function(el){
      el.textContent = i18next.t(el.dataset.i18n);
    });
    document.querySelectorAll('[data-i18n-html]').forEach(function(el){
      el.innerHTML = i18next.t(el.dataset.i18nHtml);
    });
  }

  function finish() {
    localizeDOM();
    setMeta();
    if (!foucDone) { foucDone = true; clearTimeout(foucTimer); style.remove(); }
    document.dispatchEvent(new CustomEvent('i18n:loaded', { detail: { lang: i18next.language } }));
  }

  // Wait for both i18next initialized AND DOM ready
  i18next.on('initialized', function() {
    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', finish);
    } else {
      finish();
    }
  });

  // ── hreflang + canonical ──
  var base = location.protocol + '//' + location.host;
  var path = location.pathname.replace(/^\/(zh-Hans|zh-Hant|en)\/?/, '/');
  if (path !== '/' && !path.startsWith('/')) path = '/' + path;
  var HREFLANG = { 'zh-Hans': 'zh-Hans', 'zh-Hant': 'zh-Hant', en: 'en' };
  LOCALES.forEach(function(loc){
    var link = document.createElement('link');
    link.rel = 'alternate';
    link.hreflang = HREFLANG[loc];
    link.href = base + '/' + loc + (path === '/' ? '/' : path);
    document.head.appendChild(link);
  });
  var xd = document.createElement('link');
  xd.rel = 'alternate'; xd.hreflang = 'x-default';
  xd.href = base + '/en' + (path === '/' ? '/' : path);
  document.head.appendChild(xd);
  var canonical = document.createElement('link');
  canonical.rel = 'canonical';
  canonical.href = base + '/' + l + (path === '/' ? '/' : path);
  document.head.appendChild(canonical);

  // redirect root → /{lang}/
  if (location.pathname === '/' || location.pathname === '/index.html') {
    location.replace('/' + l + '/');
  }

  function setMeta() {
    var titleKey = document.documentElement.dataset.i18nTitle;
    if (titleKey) {
      var suffix = document.documentElement.dataset.i18nTitleSuffix || '';
      document.title = i18next.t(titleKey) + (suffix ? ' · ' + i18next.t(suffix) : '');
    }
    document.querySelectorAll('[data-i18n-content]').forEach(function(el){
      el.setAttribute('content', i18next.t(el.dataset.i18nContent));
    });
  }
})();
