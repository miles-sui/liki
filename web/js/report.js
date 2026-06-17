import { html, render } from './lit-html.js';

const t = (key) => i18next.t(key);

const svgClock = html`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>`;
const svgX = html`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>`;

function renderIcon(def) {
  const icon = def.icon;
  if (icon === 'spinner-lg') return html`<div class="spinner mx-auto mb-4" style="width:2.5rem;height:2.5rem;border-width:3px;"></div>`;
  if (!icon) return '';
  return html`<div class="status-icon">
    ${icon === 'spinner' ? html`<span class="spin"></span>` :
      icon === 'clock' ? svgClock :
      icon === 'x' ? svgX : ''}
  </div>`;
}

const PHASE_CARD = {
  loading: { icon: 'spinner-lg' },
  error: {
    variant: 'error', icon: 'x',
    titleFn(app) { return app.error; },
    actions: [{ labelKey: 'site.backHome', url: '/', cls: 'btn btn-primary' }]
  },
  payment: {
    variant: 'payment', icon: 'spinner',
    titleKey: 'report.polling', hintKey: 'report.pollingHint',
    actions: [{ labelKey: 'report.pay', click: 'goPay', loading: 'payLoading' }]
  },
  generating: {
    variant: 'generating', icon: 'spinner',
    titleKey: 'report.generating', hintKey: 'report.generatingHint'
  },
  timeout_payment: {
    variant: 'timeout', icon: 'clock',
    titleKey: 'report.timeout', hintKey: 'report.timeoutHint',
    actions: [
      { labelKey: 'report.pay', click: 'goPay', loading: 'payLoading' },
      { labelKey: 'report.checkStatus', click: 'checkPaymentStatus', cls: 'btn btn-sm bg-white border border-stone-200 text-stone-600 rounded' },
      { labelKey: 'report.refresh', click: 'retryPoll', cls: 'btn btn-sm bg-white border border-stone-200 text-stone-600 rounded' },
      { labelKey: 'report.copyLink', click: 'copyLink', cls: 'btn btn-sm bg-white border border-stone-200 text-stone-600 rounded' }
    ]
  },
  timeout_generating: {
    variant: 'timeout', icon: 'clock',
    titleKey: 'report.generating', hintKey: 'report.generatingHint',
    actions: [
      { labelKey: 'report.refresh', click: 'retryPoll', cls: 'btn btn-primary' },
      { labelKey: 'report.copyLink', click: 'copyLink', cls: 'btn btn-sm bg-white border border-stone-200 text-stone-600 rounded' }
    ]
  }
};

class ReportApp {
  constructor() {
    this.phase = 'loading';
    this.error = '';
    this.payLoading = false;
    this.bannerClosed = sessionStorage.getItem('report-banner-closed') === '1';
    this.copyBtnText = t('report.copyLink');
    this.product = '';
    this.chart = {};
    this.bond = { gan_rel: [], zhi_rel: [], key_hints: [] };
    this.namingData = { analysis: {}, candidates: [] };
    this.pillars = [];
    this.llmHTML = '';
    this.elapsed = 0;
    this.elapsedTimer = null;
    this.pollProgress = 0;
    this.pollMax = 100;

    this.dom = {
      banner: document.getElementById('report-banner'),
      bannerCopyBtn: document.getElementById('banner-copy-btn'),
      bannerCopyText: document.getElementById('banner-copy-text'),
      headerTitle: document.getElementById('report-header-title'),
      transition: document.getElementById('transition-overlay'),
      transitionText: document.getElementById('transition-text'),
      statusArea: document.getElementById('status-area'),
      reportContent: document.getElementById('report-content'),
      printBrand: document.getElementById('print-brand'),
      chartSection: document.getElementById('chart-section'),
      bondSection: document.getElementById('bond-section'),
      namingSection: document.getElementById('naming-section'),
      chartSummary: document.getElementById('chart-summary'),
      chartTableBody: document.getElementById('chart-table-body'),
      chartWuxing: document.getElementById('chart-wuxing'),
      chartTiaohou: document.getElementById('chart-tiaohou'),
      chartDayun: document.getElementById('chart-dayun'),
      chartInterpretation: document.getElementById('chart-interpretation'),
      bondDayMasters: document.getElementById('bond-day-masters'),
      bondGanRel: document.getElementById('bond-gan-rel'),
      bondZhiRel: document.getElementById('bond-zhi-rel'),
      bondKeyHints: document.getElementById('bond-key-hints'),
      bondInterpretation: document.getElementById('bond-interpretation'),
      namingSummary: document.getElementById('naming-summary'),
      namingCandidates: document.getElementById('naming-candidates'),
      namingInterpretation: document.getElementById('naming-interpretation')
    };

    this.bindEvents();
    if (i18next.isInitialized) this.render();
    else document.addEventListener('i18n:loaded', () => this.render(), { once: true });
    this.init();
  }

