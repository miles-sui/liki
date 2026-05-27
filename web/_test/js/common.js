// 25 Types — shared client JS (Alpine stores, API, helpers, ECharts loader)

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

function generateAnonToken() {
  var stored = sessionStorage.getItem('anon_token');
  var token = stored || 'anon-' + (crypto.randomUUID ? crypto.randomUUID() : Math.random().toString(36).slice(2, 10));
  try { sessionStorage.setItem('anon_token', token); } catch (_) {}
  return token;
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

// --- safe component composition (preserves getters, unlike Object.assign) ---
function composeComponent() {
  var target = {};
  for (var i = 0; i < arguments.length; i++) {
    Object.defineProperties(target, Object.getOwnPropertyDescriptors(arguments[i]));
  }
  return target;
}

// --- pick-two selection utility ---
function makePickTwo(getSelections) {
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
      else if (s.length < 2) { s.push(element); }
      else { s.shift(); s.push(element); }
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
      for (var k in this.answers) { if (this.answers[k] && this.answers[k].length === 2) n++; }
      return n;
    },

    roundStart() { return (this.round - 1) * 5; },
    roundEnd() { return Math.min(this.round * 5, this.totalQuestions) - 1; },

    answersInRound() {
      var n = 0;
      for (var i = this.roundStart(); i <= this.roundEnd(); i++) {
        var q = this.allQuestions[i];
        if (q && this.answers[q.qid] && this.answers[q.qid].length === 2) n++;
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
        if (!this.answers[q.qid] || this.answers[q.qid].length < 2) return i;
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
      }, 4000);
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

    async fetchMe() {
      if (!this.token) return;
      try {
        var res = await api('/api/users/me');
        if (res.data) {
          this.id = res.data.id;
          this.name = res.data.name;
          this.email = res.data.email || '';
          this.emailVerified = res.data.email_verified;
          this.pendingEmail = res.data.pending_email || null;
          this.isPublic = res.data.is_public;
          this.supporterSince = res.data.supporter_since || null;
        }
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
      this.id = user.id;
      this.name = user.name;
      this.email = user.email || '';
      this.isPublic = user.is_public;
      this.supporterSince = user.supporter_since || null;
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

  var res;
  try {
    res = await fetch(path, { method: opts.method, body: opts.body, headers: headers });
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
    throw new Error((data && data.error && data.error.message) || msg);
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

// --- ECharts lazy loader ---
var _echartsLoadPromise = null;
function loadECharts() {
  if (window.echarts) return Promise.resolve(window.echarts);
  if (!_echartsLoadPromise) {
    _echartsLoadPromise = new Promise(function(resolve, reject) {
      var script = document.createElement('script');
      script.src = '/js/vendor/echarts.min.js';
      script.onload = function() { resolve(window.echarts); };
      script.onerror = reject;
      document.head.appendChild(script);
    });
  }
  return _echartsLoadPromise;
}

// --- ECharts render helpers (theme-aware) ---
(function() {
  function cfg() {
    try { return Alpine.store('theme').getChartConfig(); }
    catch(e) { return lightFallback(); }
  }
  function colors() {
    try { return Alpine.store('theme').chartColors; }
    catch(e) { return fallbackColors(); }
  }
  function parseElementColors(styles) {
    return ELEMENT_CSS_PROPS.map(function(p) { return styles.getPropertyValue(p).trim(); });
  }
  function fallbackColors() {
    try {
      return parseElementColors(getComputedStyle(document.documentElement));
    } catch(e) {
      return ['#1A7A6F', '#CB3B2D', '#9A7410', '#8E8880', '#1C2238'];
    }
  }
  function lightFallback() { return LIGHT_CHART_CONFIG; }

  var _resizeHandlers = {};
  function initChart(el) {
    if (!el || !window.echarts) return null;
    var id = el.id;
    if (_resizeHandlers[id]) {
      window.removeEventListener('resize', _resizeHandlers[id]);
      delete _resizeHandlers[id];
    }
    var inst = echarts.getInstanceByDom(el);
    if (inst) inst.dispose();
    return echarts.init(el);
  }
  function addResize(inst, el) {
    var id = el.id;
    if (_resizeHandlers[id]) {
      window.removeEventListener('resize', _resizeHandlers[id]);
    }
    var handler = function() { inst.resize(); };
    _resizeHandlers[id] = handler;
    window.addEventListener('resize', handler);
  }
  function radarMax(series) {
    var mv = 0.05;
    (series || []).forEach(function(s) {
      (s.data || []).forEach(function(d) {
        (d.value || []).forEach(function(v) { mv = Math.max(mv, Math.abs(v)); });
      });
    });
    var a = Math.ceil(mv / 0.8 * 20) / 20;
    return a < 0.05 ? 0.05 : a;
  }

  function roundRect(ctx, x, y, w, h, r) {
    ctx.beginPath();
    ctx.moveTo(x + r, y);
    ctx.lineTo(x + w - r, y);
    ctx.arcTo(x + w, y, x + w, y + r, r);
    ctx.lineTo(x + w, y + h - r);
    ctx.arcTo(x + w, y + h, x + w - r, y + h, r);
    ctx.lineTo(x + r, y + h);
    ctx.arcTo(x, y + h, x, y + h - r, r);
    ctx.lineTo(x, y + r);
    ctx.arcTo(x, y, x + r, y, r);
    ctx.closePath();
  }

  function loadImage(src) {
    return new Promise(function(resolve, reject) {
      var img = new Image();
      img.onload = function() { resolve(img); };
      img.onerror = reject;
      img.src = src;
    });
  }

  window._loadECharts = loadECharts;

  window.Charts = {

    // Unified element radar — all element radar charts use this single path.
    // data.series: [{ value: [pWood,pFire,pEarth,pMetal,pWater], name, colorIdx?, lineStyle?, areaStyle? }, ...]
    // Indicators built from ELEMENT_CODES; max = max(|value|) / 0.8 so data fills ≤80% of axis.
    renderElementRadar: function(el, data, overrides) {
      if (!window.echarts) {
        var self = this, a = arguments;
        loadECharts().then(function() { self.renderElementRadar.apply(self, a); });
        return;
      }
      var inst = initChart(el); if (!inst) return;
      var c = cfg(), col = colors();

      var echartsSeries = (data.series || []).map(function(s, i) {
        var ci = (s.colorIdx != null) ? s.colorIdx : i;
        var color = col[ci % col.length];
        var extra = {};
        for (var k in s) {
          if (['value','name','colorIdx','symbol','symbolSize','lineStyle','itemStyle','areaStyle'].indexOf(k) < 0) {
            extra[k] = s[k];
          }
        }
        return Object.assign({
          type: 'radar',
          name: s.name,
          data: [{ value: s.value, name: s.name }],
          symbol: s.symbol || 'circle', symbolSize: s.symbolSize || 4,
          lineStyle: s.lineStyle || { color: color, width: 2 },
          itemStyle: s.itemStyle || { color: color },
          areaStyle: s.areaStyle || { color: color, opacity: 0.08 },
        }, extra);
      });

      var am = radarMax(echartsSeries);
      var indicators = window.ELEMENT_CODES.map(function(code, i) {
        return { name: window.ELEMENT_NAMES[code], max: am, color: col[i] };
      });

      var base = {
        backgroundColor: 'transparent',
        tooltip: { trigger: 'item', backgroundColor: c.tooltipBgColor, textStyle: { color: c.tooltipTextColor }, valueFormatter: function(v) { return v != null ? v.toFixed(2) : '-'; } },
        radar: {
          center: ['50%', '48%'], radius: '65%', indicator: indicators,
          axisName: { color: c.axisNameColor, fontSize: 11 },
          splitArea: { areaStyle: { color: c.splitAreaColor } },
          splitLine: { lineStyle: { color: c.splitLineColor } },
          axisLine: { lineStyle: { color: c.axisLineColor } },
        },
        series: echartsSeries,
      };
      if (data.legend) {
        base.legend = Object.assign({ bottom: 0, textStyle: { color: c.legendTextColor } },
          data.legend === true ? { data: echartsSeries.map(function(s) { return s.name; }) } : data.legend);
      }
      var opts = Object.assign({}, base, overrides || {});
      if (overrides && overrides.radar) opts.radar = Object.assign({}, base.radar, overrides.radar);
      inst.setOption(opts);
      addResize(inst, el);
      return inst;
    },

    // Unified element line chart — all element-based line/river/history charts use this.
    // data.series: [{ data: [values...], colorIdx?, name?, color?, ... }, ...]
    // If colorIdx is set, auto-fills name from ELEMENT_NAMES and color from theme.
    // data.categories: x-axis labels.
    renderElementLine: function(el, data, overrides) {
      if (!window.echarts) {
        var self = this, a = arguments;
        loadECharts().then(function() { self.renderElementLine.apply(self, a); });
        return;
      }
      var inst = initChart(el); if (!inst) return;
      var c = cfg(), col = colors();
      var series = (data.series || []).map(function(s, i) {
        var ci = (s.colorIdx != null) ? s.colorIdx : i;
        var color = s.color || col[ci % col.length];
        var name = s.name;
        if (!name && window.ELEMENT_CODES && window.ELEMENT_NAMES) {
          name = window.ELEMENT_NAMES[window.ELEMENT_CODES[ci]] || ('Series ' + (ci + 1));
        }
        return Object.assign({
          type: 'line', smooth: true,
          symbol: 'circle', symbolSize: 6,
          lineStyle: { color: color, width: 2 },
          itemStyle: { color: color },
          name: name || ('Series ' + (ci + 1)),
        }, s);
      });
      var opts = Object.assign({
        backgroundColor: 'transparent',
        tooltip: { trigger: 'axis', backgroundColor: c.tooltipBgColor, textStyle: { color: c.tooltipTextColor }, valueFormatter: function(v) { return v != null ? v.toFixed(2) : '-'; } },
        grid: { left: '8%', right: '5%', top: 20, bottom: 30 },
        xAxis: { type: 'category', data: data.categories, boundaryGap: false, axisLine: { lineStyle: { color: c.axisLineColor } }, axisLabel: { color: c.axisNameColor } },
        yAxis: { type: 'value', splitLine: { lineStyle: { color: c.splitLineColor } }, axisLabel: { color: c.axisNameColor, formatter: function(v) { return v.toFixed(2); } } },
        series: series,
      }, overrides || {});
      inst.setOption(opts);
      addResize(inst, el);
      return inst;
    },

    // Render a two-series bond influence radar (original shape + influenced shape).
    renderBondInfluenceChart: function(el, opts) {
      if (!el) return;
      Charts.renderElementRadar(el, {
        series: [
          { value: opts.origData, name: opts.yourLabel, lineStyle: { width: 1.5, type: 'dashed', opacity: 0.5 }, areaStyle: { opacity: 0.04 } },
          { value: opts.pData, name: opts.influencedLabel, colorIdx: opts.colorIdx != null ? opts.colorIdx : 2 },
        ],
        legend: true,
      }, opts.overrides || { radar: { center: ['50%', '50%'], radius: '60%' } });
    },

    generateShareCard: async function(options) {
      if (!options.radarInst) return null;

      // ── Constants ──
      var W = 640, H = 960;
      var dark = (Alpine && Alpine.store && Alpine.store('theme')) ? Alpine.store('theme').dark : false;
      var bg = options.bgColor || (dark ? '#1a1a2e' : '#FDF8F0');
      var textColor = options.textColor || (dark ? '#E0E0E0' : '#3D3226');
      var subColor = options.subColor || (dark ? '#9ca3af' : '#6b7280');
      var gold = dark ? '#FFC847' : '#9A7410';
      var goldSubtle = dark ? 'rgba(255,200,71,0.18)' : 'rgba(154,116,16,0.18)';
      var goldGlow = dark ? 'rgba(255,200,71,0.4)' : 'rgba(154,116,16,0.3)';
      var barTrackBg = dark ? 'rgba(255,255,255,0.06)' : 'rgba(61,50,38,0.08)';
      var codes = window.ELEMENT_CODES || ['W','F','E','M','R'];
      var names = window.ELEMENT_NAMES || {};
      var pValues = options.pValues;

      // Dominant element from identity label first character
      var domCode = (options.identityLabel || 'W')[0];
      var domIdx = codes.indexOf(domCode);
      if (domIdx < 0) domIdx = 0;
      var domColor = elementColor(domIdx);

      // Card number: 1–25, deterministic from identity
      var label = options.identityLabel || '';
      var id0 = codes.indexOf(label[0]), id1 = codes.indexOf(label[1]);
      var cardNum = (id0 >= 0 && id1 >= 0) ? id0 * 5 + id1 + 1 : 0;

      // ── Radar image ──
      var radarDataURL = options.radarInst.getDataURL({
        type: 'png', pixelRatio: 2, backgroundColor: bg
      });
      var radarImg = await loadImage(radarDataURL);

      // ── Canvas setup ──
      var canvas = document.createElement('canvas');
      canvas.width = W; canvas.height = H;
      var ctx = canvas.getContext('2d');

      // ── Helpers ──
      function lighter(hex, factor) {
        var r = parseInt(hex.slice(1,3), 16);
        var g = parseInt(hex.slice(3,5), 16);
        var b = parseInt(hex.slice(5,7), 16);
        r = Math.min(255, Math.round(r + (255 - r) * factor));
        g = Math.min(255, Math.round(g + (255 - g) * factor));
        b = Math.min(255, Math.round(b + (255 - b) * factor));
        return 'rgb(' + r + ',' + g + ',' + b + ')';
      }
      function diamond(cx, cy, size, color) {
        ctx.fillStyle = color;
        ctx.beginPath();
        ctx.moveTo(cx, cy - size/2);
        ctx.lineTo(cx + size/2, cy);
        ctx.lineTo(cx, cy + size/2);
        ctx.lineTo(cx - size/2, cy);
        ctx.closePath();
        ctx.fill();
      }

      // ── 1. Background ──
      ctx.fillStyle = bg;
      roundRect(ctx, 0, 0, W, H, 16);
      ctx.fill();

      // ── 2. Decorative frame ──
      var inset = 14;
      ctx.save();
      ctx.strokeStyle = gold;
      ctx.lineWidth = 2.5;
      ctx.shadowBlur = 8;
      ctx.shadowColor = goldGlow;
      roundRect(ctx, inset, inset, W - 2*inset, H - 2*inset, 12);
      ctx.stroke();
      ctx.restore();

      // Corner diamonds
      var cd = 30, cs = 18;
      diamond(cd, cd, cs, gold);
      diamond(W - cd, cd, cs, gold);
      diamond(cd, H - cd, cs, gold);
      diamond(W - cd, H - cd, cs, gold);

      // ── 3. Pentagon constellation (behind radar) ──
      var radarCX = 320, radarCY = 460, constellationR = 170;
      var vertices = [];
      for (var vi = 0; vi < 5; vi++) {
        var angle = (-Math.PI / 2) + vi * (2 * Math.PI / 5);
        vertices.push({ x: radarCX + constellationR * Math.cos(angle), y: radarCY + constellationR * Math.sin(angle) });
      }

      // Thin pentagon lines (generating cycle: Wood→Fire→Earth→Metal→Water)
      ctx.save();
      ctx.strokeStyle = goldSubtle;
      ctx.lineWidth = 0.6;
      ctx.beginPath();
      ctx.moveTo(vertices[0].x, vertices[0].y);
      for (vi = 1; vi < 5; vi++) { ctx.lineTo(vertices[vi].x, vertices[vi].y); }
      ctx.closePath();
      ctx.stroke();
      ctx.restore();

      // Element dots at vertices
      for (vi = 0; vi < 5; vi++) {
        var vc = elementColor(vi);
        ctx.fillStyle = vc;
        ctx.beginPath();
        ctx.arc(vertices[vi].x, vertices[vi].y, 5, 0, 2 * Math.PI);
        ctx.fill();
        // Outer glow ring
        ctx.save();
        ctx.shadowBlur = 8;
        ctx.shadowColor = dark ? vc : vc + '44';
        ctx.strokeStyle = vc;
        ctx.lineWidth = 1.2;
        ctx.beginPath();
        ctx.arc(vertices[vi].x, vertices[vi].y, 8, 0, 2 * Math.PI);
        ctx.stroke();
        ctx.restore();
      }

      // Concentric rings
      for (var ri = 0; ri < 3; ri++) {
        var ringR = constellationR - 25 + ri * 25;
        ctx.strokeStyle = dark ? 'rgba(255,255,255,0.04)' : 'rgba(154,116,16,0.08)';
        ctx.lineWidth = 0.4;
        ctx.beginPath();
        ctx.arc(radarCX, radarCY, ringR, 0, 2 * Math.PI);
        ctx.stroke();
      }

      // ── 4. Radar chart image ──
      var radarSize = 400;
      var radarX = radarCX - radarSize / 2, radarY = radarCY - radarSize / 2;
      // Subtle radial glow under radar
      var glowGrad = ctx.createRadialGradient(radarCX, radarCY, radarSize * 0.25, radarCX, radarCY, radarSize * 0.55);
      glowGrad.addColorStop(0, dark ? 'rgba(255,200,71,0.06)' : 'rgba(154,116,16,0.06)');
      glowGrad.addColorStop(1, 'transparent');
      ctx.fillStyle = glowGrad;
      ctx.beginPath();
      ctx.arc(radarCX, radarCY, radarSize * 0.55, 0, 2 * Math.PI);
      ctx.fill();

      ctx.drawImage(radarImg, radarX, radarY, radarSize, radarSize);

      // ── 5. Header: element badge + card number ──
      // Element badge (left)
      var badgeCX = 70, badgeCY = 66, badgeR = 28;
      ctx.save();
      ctx.shadowBlur = 14;
      ctx.shadowColor = dark ? domColor : domColor + '66';
      ctx.fillStyle = domColor;
      ctx.beginPath();
      ctx.arc(badgeCX, badgeCY, badgeR, 0, 2 * Math.PI);
      ctx.fill();
      ctx.restore();
      // Badge text
      var badgeChar = (names[domCode] || domCode)[0];
      ctx.fillStyle = '#fff';
      ctx.font = 'bold 24px "Noto Serif SC", "Noto Serif", serif';
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.fillText(badgeChar, badgeCX, badgeCY);
      ctx.textBaseline = 'alphabetic';

      // Card number badge (right)
      var numW = 72, numH = 28, numRX = 14;
      var numX = W - 54 - numW, numY = badgeCY - numH / 2;
      ctx.strokeStyle = gold;
      ctx.lineWidth = 1.5;
      roundRect(ctx, numX, numY, numW, numH, numRX);
      ctx.stroke();
      ctx.fillStyle = subColor;
      ctx.font = '12px "Noto Serif", serif';
      ctx.textAlign = 'center';
      ctx.fillText('No. ' + cardNum, numX + numW / 2, numY + numH / 2 + 1);
      ctx.textAlign = 'left';

      // Header divider
      var hdY = 108;
      ctx.strokeStyle = goldSubtle;
      ctx.lineWidth = 0.8;
      ctx.beginPath();
      ctx.moveTo(54, hdY); ctx.lineTo(W - 54, hdY);
      ctx.stroke();

      // ── 6. Identity display ──
      var idY = 195;
      ctx.fillStyle = textColor;
      ctx.font = 'bold 72px "Crimson Pro", "Noto Serif", serif';
      ctx.textAlign = 'center';
      ctx.save();
      ctx.shadowBlur = 4;
      ctx.shadowColor = 'rgba(0,0,0,0.12)';
      ctx.shadowOffsetY = 2;
      ctx.fillText(label, W / 2, idY);
      ctx.restore();

      if (options.typeName) {
        ctx.font = '18px "Noto Serif", "Noto Serif SC", serif';
        ctx.fillStyle = subColor;
        ctx.shadowBlur = 0;
        ctx.shadowOffsetY = 0;
        var descText = options.typeName;
        var descMaxW = 480;
        if (ctx.measureText(descText).width > descMaxW) {
          while (descText.length > 3 && ctx.measureText(descText + '…').width > descMaxW) {
            descText = descText.slice(0, -1);
          }
          descText += '…';
        }
        ctx.fillText(descText, W / 2, idY + 42);
      }

      // Identity separator: diamond between two lines
      var sepY = idY + 62;
      ctx.strokeStyle = goldSubtle;
      ctx.lineWidth = 0.7;
      ctx.beginPath();
      ctx.moveTo(220, sepY); ctx.lineTo(W / 2 - 14, sepY);
      ctx.moveTo(W / 2 + 14, sepY); ctx.lineTo(420, sepY);
      ctx.stroke();
      diamond(W / 2, sepY, 8, gold);
      ctx.textAlign = 'left';

      // ── 7. Element stat bars ──
      if (pValues && pValues.length === 5) {
        var barStartY = 704, barSpacing = 36;
        var barTrackX = 150, barTrackW = 330, barH = 14, barR = 7;
        var barPercentX = barTrackX + barTrackW + 12;
        var dotX = 56, nameX = 140;

        for (var bi = 0; bi < 5; bi++) {
          var idx = ELEMENT_HTML_ORDER[bi];
          var v = pValues[idx];
          var color = elementColor(idx);
          var barY = barStartY + bi * barSpacing;
          var name = names[codes[idx]] || codes[idx];

          // Color dot
          ctx.fillStyle = color;
          ctx.beginPath();
          ctx.arc(dotX, barY, 6, 0, 2 * Math.PI);
          ctx.fill();

          // Element name
          ctx.fillStyle = subColor;
          ctx.font = '14px "Noto Serif", "Noto Serif SC", serif';
          ctx.textAlign = 'right';
          ctx.fillText(name, nameX, barY + 5);
          ctx.textAlign = 'left';

          // Bar track
          ctx.fillStyle = barTrackBg;
          roundRect(ctx, barTrackX, barY - barH / 2, barTrackW, barH, barR);
          ctx.fill();

          // Bar fill with gradient + glow
          var fillW = Math.max(v * barTrackW, 10);
          var barGrad = ctx.createLinearGradient(barTrackX, 0, barTrackX + barTrackW, 0);
          barGrad.addColorStop(0, lighter(color, 0.45));
          barGrad.addColorStop(1, color);
          ctx.save();
          ctx.fillStyle = barGrad;
          ctx.shadowBlur = 10;
          ctx.shadowColor = dark ? color + '55' : color + '33';
          roundRect(ctx, barTrackX, barY - barH / 2, fillW, barH, barR);
          ctx.fill();
          ctx.restore();

          // Percentage
          ctx.fillStyle = color;
          ctx.font = 'bold 12px "Noto Serif", serif';
          ctx.fillText((v * 100).toFixed(0) + '%', barPercentX, barY + 5);
        }
      }

      // ── 8. Footer ──
      var footY = 898;
      // Separator line with gradient fade
      var sepGrad = ctx.createLinearGradient(80, 0, 560, 0);
      sepGrad.addColorStop(0, 'transparent');
      sepGrad.addColorStop(0.12, goldSubtle);
      sepGrad.addColorStop(0.5, goldSubtle);
      sepGrad.addColorStop(0.88, goldSubtle);
      sepGrad.addColorStop(1, 'transparent');
      ctx.strokeStyle = sepGrad;
      ctx.lineWidth = 0.7;
      ctx.beginPath();
      ctx.moveTo(80, footY - 18); ctx.lineTo(560, footY - 18);
      ctx.stroke();

      // Brand with simulated letter-spacing
      var brand = '25types.com';
      var brandFont = '600 16px "Cinzel", "Noto Serif", serif';
      ctx.font = brandFont;
      ctx.fillStyle = subColor;
      var bx = 56;
      for (var ci = 0; ci < brand.length; ci++) {
        ctx.fillText(brand[ci], bx, footY);
        bx += ctx.measureText(brand[ci]).width + 2.4;
      }

      // CTA text (right-aligned)
      var locale = (Alpine && Alpine.store && Alpine.store('locale')) ? Alpine.store('locale').current : 'en';
      var ctaText = locale === 'zh-CN' ? '发现你的类型' : 'Discover your type';
      ctx.font = '12px "Noto Serif", "Noto Serif SC", serif';
      ctx.fillStyle = dark ? 'rgba(255,255,255,0.35)' : 'rgba(61,50,38,0.45)';
      ctx.textAlign = 'right';
      ctx.fillText(ctaText, W - 56 - 38, footY);
      ctx.textAlign = 'left';

      // QR-style decorative pattern (bottom-right)
      var qrX = W - 56 - 34, qrY = footY - 30, qrS = 34, qrN = 5;
      var qrCell = qrS / qrN;
      ctx.fillStyle = dark ? 'rgba(255,255,255,0.18)' : 'rgba(61,50,38,0.2)';
      for (var qr = 0; qr < qrN; qr++) {
        for (var qc = 0; qc < qrN; qc++) {
          // Deterministic pattern from card number
          if (((cardNum + qr * 13 + qc * 7) & 3) !== 0) {
            ctx.fillRect(qrX + qc * qrCell + 1, qrY + qr * qrCell + 1, qrCell - 2, qrCell - 2);
          }
        }
      }

      return new Promise(function(resolve) {
        canvas.toBlob(function(blob) { resolve(blob); }, 'image/png');
      });
    },

    // Shared share-card button handler — deduplicates logic from result.js and profile.js.
    // Guards against concurrent calls (double-click).
    generateAndSaveShareCard: async function(radarInst, identity, pValues) {
      if (!radarInst || !identity) return;
      if (Charts._shareCardBusy) return;
      Charts._shareCardBusy = true;
      try {
        var t = Alpine.store('theme');
        var dark = t.dark;
        var blob = await Charts.generateShareCard({
          radarInst: radarInst,
          identityLabel: identity.label,
          typeName: typeDesc(identity.id),
          pValues: pValues,
          bgColor: dark ? '#1a1a2e' : '#FDF8F0',
          textColor: dark ? '#E0E0E0' : '#3D3226',
          subColor: dark ? '#9ca3af' : '#6b7280',
        });
        if (blob) saveOrShare(blob, '25types-' + identity.label + '.png');
      } catch (e) {
        console.error(e);
        try { Alpine.store('toast').error(Alpine.store('locale').t('toast.shareCardFailed')); } catch (_) {}
      } finally {
        Charts._shareCardBusy = false;
      }
    },
  };
  Charts._shareCardBusy = false;
})();
