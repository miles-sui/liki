// 25 Types — ECharts rendering helpers (lazy-loaded, only on pages that use charts)
// Depends on common.js (Alpine stores, element helpers, saveOrShare, etc.)

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
    catch(e) { return LIGHT_CHART_CONFIG; }
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

  var _resizeHandlers = {};
  function initChart(el) {
    if (!el || !window.echarts) return null;
    var id = el.id;
    if (_resizeHandlers[id]) {
      var rh = _resizeHandlers[id];
      if (rh.observer) rh.observer.disconnect();
      window.removeEventListener('resize', rh.handler);
      delete _resizeHandlers[id];
    }
    var inst = echarts.getInstanceByDom(el);
    if (inst) inst.dispose();
    return echarts.init(el);
  }
  function addResize(inst, el) {
    var id = el.id;
    if (_resizeHandlers[id]) {
      var rh = _resizeHandlers[id];
      if (rh.observer) rh.observer.disconnect();
      window.removeEventListener('resize', rh.handler);
    }
    var handler = function() { inst.resize(); };
    window.addEventListener('resize', handler);
    var observer = new ResizeObserver(function() { inst.resize(); });
    observer.observe(el);
    _resizeHandlers[id] = { handler: handler, observer: observer };
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

    // Flow river chart — grouped bar chart with direction arrows.
    // months: [{label, p: [5]float, generates: int, restrains: int}, ...]
    // currentIdx: index of the current month in the months array.
    renderFlowRiver: function(el, months, currentIdx) {
      if (!window.echarts) {
        var self = this, a = arguments;
        loadECharts().then(function() { self.renderFlowRiver.apply(self, a); });
        return;
      }
      var inst = initChart(el); if (!inst) return;
      var c = cfg(), col = colors();
      var dark = Alpine.store('theme').dark;
      var genColor = dark ? '#c4a35a' : '#8a7030';
      var resColor = dark ? '#a0a0a0' : '#777';

      // Arrow markPoint helper
      function arrow(xi, yv, rot, clr) {
        return { xAxis: xi, yAxis: yv, symbol: 'arrow', symbolSize: 16, symbolRotate: rot, symbolOffset: [0, '-140%'], itemStyle: { color: clr }, label: { show: false } };
      }

      var series = [0, 1, 2, 3, 4].map(function(idx) {
        var barData = months.map(function(m) { return m.p[idx]; });
        var s = {
          name: window.ELEMENT_NAMES[window.ELEMENT_CODES[idx]] || window.ELEMENT_CODES[idx],
          type: 'bar', barMaxWidth: 28, data: barData,
          itemStyle: { color: col[idx], borderRadius: [3, 3, 0, 0] },
          emphasis: { itemStyle: { color: col[idx] } },
        };

        var mp = [];
        months.forEach(function(m, mi) {
          if (m.generates === idx) mp.push(arrow(mi, m.p[idx], 0, genColor));
          if (m.restrains === idx) mp.push(arrow(mi, m.p[idx], 180, resColor));
        });
        if (mp.length) s.markPoint = { silent: true, symbolKeepAspect: true, data: mp };

        return s;
      });

      var maxP = 0.2;
      for (var i = 0; i < months.length; i++)
        for (var j = 0; j < 5; j++)
          if (months[i].p[j] > maxP) maxP = months[i].p[j];
      var yMax = Math.max(0.28, Math.ceil(maxP / 0.8 * 20) / 20);

      var opts = {
        backgroundColor: 'transparent',
        tooltip: {
          trigger: 'axis', axisPointer: { type: 'shadow' },
          backgroundColor: c.tooltipBgColor, textStyle: { color: c.tooltipTextColor },
          valueFormatter: function(v) { return v != null ? v.toFixed(3) : '-'; },
        },
        grid: { left: '8%', right: '5%', top: 30, bottom: 40, containLabel: true },
        xAxis: {
          type: 'category', data: months.map(function(m) { return m.label; }), boundaryGap: true,
          axisLine: { lineStyle: { color: c.axisLineColor } },
          axisLabel: { color: c.axisNameColor, fontSize: 10, interval: 0, rotate: months[0] && months[0].label.length > 5 ? 30 : 0 },
        },
        yAxis: {
          type: 'value', min: 0, max: yMax,
          splitLine: { lineStyle: { color: c.splitLineColor } },
          axisLabel: { color: c.axisNameColor, formatter: function(v) { return v.toFixed(2); } },
        },
        legend: { data: series.map(function(s) { return s.name; }), bottom: 0, textStyle: { color: c.legendTextColor } },
        series: series,
      };

      if (currentIdx != null && currentIdx >= 0 && currentIdx < months.length) {
        series[0].markLine = {
          silent: true, symbol: 'none',
          lineStyle: { color: dark ? '#888' : '#bbb', type: 'dashed', width: 1.5 },
          data: [{ xAxis: currentIdx }],
        };
      }

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
      var bg = options.bgColor || '#0c0c18';
      var textColor = options.textColor || '#e8e0d5';
      var subColor = options.subColor || '#8a8078';
      var gold = dark ? '#c4a35a' : '#8a7030';
      var goldFaint = dark ? 'rgba(196,163,90,0.10)' : 'rgba(138,112,48,0.10)';
      var codes = window.ELEMENT_CODES || ['W','F','E','M','R'];
      var names = window.ELEMENT_NAMES || {};
      var pValues = options.pValues;

      // Dominant element
      var domCode = (options.identityLabel || 'W')[0];
      var domIdx = codes.indexOf(domCode);
      if (domIdx < 0) domIdx = 0;
      var domColor = elementColor(domIdx);

      // Card number: 1–25, matching types.yaml order
      var label = options.identityLabel || '';
      var TYPE_NUM = { W:1,F:2,E:3,M:4,R:5, WF:6,FE:7,EM:8,MR:9,RW:10, FW:11,EF:12,ME:13,RM:14,WR:15, WE:16,FM:17,ER:18,MW:19,RF:20, EW:21,MF:22,RE:23,WM:24,FR:25 };
      var cardNum = TYPE_NUM[label] || 0;

      // ── Radar image ──
      var radarDataURL = options.radarInst.getDataURL({
        type: 'png', pixelRatio: 2, backgroundColor: bg
      });
      var radarImg = await loadImage(radarDataURL);

      // ── Canvas ──
      var canvas = document.createElement('canvas');
      canvas.width = W; canvas.height = H;
      var ctx = canvas.getContext('2d');

      // ── 1. Background with subtle vignette ──
      var vignette = ctx.createRadialGradient(W / 2, H * 0.38, H * 0.22, W / 2, H / 2, H * 0.78);
      vignette.addColorStop(0, dark ? '#1a1a2e' : '#FFFBF5');
      vignette.addColorStop(1, bg);
      ctx.fillStyle = vignette;
      ctx.fillRect(0, 0, W, H);

      // ── 2. Header: element badge(s) + type code ──
      var badgeCY = 72, badgeR = 18, badgeCX = 48;
      var subCode = label[1]; // second element for composites
      var isComposite = subCode && subCode !== domCode;
      var subIdx = isComposite ? codes.indexOf(subCode) : -1;
      var subColor2 = subIdx >= 0 ? elementColor(subIdx) : domColor;

      // First badge
      ctx.fillStyle = domColor;
      ctx.beginPath();
      ctx.arc(badgeCX, badgeCY, badgeR, 0, 2 * Math.PI);
      ctx.fill();

      var badgeChar = (names[domCode] || domCode)[0];
      ctx.fillStyle = '#fff';
      ctx.font = 'bold 16px "Noto Serif SC", "Noto Serif", serif';
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.fillText(badgeChar, badgeCX, badgeCY);

      // Second badge (composite only)
      if (isComposite) {
        var badge2CX = badgeCX + badgeR * 2 + 6;
        ctx.fillStyle = subColor2;
        ctx.beginPath();
        ctx.arc(badge2CX, badgeCY, badgeR, 0, 2 * Math.PI);
        ctx.fill();
        var badge2Char = (names[subCode] || subCode)[0];
        ctx.fillStyle = '#fff';
        ctx.fillText(badge2Char, badge2CX, badgeCY);
      }
      ctx.textBaseline = 'alphabetic';

      // Type code (right)
      ctx.fillStyle = subColor;
      ctx.font = '13px "Noto Serif", serif';
      ctx.textAlign = 'right';
      ctx.fillText(label, W - 56, badgeCY + 5);

      // ── 3. Radar hero ──
      var radarSize = 420, radarCX = W / 2, radarCY = 340;
      var radarX = radarCX - radarSize / 2, radarY = radarCY - radarSize / 2;
      // Subtle glow behind radar
      var radarGlow = ctx.createRadialGradient(radarCX, radarCY, radarSize * 0.25, radarCX, radarCY, radarSize * 0.58);
      radarGlow.addColorStop(0, goldFaint);
      radarGlow.addColorStop(1, 'transparent');
      ctx.fillStyle = radarGlow;
      ctx.beginPath();
      ctx.arc(radarCX, radarCY, radarSize * 0.58, 0, 2 * Math.PI);
      ctx.fill();
      // Radar image
      ctx.drawImage(radarImg, radarX, radarY, radarSize, radarSize);

      // ── 4. Type label ──
      var typeLabelText = options.typeLabel || options.identityLabel || '';
      ctx.fillStyle = textColor;
      ctx.font = '40px "Crimson Pro", "Noto Serif", serif';
      ctx.textAlign = 'center';
      ctx.fillText(typeLabelText, W / 2, 602);

      // ── 5. Tagline ──
      if (options.typeName) {
        ctx.fillStyle = gold;
        ctx.font = 'italic 16px "Noto Serif", "Noto Serif SC", serif';
        var tl = options.typeName;
        if (ctx.measureText(tl).width > 520) {
          while (tl.length > 3 && ctx.measureText(tl + '…').width > 520) tl = tl.slice(0, -1);
          tl += '…';
        }
        ctx.fillText(tl, W / 2, 642);
      }

      // ── 6. Element indicators (proportional circles) ──
      if (pValues && pValues.length === 5) {
        var dotCY = 700, dotSpacing = 64, dotBaseR = 4;
        var dotStartX = W / 2 - (5 - 1) * dotSpacing / 2;
        var ringR = 16;
        for (var bi = 0; bi < 5; bi++) {
          var idx = ELEMENT_HTML_ORDER[bi];
          var dx = dotStartX + bi * dotSpacing;
          ctx.strokeStyle = goldFaint;
          ctx.lineWidth = 0.5;
          ctx.beginPath();
          ctx.arc(dx, dotCY, ringR, 0, 2 * Math.PI);
          ctx.stroke();
        }
        for (bi = 0; bi < 5; bi++) {
          idx = ELEMENT_HTML_ORDER[bi];
          var v = pValues[idx];
          var color = elementColor(idx);
          var dx = dotStartX + bi * dotSpacing;
          var r = dotBaseR + v * (ringR - dotBaseR);
          ctx.fillStyle = color;
          ctx.beginPath();
          ctx.arc(dx, dotCY, r, 0, 2 * Math.PI);
          ctx.fill();
          ctx.fillStyle = subColor;
          ctx.font = '10px "Noto Serif", serif';
          ctx.textAlign = 'center';
          ctx.fillText(codes[idx], dx, dotCY + ringR + 13);
        }
        ctx.textAlign = 'left';
      }

      // ── 7. Gold rule ──
      var ruleY = 768;
      ctx.strokeStyle = goldFaint;
      ctx.lineWidth = 0.7;
      ctx.beginPath();
      ctx.moveTo(W / 2 - 160, ruleY);
      ctx.lineTo(W / 2 + 160, ruleY);
      ctx.stroke();

      // ── 8. Footer ──
      var footY = 798;
      ctx.fillStyle = subColor;
      ctx.font = '12px "Noto Serif", serif';
      ctx.textAlign = 'center';
      ctx.fillText('25types.com  ·  No. ' + cardNum + ' of 25', W / 2, footY);

      // Locale-aware CTA
      var locale = (Alpine && Alpine.store && Alpine.store('locale')) ? Alpine.store('locale').current : 'en';
      var ctaText = locale === 'zh-CN' ? '发现你的类型  →' : 'Discover your type  →';
      ctx.fillStyle = gold;
      ctx.font = '12px "Noto Serif", "Noto Serif SC", serif';
      ctx.fillText(ctaText, W / 2, footY + 24);

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
          typeLabel: typeLabel(identity.id),
          pValues: pValues,
          bgColor: dark ? '#0c0c18' : '#FDF8F0',
          textColor: dark ? '#e8e0d5' : '#3D3226',
          subColor: dark ? '#8a8078' : '#8E8880',
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