  bindEvents() {
    if (this.dom.bannerCopyBtn) this.dom.bannerCopyBtn.addEventListener('click', () => this.copyLink());
    const bannerCloseBtn = document.getElementById('banner-close-btn');
    if (bannerCloseBtn) bannerCloseBtn.addEventListener('click', () => this.closeBanner());
    const btnPrint = document.getElementById('btn-print');
    if (btnPrint) btnPrint.addEventListener('click', () => window.print());
    const btnShare = document.getElementById('btn-share');
    if (btnShare) btnShare.addEventListener('click', () => this.share());
  }

  progressPercent() {
    return this.pollMax > 0 ? Math.round((this.pollProgress / this.pollMax) * 100) : 0;
  }

  elapsedText() {
    const m = Math.floor(this.elapsed / 60);
    const s = this.elapsed % 60;
    return m > 0 ? m + 'm ' + s + 's' : s + 's';
  }

  // ── render ──

  render() {
    if (this.dom.banner) this.dom.banner.style.display = this.bannerClosed ? 'none' : '';
    if (this.dom.headerTitle) this.dom.headerTitle.textContent = t('report.title.' + (this.product || 'default'));
    if (this.dom.transition) this.dom.transition.style.display = this.phase === 'unlocking' ? '' : 'none';

    this.renderStatusCard();

    if (this.dom.reportContent) this.dom.reportContent.style.display = this.phase === 'ready' ? '' : 'none';
    if (this.phase === 'ready') this.renderProductSections();
  }

  renderStatusCard() {
    const el = this.dom.statusArea;
    if (!el) return;
    const def = PHASE_CARD[this.phase];
    if (!def) { render('', el); return; }

    const variant = def.variant || '';
    const showProgress = this.phase === 'payment' || this.phase === 'generating';

    render(html`
      <div class="status-card${variant ? ' status-' + variant : ''}">
        ${renderIcon(def)}

        <p class="status-title${!def.variant ? ' text-stone-600' : ''}">
          ${def.titleKey ? t(def.titleKey) : (def.titleFn ? def.titleFn(this) : '')}
        </p>

        ${this.elapsed > 0 ? html`<p class="status-elapsed" style="font-size:.8125rem;color:var(--stone-500);margin-bottom:.25rem;">${this.elapsedText()}</p>` : ''}

        ${showProgress ? html`
          <div class="progress-wrap" style="width:100%;max-width:280px;margin:0 auto .5rem;">
            <progress id="poll-progress" value="${this.pollProgress}" max="${this.pollMax}" style="width:100%;height:6px;border-radius:3px;appearance:none;"></progress>
            <span id="poll-pct" class="text-xs text-stone-500">${this.progressPercent()}%</span>
          </div>` : ''}

        ${def.hintKey ? html`<p class="status-hint">${t(def.hintKey)}</p>` : ''}

        ${def.actions ? html`
          <div class="status-actions">
            ${def.actions.map(a => a.url
              ? html`<a href="${a.url}" class="${a.cls || ''}">${t(a.labelKey)}</a>`
              : html`<button class="${a.cls || 'btn btn-primary'}" ?disabled=${a.loading ? this[a.loading] : false} @click=${() => this[a.click]()}>${t(a.labelKey)}</button>`
            )}
          </div>` : ''}
      </div>
    `, el);
  }

