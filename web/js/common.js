// 25 Types — shared client JS (Alpine stores, API, helpers, ECharts loader)

// API host routing: mingli calls go to the stateless mingli-server,
// app calls (auth, users, payments, assessments, profiles, flow) go to app-server.
// In dev both run on the same origin so relative URLs work for both.
var API_MINGLI = window.API_MINGLI || '';
var API_APP  = window.API_APP  || '';

var MINGLI_PATH_PREFIXES = ['/api/mingli/bazi/', '/api/mingli/huangli/', '/api/mingli/fengshui/', '/api/qiming/', '/api/reference/', '/api/solar-terms', '/api/location'];

function apiHost(path) {
  for (var i = 0; i < MINGLI_PATH_PREFIXES.length; i++) {
    if (path.indexOf(MINGLI_PATH_PREFIXES[i]) === 0) return API_MINGLI;
  }
  return API_APP;
}

// --- a11y & locale-link patching ---
document.addEventListener('DOMContentLoaded', () => {
  document.querySelectorAll('.navbar button.lg\\:hidden').forEach(btn => {
    if (!btn.hasAttribute('aria-label')) btn.setAttribute('aria-label', 'Toggle menu');
  });
  document.querySelectorAll('.loading-spinner').forEach(el => {
    if (!el.hasAttribute('role')) el.setAttribute('role', 'status');
  });

  var loc = window.CURRENT_LOCALE || 'en';
  document.querySelectorAll('a[href^="/"]').forEach(function(a) {
    var href = a.getAttribute('href');
    if (/^\/(api|js|css|fonts|img|healthz)\//.test(href)) return;
    if (/^\/(en|zh-CN)\//.test(href)) return;
    if (/\.[a-z]{2,5}$/i.test(href)) return;
    a.setAttribute('href', '/' + loc + href);
  });
});

// --- static page content ---
function staticPage(key) {
  return (window.PAGE_CONTENT && window.PAGE_CONTENT[key]) || '';
}

// --- shared utilities ---
function formatDate(s, locale) {
  if (!s) return '';
  var d = new Date(s);
  return d.toLocaleString(locale === 'zh-CN' ? 'zh-CN' : 'en-US', { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
}

// Normalize loose time input to HH:MM. Accepts:
//   "0030" → "00:30", "930" → "09:30", "12" → "12:00", "0:30" → "00:30"
function normalizeTime(raw) {
  if (!raw || !raw.trim()) return raw;
  var s = raw.trim();
  // Already canonical HH:MM
  if (/^([01][0-9]|2[0-3]):[0-5][0-9]$/.test(s)) return s;
  // Handle H:MM or HH:MM with loose validation
  var cm = /^(\d{1,2}):(\d{2})$/.exec(s);
  if (cm) {
    var h = parseInt(cm[1], 10), m = parseInt(cm[2], 10);
    if (h > 23) h = 23;
    if (m > 59) m = 59;
    return ('0' + h).slice(-2) + ':' + ('0' + m).slice(-2);
  }
  // Digits only
  var d = s.replace(/\D/g, '');
  if (!d) return raw;
  if (d.length <= 2) {
    var hh = parseInt(d, 10);
    if (hh > 23) hh = 23;
    return ('0' + hh).slice(-2) + ':00';
  }
  if (d.length === 3) d = '0' + d;
  // Take first 4 digits
  var hh = parseInt(d.substring(0, 2), 10);
  var mm = parseInt(d.substring(2, 4), 10);
  if (hh > 23) hh = 23;
  if (mm > 59) mm = 59;
  return ('0' + hh).slice(-2) + ':' + ('0' + mm).slice(-2);
}

function generateAnonToken() {
  var stored = sessionStorage.getItem('anon_token');
  var token = stored || 'anon-' + (crypto.randomUUID ? crypto.randomUUID() : Math.random().toString(36).slice(2, 10));
  try { sessionStorage.setItem('anon_token', token); } catch (_) {}
  return token;
}

function passwordStrength(pw) {
  if (!pw) return { score: 0 };
  var s = 0;
  if (pw.length >= 8) s++;
  if (pw.length >= 12) s++;
  if (/[a-z]/.test(pw) && /[A-Z]/.test(pw)) s++;
  if (/\d/.test(pw)) s++;
  if (/[^a-zA-Z0-9]/.test(pw)) s++;
  return { score: Math.min(s, 5) };
}

function pwBarClass(score, n) {
  if (score < n) return 'bg-base-300';
  if (n <= 2) return score <= 2 ? 'bg-error' : score === 3 ? 'bg-warning' : 'bg-success';
  if (n === 3) return 'bg-warning';
  return 'bg-success';
}

// --- element helpers ---
// Visual order: 火→木→土→金→水, top to bottom
// Natural indices into ELEMENT_CODES / chartColors / namedToArray result
// ELEMENT_DISPLAY_ORDER: reversed for ECharts (legends render bottom-up)
// ELEMENT_HTML_ORDER:   direct order for HTML top-to-bottom rendering
var ELEMENT_DISPLAY_ORDER = [4, 3, 2, 0, 1];
var ELEMENT_HTML_ORDER = [1, 0, 2, 3, 4];
var ELEMENT_CSS_PROPS = ['--wood', '--fire', '--earth', '--metal', '--water'];
var ERR_TOKEN_EXPIRED = 'token_expired';

function elementName(idx) {
  return window.ELEMENT_NAMES[window.ELEMENT_CODES[idx]] || window.ELEMENT_CODES[idx];
}
function elementColor(arg) {
  var store = Alpine.store('theme');
  if (typeof arg === 'number') return store.chartColors[arg] || '#888';
  return store.colors[arg] || '#888';
}

function namedToArray(v) {
  if (!v) return [];
  if (Array.isArray(v)) return v;
  var keys = ['wood', 'fire', 'earth', 'metal', 'water'];
  return keys.map(function(k) { return v[k] || 0; });
}

// Derive a simple identity { label, id, category } from d values via top-2 elements.
function identityFromD(d) {
  var arr = namedToArray(d);
  var codes = window.ELEMENT_CODES || ['W', 'F', 'E', 'M', 'R'];
  var names = window.ELEMENT_NAMES || {};
  var indexed = arr.map(function(v, i) { return { v: v, i: i }; });
  indexed.sort(function(a, b) { return b.v - a.v; });
  var top = indexed[0].i, second = indexed[1].i;
  var id = codes[top] + codes[second];
  return { label: id, id: id, category: names[codes[top]] || codes[top] };
}

function typeDesc(id) {
  return (window.TYPE_DESCS && window.TYPE_DESCS[id]) || '';
}
function typeLabel(id) {
  return (window.TYPE_LABELS && window.TYPE_LABELS[id]) || '';
}

// d → p: p[i] = 0.2 + d[i]/5, matches engine.Deviation.ToProportion()
function deviationToProportion(d) {
  var arr = Array.isArray(d) ? d : namedToArray(d);
  if (!arr.length) return [];
  return arr.map(function(v) { return 0.2 + v / 5; });
}

// computeBondShapes(bond) → {origSelf, origOther, pSelf, pOther, selfDeltas, otherDeltas} or null
// bond is {self, other, delta_a, delta_b} where each is a named object {wood, fire, earth, metal, water}.
function computeBondShapes(bond) {
  if (!bond) return null;
  var dEffSelf = namedToArray(bond.self);
  var dEffOther = namedToArray(bond.other);
  var arrA = namedToArray(bond.delta_a);
  var arrB = namedToArray(bond.delta_b);
  var origSelf = deviationToProportion(dEffSelf.map(function(v, i) { return v - arrA[i]; }));
  var origOther = deviationToProportion(dEffOther.map(function(v, i) { return v - arrB[i]; }));
  var pSelf = deviationToProportion(dEffSelf);
  var pOther = deviationToProportion(dEffOther);
  return {
    origSelf: origSelf, origOther: origOther,
    pSelf: pSelf, pOther: pOther,
    selfDeltas: ELEMENT_HTML_ORDER.map(function(i) { return { idx: i, delta: pSelf[i] - origSelf[i] }; }),
    otherDeltas: ELEMENT_HTML_ORDER.map(function(i) { return { idx: i, delta: pOther[i] - origOther[i] }; }),
  };
}

// --- concord display helper ---
// Maps raw concord value (API/internal: 顺/逆/平) to locale-aware display props.
function concordProps(raw) {
  if (!raw) return { label: '', badgeClass: '', desc: '' };
  var map = {
    '顺': { labelKey: 'bond.concordShun', descKey: 'bond.concordShunDesc', cls: 'badge-success' },
    '逆': { labelKey: 'bond.concordNi', descKey: 'bond.concordNiDesc', cls: 'badge-error' },
    '平': { labelKey: 'bond.concordPing', descKey: 'bond.concordPingDesc', cls: 'badge-ghost' },
  };
  var m = map[raw];
  if (!m) return { label: raw, badgeClass: '', desc: '' };
  var t = function(k) {
    try { return Alpine.store('locale').t(k); } catch(e) { return k; }
  };
  return { label: t(m.labelKey), badgeClass: m.cls, desc: t(m.descKey) };
}

// --- safe component composition (preserves getters, unlike Object.assign) ---
function composeComponent() {
  var target = {};
  for (var i = 0; i < arguments.length; i++) {
    Object.defineProperties(target, Object.getOwnPropertyDescriptors(arguments[i]));
  }
  return target;
}

// --- free-choice selection utility ---
function makePickAny(getSelections) {
  return {
    isSelected(qid, element) {
      var s = getSelections.call(this)[qid];
      return s && s.includes(element);
    },
    toggleSelect(qid, element) {
      var sel = getSelections.call(this);
      var s = sel[qid];
      if (!s) s = sel[qid] = [];
      var idx = s.indexOf(element);
      if (idx >= 0) { s.splice(idx, 1); }
      else { s.push(element); }
      sel[qid] = s;
    },
  };
}

// --- shared assessment round navigation (used by assess.js & match-landing.js) ---
function assessmentNavigation() {
  return {
    get round() { return Math.floor(this.currentQIndex / 5) + 1; },
    get totalRounds() { return Math.ceil(this.allQuestions.length / 5) || 6; },
    get totalQuestions() { return this.allQuestions.length || 30; },
    get currentQuestion() { return this.allQuestions[this.currentQIndex]; },

    get answeredCount() {
      var n = 0;
      for (var k in this.answers) { if (this.answers[k]) n += this.answers[k].length; }
      return n;
    },

    roundStart() { return (this.round - 1) * 5; },
    roundEnd() { return Math.min(this.round * 5, this.totalQuestions) - 1; },

    answersInRound() {
      var n = 0;
      for (var i = this.roundStart(); i <= this.roundEnd(); i++) {
        var q = this.allQuestions[i];
        if (q && this.answers[q.qid]) n += this.answers[q.qid].length;
      }
      return n;
    },

    isRoundComplete() {
      return this.answersInRound() >= Math.min(5, this.totalQuestions - this.roundStart());
    },

    canGoPrev() { return this.currentQIndex > 0; },
    canGoNext() { return this.currentQIndex < this.totalQuestions - 1; },
    goPrev() { if (this.canGoPrev()) this.currentQIndex--; },
    goNext() { if (this.canGoNext()) this.currentQIndex++; },

    get firstUnanswered() {
      for (var i = 0; i < this.allQuestions.length; i++) {
        var q = this.allQuestions[i];
        if (!this.answers[q.qid] || this.answers[q.qid].length === 0) return i;
      }
      return -1;
    },

    jumpToUnanswered() {
      var idx = this.firstUnanswered;
      if (idx >= 0) this.currentQIndex = idx;
    },
  };
}

async function loadAssessmentQuestions(locale) {
  try {
    var resp = await api('/api/assessments/questions?locale=' + locale);
    var data = (resp && resp.data) ? resp.data : resp;
    var flat = [];
    if (data && data.rounds) {
      for (var ri = 0; ri < data.rounds.length; ri++) {
        for (var qi = 0; qi < data.rounds[ri].questions.length; qi++) {
          flat.push(data.rounds[ri].questions[qi]);
        }
      }
    }
    return flat;
  } catch (e) {
    console.error(e);
    return [];
  }
}

function localePath(path) {
  var loc = window.CURRENT_LOCALE || 'en';
  if (typeof Alpine !== 'undefined') {
    try { loc = Alpine.store('locale').current; } catch(e) {}
  }
  return '/' + loc + path;
}

// --- Alpine stores ---
var LIGHT_CHART_CONFIG = {
  splitAreaColor: ['rgba(0,0,0,0.02)','rgba(0,0,0,0.02)'],
  splitLineColor: 'rgba(0,0,0,0.08)',
  axisLineColor: 'rgba(0,0,0,0.10)',
  axisNameColor: '#777',
  tooltipBgColor: 'rgba(255,255,255,0.95)', tooltipTextColor: '#3D3226',
  legendTextColor: '#777',
};
document.addEventListener('alpine:init', () => {
  Alpine.store('locale', {
    current: window.CURRENT_LOCALE || 'en',

    init() {
      this.current = window.CURRENT_LOCALE || 'en';
      document.documentElement.lang = this.current;
    },

    t(key) {
      var val = window.TRANSLATIONS[key];
      return val || key;
    },
  });

  Alpine.store('theme', {
    current: 'wuxing',
    dark: false,
    colors: {},
    chartColors: [],

    init() {
      this.syncColors();
      var saved = localStorage.getItem('theme');
      if (saved === 'wuxing-dark' || saved === 'wuxing') {
        this.setTheme(saved);
      } else {
        this.apply();
      }
    },

    syncColors() {
      var styles = getComputedStyle(document.documentElement);
      var codes = window.ELEMENT_CODES || ['W','F','E','M','R'];
      this.colors = {};
      this.chartColors = [];
      for (var i = 0; i < codes.length; i++) {
        var hex = styles.getPropertyValue(ELEMENT_CSS_PROPS[i]).trim();
        this.colors[codes[i]] = hex;
        this.chartColors.push(hex);
      }
    },

    setTheme(name) {
      this.current = name;
      this.dark = (name === 'wuxing-dark');
      this.apply();
      this.syncColors();
    },

    toggle() {
      this.setTheme(this.dark ? 'wuxing' : 'wuxing-dark');
    },

    apply() {
      document.documentElement.setAttribute('data-theme', this.current);
      localStorage.setItem('theme', this.current);
      window.dispatchEvent(new CustomEvent('theme-changed', { detail: { theme: this.current, dark: this.dark } }));
    },

    getChartConfig() {
      if (this.dark) {
        return {
          splitAreaColor: ['rgba(255,255,255,0.02)', 'rgba(255,255,255,0.02)'],
          splitLineColor: 'rgba(255,255,255,0.08)',
          axisLineColor: 'rgba(255,255,255,0.08)',
          axisNameColor: '#aaa',
          tooltipBgColor: 'rgba(26,26,46,0.92)',
          tooltipTextColor: '#ddd',
          legendTextColor: '#888',
        };
      }
      return Object.assign({}, LIGHT_CHART_CONFIG);
    },
  });

  Alpine.store('ui', {
    mobileOpen: false,
  });

  Alpine.store('toast', {
    messages: [],

    show(msg, type) {
      type = type || 'info';
      var id = Date.now();
      this.messages.push({ id: id, msg: msg, type: type });
      setTimeout(() => {
        this.messages = this.messages.filter(function(m) { return m.id !== id; });
      }, 8000);
    },

    success(msg) { this.show(msg, 'success'); },
    error(msg) { this.show(msg, 'error'); },
  });

  Alpine.store('auth', {
    id: null,
    name: '',
    token: localStorage.getItem('token') || '',
    email: '',
    emailVerified: false,
    pendingEmail: null,
    showVerifyBanner: false,
    isPublic: false,
    supporterSince: null,

    init() {
      var t = localStorage.getItem('token');
      if (t) {
        this.token = t;
        this.fetchMe();
      }
    },

    get isSupporter() {
      return !!this.supporterSince;
    },

    _hydrate(data) {
      this.id = data.id;
      this.name = data.name;
      this.email = data.email || '';
      this.emailVerified = data.email_verified || false;
      this.pendingEmail = data.pending_email || null;
      this.isPublic = data.is_public || false;
      this.supporterSince = data.supporter_since || null;
      this.showVerifyBanner = !!(this.token && this.id && !this.emailVerified);
    },

    async fetchMe() {
      if (!this.token) return;
      try {
        var res = await api('/api/users/me');
        if (res.data) this._hydrate(res.data);
      } catch (e) {
        console.error(e);
      }
    },

    logout() {
      if (this.token) {
        api('/api/auth/logout', { method: 'POST' }).catch(function(e) { console.error(e); });
      }
      localStorage.removeItem('token');
      this.id = null;
      this.name = '';
      this.token = '';
      this.email = '';
      this.emailVerified = false;
      this.pendingEmail = null;
      window.location = localePath('/login');
    },

    login(token, user) {
      localStorage.setItem('token', token);
      this.token = token;
      this._hydrate(user);
    },
  });

  // Location store — city-level geolocation, like e-commerce region selector.
  // City data: { name, name_zh, country, lat, lng, detected }
  Alpine.store('location', {
    city: null,
    detected: false,
    searchOpen: false,
    searchQuery: '',
    searchResults: [],
    searchLoading: false,
    _searchTimer: null,

    popularCities: [
      { name: 'Beijing', name_zh: '北京', country: 'China', lat: 39.9042, lng: 116.4074 },
      { name: 'Shanghai', name_zh: '上海', country: 'China', lat: 31.2304, lng: 121.4737 },
      { name: 'Shenzhen', name_zh: '深圳', country: 'China', lat: 22.5431, lng: 114.0579 },
      { name: 'Guangzhou', name_zh: '广州', country: 'China', lat: 23.1292, lng: 113.2644 },
      { name: 'Chengdu', name_zh: '成都', country: 'China', lat: 30.5728, lng: 104.0668 },
      { name: 'Changsha', name_zh: '长沙', country: 'China', lat: 28.2282, lng: 112.9388 },
      { name: 'Hangzhou', name_zh: '杭州', country: 'China', lat: 30.2741, lng: 120.1551 },
      { name: 'Wuhan', name_zh: '武汉', country: 'China', lat: 30.5928, lng: 114.3055 },
      { name: 'Chongqing', name_zh: '重庆', country: 'China', lat: 29.4316, lng: 106.9123 },
      { name: 'Nanjing', name_zh: '南京', country: 'China', lat: 32.0603, lng: 118.7969 },
      { name: 'London', name_zh: '伦敦', country: 'United Kingdom', lat: 51.5074, lng: -0.1278 },
      { name: 'New York', name_zh: '纽约', country: 'United States', lat: 40.7128, lng: -74.006 },
      { name: 'Tokyo', name_zh: '东京', country: 'Japan', lat: 35.6762, lng: 139.6503 },
      { name: 'Singapore', name_zh: '新加坡', country: 'Singapore', lat: 1.3521, lng: 103.8198 },
      { name: 'Sydney', name_zh: '悉尼', country: 'Australia', lat: -33.8688, lng: 151.2093 },
      { name: 'Hong Kong', name_zh: '香港', country: 'China', lat: 22.3193, lng: 114.1694 },
      { name: 'Taipei', name_zh: '台北', country: 'Taiwan', lat: 25.033, lng: 121.5654 },
    ],

    get hasCity() { return !!this.city; },
    get displayName() {
      if (!this.city) return '';
      var loc = window.CURRENT_LOCALE || 'en';
      return loc === 'zh-CN' ? (this.city.name_zh || this.city.name) : this.city.name;
    },

    init() {
      var saved = localStorage.getItem('location_city');
      if (saved) {
        try { this.city = JSON.parse(saved); this.detected = false; } catch (e) {}
      }
      if (!this.city) {
        // Only attempt IP detection once per session — prevents redundant /api/location calls on every navigation.
        if (!sessionStorage.getItem('loc_detected')) {
          this._detectFromIP();
        }
      }
    },

    async _detectFromIP() {
      try {
        sessionStorage.setItem('loc_detected', '1');
        var res = await api('/api/location', { quiet: true });
        if (res && res.data && res.data.city) {
          var enName = res.data.city;
          var zhName = '';
          for (var i = 0; i < this.popularCities.length; i++) {
            if (this.popularCities[i].name.toLowerCase() === enName.toLowerCase()) {
              zhName = this.popularCities[i].name_zh;
              break;
            }
          }
          this.city = {
            name: enName,
            name_zh: zhName,
            country: res.data.country || '',
            lat: res.data.lat,
            lng: res.data.lng,
          };
          this.detected = true;
          localStorage.setItem('location_city', JSON.stringify(this.city));

          // If the backend fell back to Beijing (private IP in dev), try browser geolocation.
          if (enName === 'Beijing' && navigator.geolocation) {
            this._detectFromBrowser();
          }
        }
      } catch (e) { /* silently ignore */ }
    },

    _detectFromBrowser() {
      var self = this;
      navigator.geolocation.getCurrentPosition(function(pos) {
        var closest = null;
        var minDist = Infinity;
        for (var i = 0; i < self.popularCities.length; i++) {
          var c = self.popularCities[i];
          var d = Math.hypot(c.lat - pos.coords.latitude, c.lng - pos.coords.longitude);
          if (d < minDist) { minDist = d; closest = c; }
        }
        if (closest && minDist < 3) { // within ~3° — generous city-level match for sparse popular city coverage
          self.city = {
            name: closest.name,
            name_zh: closest.name_zh,
            country: closest.country,
            lat: closest.lat,
            lng: closest.lng,
          };
          self.detected = true;
          localStorage.setItem('location_city', JSON.stringify(self.city));
        }
      }, function() { /* user denied — keep Beijing fallback */ },
         { enableHighAccuracy: false, timeout: 5000, maximumAge: 600000 });
    },

    selectCity(c) {
      this.city = { name: c.name, name_zh: c.name_zh || '', country: c.country || '', lat: c.lat, lng: c.lng };
      this.detected = false;
      this.searchOpen = false;
      this.searchQuery = '';
      this.searchResults = [];
      localStorage.setItem('location_city', JSON.stringify(this.city));
    },

    onSearchInput() {
      var self = this;
      clearTimeout(this._searchTimer);
      var q = this.searchQuery.trim();
      if (q.length < 1) { this.searchResults = []; return; }
      this.searchLoading = true;
      this._searchTimer = setTimeout(function() {
        api('/api/mingli/bazi/cities?q=' + encodeURIComponent(q), { quiet: true }).then(function(res) {
          self.searchResults = (res && res.data) || [];
          self.searchLoading = false;
        }).catch(function() { self.searchLoading = false; });
      }, 200);
    },
  });

});

// --- Shared clipboard helper ---
async function copyText(text, opts) {
  opts = opts || {};
  try {
    await navigator.clipboard.writeText(text);
    if (!opts.quiet) safeToast(safeT(opts.successKey || 'copied'), 'success');
  } catch (e) {
    console.error(e);
    if (!opts.quiet) safeToast(safeT(opts.errorKey || 'toast.copyFailed'), 'error');
  }
}

// --- Share card helpers ---
function saveOrShare(blob, filename) {
  if (navigator.share && navigator.canShare) {
    var file = new File([blob], filename, { type: 'image/png' });
    if (navigator.canShare({ files: [file] })) {
      return navigator.share({ files: [file], title: filename });
    }
  }
  var url = URL.createObjectURL(blob);
  var a = document.createElement('a');
  a.href = url; a.download = filename;
  a.click();
  setTimeout(function() { URL.revokeObjectURL(url); }, 100);
}

// --- Locale switcher ---
function switchLocale(locale) {
  localStorage.setItem('locale', locale);
  var path = window.location.pathname.replace(/\/+$/, '');
  var parts = path.split('/');
  if (parts[1] === 'en' || parts[1] === 'zh-CN') {
    parts[1] = locale;
  }
  window.location = parts.join('/') + window.location.search + window.location.hash;
}

// --- Navbar user menu (available on all pages) ---
function userMenu() {
  return {
    open: false,

    get auth() { return Alpine.store('auth'); },
    get locale() { return Alpine.store('locale'); },
  };
}

// --- Safe helpers (Alpine may not be booted yet) ---
function safeT(key) {
  try { return Alpine.store('locale').t(key); } catch (_) { return key; }
}
function safeToast(msg, type) {
  try { Alpine.store('toast').show(msg, type); } catch (_) {}
}

// --- API client ---
async function handle401(path, res) {
  var body = {};
  try { body = await res.json(); } catch (e) { console.error(e); }
  if (body.error && body.error.code === ERR_TOKEN_EXPIRED) {
    safeToast(safeT('error.sessionExpired'), 'error');
  }
  localStorage.removeItem('token');
  try { Alpine.store('auth').id = null; } catch (_) {}
  try { Alpine.store('auth').token = ''; } catch (_) {}
  if (path !== '/api/auth/login' && path !== '/api/auth/register') {
    window.location = localePath('/login') + '?redirect=' + encodeURIComponent(location.pathname);
  }
  throw new Error(body.error ? body.error.message : safeT('error.unauthorized'));
}

async function api(path, opts) {
  opts = opts || {};
  var quiet = opts.quiet;
  var token = localStorage.getItem('token');
  var headers = { 'Content-Type': 'application/json' };
  if (opts.headers) Object.assign(headers, opts.headers);
  if (token) headers['Authorization'] = 'Bearer ' + token;
  try { headers['X-Locale'] = Alpine.store('locale').current; } catch (_) {}

  var url = apiHost(path) + path;

  var res;
  try {
    res = await fetch(url, { method: opts.method, body: opts.body, headers: headers });
  } catch (e) {
    console.error(e);
    if (!quiet) safeToast(safeT('error.network'), 'error');
    throw new Error(safeT('error.network'));
  }

  if (res.status === 401) await handle401(path, res);

  var data = {};
  try { data = await res.json(); } catch (e) { console.error(e); }

  if (!res.ok) {
    var msg = (data && data.error && data.error.message) || safeT('error.serverError');
    if (!quiet) safeToast(msg, 'error');
    var err = new Error(msg);
    err.code = (data && data.error && data.error.code) || null;
    throw err;
  }
  return data;
}

// --- frontend error collection ---
window.addEventListener('error', function(e) {
  if (!e.filename) return;
  var payload = {
    message: e.message,
    filename: e.filename.replace(location.origin, ''),
    lineno: e.lineno,
    colno: e.colno,
    stack: (e.error && e.error.stack) ? e.error.stack.slice(0, 1000) : '',
    url: location.pathname,
  };
  navigator.sendBeacon('/api/errors/frontend', JSON.stringify(payload));
});

// ECharts rendering code moved to /js/charts.js — only load on pages that use charts.
