import { html, render } from './lit-html.js';

const t = window.Liki.t;

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
    actions: [{ labelKey: 'report.pay', click: 'handlePay', loading: 'payLoading' }]
  },
  timeout_payment: {
    variant: 'timeout', icon: 'clock',
    titleKey: 'report.timeout', hintKey: 'report.timeoutHint',
    actions: [
      { labelKey: 'report.pay', click: 'handlePay', loading: 'payLoading' },
      { labelKey: 'report.checkStatus', click: 'checkPaymentStatus', cls: 'btn btn-sm bg-white border border-stone-200 text-stone-600 rounded' },
      { labelKey: 'report.refresh', click: 'retryPoll', cls: 'btn btn-sm bg-white border border-stone-200 text-stone-600 rounded' },
      { labelKey: 'report.copyLink', click: 'copyLink', cls: 'btn btn-sm bg-white border border-stone-200 text-stone-600 rounded' }
    ]
  },
};

class ReportApp {
  constructor() {
    this.phase = 'loading';
    this.error = '';
    this.payLoading = false;
    this.bannerClosed = sessionStorage.getItem('report-banner-closed') === '1';
    this.copyBtnText = t('report.copyLink');
    this.product = '';
    this.namingData = { analysis: {}, candidates: [] };
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
      namingSection: document.getElementById('naming-section'),
      namingSummary: document.getElementById('naming-summary'),
      namingCandidates: document.getElementById('naming-candidates'),
      namingInterpretation: document.getElementById('naming-interpretation')
    };

    this.bindEvents();
    if (i18next.isInitialized) {
      this.render();
      this.init();
    } else {
      document.addEventListener('i18n:loaded', () => {
        this.render();
        this.init();
      }, { once: true });
    }
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
    const showProgress = this.phase === 'payment';

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
    if (this.dom.namingSection) this.dom.namingSection.style.display = this.product === 'naming' ? '' : 'none';
    if (this.product === 'naming') this.renderNaming();
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
                  ${(c.characters || []).map(ch => html`<span>${ch.hanzi}(${ch.strokes || 0}${i18next.t('report.strokesUnit')})${ch.wuxing || ''}</span>`)}
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
    this.llmHTML = llmJSON ? window.Liki.renderMD(llmJSON) : '';
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
      if (this.product && this.product !== 'naming') {
        this.phase = 'error';
        this.error = t('report.statusError');
        this.render();
        return;
      }
      if (data.chart_json && data.status !== 'pending') {
        this.parseChartData(data.chart_json);
      }
      if (data.status === 'paid' && data.llm_json) {
        this.showReport(data.llm_json);
      } else if (data.status === 'paid' && !data.llm_json) {
        this.showReport('');
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

  parseChartData(chartJSON) {
    try {
      const raw = JSON.parse(chartJSON);
      this.namingData = raw.naming || raw;
    } catch (e) {
      console.error('parse chart data:', e);
      this.error = t('report.loadError');
      this.phase = 'error';
      this.render();
    }
  }

  startPolling(orderID, type) {
    const POLL_BASE = 2000;
    const POLL_MAX = 16000;
    const MAX_TRIES = 15;
    const MAX_ERRORS = 6;

    this.phase = type;
    this.startElapsed();
    this.pollMax = MAX_TRIES;
    this.pollProgress = 0;
    this.render();

    let tries = 0;
    let errors = 0;
    let delay = POLL_BASE;

    let timer = null;
    let paused = false;

    const schedule = (fn) => { timer = setTimeout(fn, delay); };

    const onVisibility = () => {
      if (document.hidden) {
        paused = true;
        if (timer) { clearTimeout(timer); timer = null; }
        this.stopElapsed();
      } else if (paused && this.phase === type) {
        paused = false;
        this.startElapsed();
        delay = POLL_BASE;
        schedule(doPoll);
      }
    };
    document.addEventListener('visibilitychange', onVisibility);

    const cleanup = () => {
      document.removeEventListener('visibilitychange', onVisibility);
      if (timer) { clearTimeout(timer); timer = null; }
    };

    const doPoll = () => {
      if (document.hidden) { paused = true; return; }
      tries++;
      this.updateProgress(tries, MAX_TRIES);
      if (tries > MAX_TRIES) {
        cleanup();
        this.stopElapsed();
        this.phase = 'timeout_payment';
        this.render();
        return;
      }
      schedule(async () => {
        if (this.phase !== type) { cleanup(); return; }
        try {
          const data = await window.Liki.apiGet('/reports/' + orderID);
          errors = 0;
          delay = POLL_BASE;
          if (data.status === 'paid' && data.llm_json) {
            cleanup();
            this.stopElapsed();
            this.showReport(data.llm_json, true);
            return;
          }
          if (data.status === 'pending') {
            doPoll();
            return;
          }
          cleanup();
          this.stopElapsed();
          this.phase = 'error';
          this.error = t('report.statusError');
          this.render();
        } catch (e) {
          errors++;
          if (errors > MAX_ERRORS) {
            cleanup();
            this.stopElapsed();
            this.phase = 'timeout_payment';
            this.render();
            return;
          }
          delay = Math.min(delay * 2, POLL_MAX);
          doPoll();
        }
      });
    };
    doPoll();
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
        this.showReport('');
        return;
      }
      this.retryPoll();
    } catch (e) {
      window.Liki.showToast(t('report.retryFailed'), 'error');
    }
  }

  async handlePay() {
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
    const ogUrl = document.getElementById('og-url');
    if (ogUrl) ogUrl.content = window.location.href;
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