  renderProductSections() {
    if (this.dom.chartSection) this.dom.chartSection.style.display = this.product === 'chart' ? '' : 'none';
    if (this.dom.bondSection) this.dom.bondSection.style.display = this.product === 'bond' ? '' : 'none';
    if (this.dom.namingSection) this.dom.namingSection.style.display = this.product === 'naming' ? '' : 'none';

    if (this.product === 'chart') this.renderChart();
    else if (this.product === 'bond') this.renderBond();
    else if (this.product === 'naming') this.renderNaming();
  }

  // ── CHART ──

  renderChart() {
    const c = this.chart;
    const p = this.pillars;

    // summary
    if (this.dom.chartSummary) {
      const yong = c.yong_shen || {};
      const fuyi = yong.fuyi || {};
      render(html`
        <div class="summary-grid">
          <div class="summary-card"><div class="label">${t('chart.dayMaster')}</div><div class="value">${c.riyuan || ''}</div></div>
          <div class="summary-card"><div class="label">${t('chart.patternYong')}</div><div class="value" style="color:#b45309">${fuyi.geju ? fuyi.geju + ' / ' + fuyi.yong : ''}</div></div>
          <div class="summary-card"><div class="label">${t('chart.wangShuai')}</div><div class="value">${fuyi.qiangruo || ''}</div></div>
          <div class="summary-card"><div class="label">${t('chart.zodiac')}</div><div class="value">${c.zodiac || ''}</div></div>
        </div>
      `, this.dom.chartSummary);
    }

    // pillars table
    if (this.dom.chartTableBody) {
      const pillarLabels = [t('pillar.year'), t('pillar.month'), t('pillar.day'), t('pillar.hour')];
      render(html`
        ${p.map((pv, i) => html`
          <tr>
            <td class="text-stone-500 text-sm">${pillarLabels[i]}</td>
            <td class="font-medium">${pv.gan}<br><span class="text-xs text-stone-500">${pv.ten_god || ''}</span></td>
            <td class="font-medium">${pv.zhi}</td>
            <td class="text-xs text-stone-500">${(pv.cang_gan || []).join(' ')}</td>
            <td class="text-xs text-stone-500">${pv.nayin || ''}</td>
            <td class="text-xs">${(pv.shensha || []).map(s => html`<span class="tag tag-amber mr-1">${s}</span>`)}</td>
          </tr>
        `)}
      `, this.dom.chartTableBody);
    }

    // wuxing
    if (this.dom.chartWuxing) {
      const wx = c.wuxing;
      render(html`
        ${wx ? Object.keys(wx).map(k => html`<span class="tag tag-green text-sm">${k}: ${wx[k]}</span>`) : ''}
      `, this.dom.chartWuxing);
    }

    // tiaohou
    if (this.dom.chartTiaohou) {
      const th = (c.yong_shen && c.yong_shen.tiaohou) ? c.yong_shen.tiaohou : null;
      if (th && th.yong) {
        this.dom.chartTiaohou.style.display = '';
        render(html`
          <div class="section-card">
            <h2>${t('chart.tiaohou')}</h2>
            <div class="summary-grid">
              <div class="summary-card"><div class="label">${t('chart.tiaohouYong')}</div><div class="value" style="color:#b45309">${th.yong}</div></div>
              <div class="summary-card"><div class="label">${t('chart.tiaohouSeason')}</div><div class="value">${th.season || ''}</div></div>
              <div class="summary-card"><div class="label">${t('chart.tiaohouXi')}</div><div class="value">${th.xi || ''}</div></div>
              <div class="summary-card"><div class="label">${t('chart.tiaohouJi')}</div><div class="value">${th.ji || ''}</div></div>
            </div>
            ${th.detail ? html`<p class="text-sm text-stone-500 mt-3">${th.detail}</p>` : ''}
          </div>
        `, this.dom.chartTiaohou);
      } else {
        this.dom.chartTiaohou.style.display = 'none';
      }
    }

    // dayun
    if (this.dom.chartDayun) {
      const dy = c.dayun;
      if (dy) {
        this.dom.chartDayun.style.display = '';
        render(html`
          <div class="section-card">
            <h2>${t('chart.dayun')}</h2>
            <p class="text-sm text-stone-500 mb-3">${t('chart.dayun.startAge').replace('{0}', dy.start_age)}</p>
            <div class="flex flex-wrap gap-2">
              ${dy.pillars.map(dp => html`<span class="tag tag-green">${dp.gan + dp.zhi + ' (' + dp.age_start + '-' + dp.age_end + '岁)'}</span>`)}
            </div>
          </div>
        `, this.dom.chartDayun);
      } else {
        this.dom.chartDayun.style.display = 'none';
      }
    }

    // interpretation
    if (this.dom.chartInterpretation) this.dom.chartInterpretation.innerHTML = this.llmHTML;
  }

  // ── BOND ──

  renderBond() {
    const c = this.chart;
    const b = this.bond;

    // day masters
    if (this.dom.bondDayMasters) {
      render(html`
        <div class="section-card">
          <h2>${t('bond.dayMasters')}</h2>
          <div class="summary-grid">
            <div class="summary-card"><div class="label">${t('bond.dayMasterA')}</div><div class="value">${(c.chart_a && c.chart_a.riyuan) || ''}</div></div>
            <div class="summary-card"><div class="label">${t('bond.dayMasterB')}</div><div class="value">${(c.chart_b && c.chart_b.riyuan) || ''}</div></div>
          </div>
        </div>
      `, this.dom.bondDayMasters);
    }

    // gan relations
    if (this.dom.bondGanRel) {
      if (b.gan_rel && b.gan_rel.length) {
        this.dom.bondGanRel.style.display = '';
        render(html`
          <div class="section-card">
            <h2>${t('bond.ganRelation')}</h2>
            ${b.gan_rel.map(r => html`
              <div class="rel-row">
                <span class="text-sm">${r.from} → ${r.to}</span>
                <div><span class="tag tag-amber">${r.type}</span><span class="text-xs text-stone-500 ml-2">${r.label || ''}</span></div>
              </div>
            `)}
          </div>
        `, this.dom.bondGanRel);
      } else {
        this.dom.bondGanRel.style.display = 'none';
      }
    }

    // zhi relations
    if (this.dom.bondZhiRel) {
      if (b.zhi_rel && b.zhi_rel.length) {
        this.dom.bondZhiRel.style.display = '';
        render(html`
          <div class="section-card">
            <h2>${t('bond.zhiRelation')}</h2>
            ${b.zhi_rel.map(z => html`
              <div class="rel-row">
                <span class="text-sm">${z.from} → ${z.to}</span>
                <span class="tag tag-green">${z.type}</span>
              </div>
            `)}
          </div>
        `, this.dom.bondZhiRel);
      } else {
        this.dom.bondZhiRel.style.display = 'none';
      }
    }

    // key hints
    if (this.dom.bondKeyHints) {
      if (b.key_hints && b.key_hints.length) {
        this.dom.bondKeyHints.style.display = '';
        render(html`
          <div class="bg-amber-50 border border-amber-100 rounded-xl p-6 mb-3">
            <h2 class="text-lg font-semibold mb-3 text-amber-800">${t('bond.keyHints')}</h2>
            ${b.key_hints.map(h => html`<div class="flex items-start gap-2 mb-1"><span class="text-amber-500 mt-1">•</span><span class="text-sm text-amber-800">${h}</span></div>`)}
          </div>
        `, this.dom.bondKeyHints);
      } else {
        this.dom.bondKeyHints.style.display = 'none';
      }
    }

    // interpretation
    if (this.dom.bondInterpretation) this.dom.bondInterpretation.innerHTML = this.llmHTML;
  }

  // ── NAMING ──

  renderNaming() {
    const nd = this.namingData;
    const analysis = nd.analysis || {};

    // summary
    if (this.dom.namingSummary) {
      render(html`
        <div class="section-card">
          <h2>${t('naming.summary')}</h2>
          <div class="summary-grid">
            <div class="summary-card"><div class="label">${t('form.surname')}</div><div class="value">${analysis.surname || ''}</div></div>
            <div class="summary-card"><div class="label">${t('naming.yongShen')}</div><div class="value" style="color:#b45309">${analysis.yong_shen || ''}</div></div>
          </div>
        </div>
      `, this.dom.namingSummary);
    }

    // candidates
    if (this.dom.namingCandidates) {
      const cands = nd.candidates || [];
      render(html`
        <div class="section-card">
          <h2>${t('naming.recommendations')}</h2>
          ${cands.map(c => {
            const wuge = c.wu_ge || {};
            return html`
              <div class="name-card">
                <div class="flex justify-between items-center mb-2">
                  <span class="name-title">${c.name}</span>
                  <span class="tag tag-green">${(c.san_cai && c.san_cai.configuration) || ''}</span>
                </div>
                <div class="flex flex-wrap gap-1 mb-2">
                  ${(c.highlights || []).map(h => html`<span class="tag tag-amber">${h}</span>`)}
                </div>
                <div class="grid grid-cols-2 gap-1 text-xs text-stone-500">
                  ${(c.characters || []).map(ch => html`<span>${ch.hanzi}(${ch.strokes || 0}画)${ch.wuxing || ''}</span>`)}
                </div>
                <div class="text-xs text-stone-500 mt-1">
                  ${t('naming.wuge').replace('{0}', (wuge.tian_ge && wuge.tian_ge.stroke) || '').replace('{1}', (wuge.ren_ge && wuge.ren_ge.stroke) || '').replace('{2}', (wuge.di_ge && wuge.di_ge.stroke) || '').replace('{3}', (wuge.wai_ge && wuge.wai_ge.stroke) || '').replace('{4}', (wuge.zong_ge && wuge.zong_ge.stroke) || '')}
                </div>
              </div>
            `;
          })}
        </div>
      `, this.dom.namingCandidates);
    }

    // interpretation
    if (this.dom.namingInterpretation) this.dom.namingInterpretation.innerHTML = this.llmHTML;
  }

  // ── elapsed timer ──

  startElapsed() {
    this.elapsed = 0;
    this.stopElapsed();
    this.elapsedTimer = setInterval(() => {
      this.elapsed++;
      this.renderStatusCard();
    }, 1000);
  }

  stopElapsed() {
    if (this.elapsedTimer) { clearInterval(this.elapsedTimer); this.elapsedTimer = null; }
  }

  // ── progress update ──

  updateProgress(tries, max) {
    this.pollProgress = tries;
    this.pollMax = max;
    this.renderStatusCard();
  }

  // ── core logic ──

  init() {
    const orderID = this.orderIDFromURL();
    if (!orderID) {
      this.phase = 'error';
      this.error = t('report.notFound');
      this.render();
      return;
    }
    this.loadReport(orderID);
  }

  orderIDFromURL() {
    const m = location.pathname.match(/\/report\/([a-f0-9-]+)/i);
    return m ? m[1] : '';
  }

  showReport(llmJSON, withTransition) {
    this.stopElapsed();
    this.llmHTML = window.Liki.renderMD(llmJSON);
    this.setMeta();
    if (!withTransition) {
      this.phase = 'ready';
      this.render();
      return;
    }
    this.phase = 'unlocking';
    this.render();
    setTimeout(() => {
      const el = document.getElementById('transition-overlay');
      if (el) el.classList.add('fade-out');
      setTimeout(() => {
        this.phase = 'ready';
        this.render();
      }, 400);
    }, 1000);
  }

  async loadReport(orderID) {
    this.phase = 'loading';
    this.error = '';
    this.render();

    try {
      const data = await window.Liki.apiGet('/reports/' + orderID);
      this.product = data.product || '';
      if (this.product && ['chart', 'bond', 'naming'].indexOf(this.product) === -1) {
        this.phase = 'error';
        this.error = t('report.statusError');
        this.render();
        return;
      }
      if (data.chart_json && data.status !== 'pending') {
        this.parseChartData(data.product, data.chart_json);
      }
      if (data.status === 'paid' && data.llm_json) {
        this.showReport(data.llm_json);
      } else if (data.status === 'paid' && !data.llm_json) {
        this.startPolling(orderID, 'generating');
      } else if (data.status === 'pending') {
        this.startPolling(orderID, 'payment');
      } else {
        this.phase = 'error';
        this.error = t('report.statusError');
        this.render();
      }
    } catch (e) {
      this.phase = 'error';
      this.error = e.message || t('report.loadError');
      this.render();
    }
  }

  parseChartData(product, chartJSON) {
    try {
      const raw = JSON.parse(chartJSON);
      if (product === 'chart') {
        const cd = (raw.chart && raw.chart.chart) ? raw.chart.chart : raw;
        this.chart = cd;
        const pillars = [cd.nianzhu, cd.yuezhu, cd.rizhu, cd.shizhu];
        this.pillars = pillars.map(p => ({
          gan: (p && p.gan) || '',
          zhi: (p && p.zhi) || '',
          ten_god: (p && p.shishen || []).filter(t => t.source === 'gan').map(t => t.shishen).join(', '),
          cang_gan: p && p.canggan ? Object.values(p.canggan) : [],
          nayin: (p && p.nayin) || '',
          shensha: (p && p.shensha || []).map(s => s.name)
        }));
        this.bond = { gan_rel: [], zhi_rel: [], key_hints: [] };
      } else if (product === 'bond') {
        const ca = (raw.chart_a && raw.chart_a.chart) ? raw.chart_a.chart : (raw.chart_a || {});
        const cb = (raw.chart_b && raw.chart_b.chart) ? raw.chart_b.chart : (raw.chart_b || {});
        this.bond = raw.bond || {};
        this.pillars = [];
        this.chart = {
          chart_a: ca,
          chart_b: cb,
          riyuan: (ca.riyuan || '') + ' / ' + (cb.riyuan || '')
        };
      } else if (product === 'naming') {
        this.namingData = raw.naming || raw;
        this.chart = (raw.naming && raw.naming.analysis) ? raw.naming.analysis : (raw.analysis || {});
      }
    } catch (e) {
      console.error('parse chart data:', e);
    }
  }

  startPolling(orderID, type) {
    const POLL_BASE = 2000;
    const POLL_MAX = 16000;
    const MAX_TRIES = type === 'generating' ? 60 : 15;
    const MAX_ERRORS = 6;

    this.phase = type;
    this.startElapsed();
    this.pollMax = MAX_TRIES;
    this.pollProgress = 0;
    this.render();

    let tries = 0;
    let errors = 0;
    let delay = POLL_BASE;

    const schedule = (fn) => { setTimeout(fn, delay); };

    const poll = () => {
      tries++;
      this.updateProgress(tries, MAX_TRIES);
      if (tries > MAX_TRIES) {
        this.stopElapsed();
        this.phase = 'timeout_' + type;
        this.render();
        return;
      }
      schedule(async () => {
        if (this.phase !== type) return;
        try {
          const data = await window.Liki.apiGet('/reports/' + orderID);
          errors = 0;
          delay = POLL_BASE;
          if (data.status === 'paid' && data.llm_json) {
            this.stopElapsed();
            this.showReport(data.llm_json, true);
            return;
          }
          if (data.status === 'paid' && !data.llm_json) {
            if (this.phase !== 'generating') { this.phase = 'generating'; this.render(); }
            poll();
            return;
          }
          if (data.status === 'pending') {
            poll();
            return;
          }
          this.stopElapsed();
          this.phase = 'error';
          this.error = t('report.statusError');
          this.render();
        } catch (e) {
          errors++;
          if (errors > MAX_ERRORS) {
            this.stopElapsed();
            this.phase = 'timeout_' + type;
            this.render();
            return;
          }
          delay = Math.min(delay * 2, POLL_MAX);
          poll();
        }
      });
    };
    poll();
  }

  retryPoll() {
    const orderID = this.orderIDFromURL();
    if (orderID) this.loadReport(orderID);
  }

  async checkPaymentStatus() {
    const orderID = this.orderIDFromURL();
    if (!orderID) return;
    try {
      const data = await window.Liki.apiPost('/orders/' + orderID + '/retry');
      if (data.status === 'paid' && data.llm_json) {
        this.showReport(data.llm_json, true);
        return;
      }
      if (data.status === 'paid' && !data.llm_json) {
        this.phase = 'generating';
        this.startElapsed();
        this.startPolling(orderID, 'generating');
        return;
      }
      this.retryPoll();
    } catch (e) {
      window.Liki.showToast(t('report.retryFailed'), 'error');
    }
  }

  async goPay() {
    const orderID = this.orderIDFromURL();
    if (!orderID || this.payLoading) return;
    this.payLoading = true;
    this.renderStatusCard();
    try {
      await window.Liki.goPay(orderID);
    } catch (e) {
      this.error = e.message;
      this.payLoading = false;
      this.renderStatusCard();
    }
  }

  setMeta() {
    const titleKey = 'report.title.' + (this.product || 'default');
    const title = t(titleKey) + ' · ' + t('site.name');
    document.title = title;
    const descEl = document.querySelector('meta[name="description"]');
    if (descEl) descEl.content = t('report.meta.desc');
    const ogTitle = document.querySelector('meta[property="og:title"]');
    if (ogTitle) ogTitle.content = title;
    const twTitle = document.querySelector('meta[name="twitter:title"]');
    if (twTitle) twTitle.content = title;
  }

  copyLink() {
    navigator.clipboard.writeText(location.href).then(() => {
      this.copyBtnText = t('report.copied');
      if (this.dom.bannerCopyText) this.dom.bannerCopyText.textContent = this.copyBtnText;
      setTimeout(() => {
        this.copyBtnText = t('report.copyLink');
        if (this.dom.bannerCopyText) this.dom.bannerCopyText.textContent = this.copyBtnText;
      }, 2000);
    }).catch(() => {
      window.Liki.showToast(t('report.copyFailed'), 'error');
    });
  }

  closeBanner() {
    this.bannerClosed = true;
    sessionStorage.setItem('report-banner-closed', '1');
    if (this.dom.banner) this.dom.banner.style.display = 'none';
  }

  share() {
    const shareText = t('report.shareText');
    if (navigator.share) {
      navigator.share({ title: document.title, text: shareText, url: location.href }).catch(() => {});
    } else {
      navigator.clipboard.writeText(location.href).then(() => {
        window.Liki.showToast(t('report.copied'), 'success');
      }).catch(() => {
        window.Liki.showToast(t('report.copyFailed'), 'error');
      });
    }
  }
}

new ReportApp();
