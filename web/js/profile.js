// profile.js — Profile page at /profile/{name}
function profilePage() {
  return {
    loading: true,
    profile: null,
    bonds: [],
    reviewLinks: [],
    matchLinks: [],
    mingliMatchLinks: [],
    history: [],
    showBond: false,
    bondA: null,
    bondB: null,
    bondDeltasA: [],
    bondDeltasB: [],
    bondInfluencerA: '',
    bondInfluencerB: '',
    flowDeltas: [],
    flowYearly: null,
    hasMyProfile: false,
    myProfile: null,
    targetReviewToken: null,

    // --- BaZi state ---
    birth: {
      year: 1984, month: 2, day: 5, hour: 12, minute: 0,
      longitude: 120.0, timezone: 120.0, is_dst: false, gender: 'male',
    },
    birthDate: '1984-02-05',
    birthTime: '12:00',
    mingliLoading: false,
    mingliError: '',
    chart: null,
    editing: false,
    cityQuery: '',
    filteredCities: [],
    showCityList: false,
    selectedCityName: '',
    showLunar: isLocaleChinese(),
    calendarType: isLocaleChinese() ? 'lunar' : 'solar',
    lunarYear: 1984,
    lunarMonth: 1,
    lunarDay: 1,

    setLunarDate() {
      var m = this.lunarMonth;
      var leap = m > 100;
      if (leap) m -= 100;
      this.birthDate = lunarToDate(this.birth, this.lunarYear, m, this.lunarDay, leap);
      if (isChinaDST(this.birth.year, this.birth.month, this.birth.day)) {
        this.birth.is_dst = true;
      }
    },

    parseDatetime() {
      parseBirthDatetime(this.birthDate, this.birthTime, this.birth);
    },

    get auth() { return Alpine.store('auth'); },
    get currentLocale() { return window.CURRENT_LOCALE || 'en'; },
    get origin() { return window.location.origin; },
    get identity() { return this.profile && this.profile.profile ? this.profile.profile.identity : null; },
    get pValues() { return this.profile && this.profile.profile ? namedToArray(this.profile.profile.p) : []; },
    get peerPValues() { return this.profile && this.profile.peers ? namedToArray(this.profile.peers.p) : []; },
    get flowMonth() { return this.profile && this.profile.flow_month; },
    get peerIdentity() { return this.profile && this.profile.peers ? this.profile.peers.identity : null; },
    get hasPeers() { return this.profile && !!this.profile.peers; },
    get isOwner() { return this.profile && this.profile.is_owner; },
    get mingliChart() { return this.profile && this.profile.mingli_chart || null; },
    get birthInfo() { return this.profile && this.profile.birth_info || null; },
    get hasBonds() { return this.bonds.length > 0; },
    get latestBond() {
      for (var i = 0; i < this.bonds.length; i++) {
        if (this.bonds[i].bond) return this.bonds[i];
      }
      return null;
    },
    get hasLatestBond() { return !!this.latestBond; },
    get hasHistory() { return this.history.length > 1; },
    get hasLinks() { return this.reviewLinks.length > 0 || this.matchLinks.length > 0 || this.mingliMatchLinks.length > 0; },
    get hasFlowRiver() { return this.flowYearly && this.flowYearly.months && this.flowYearly.months.length >= 5; },
    get flowRiverMonths() {
      if (!this.flowYearly || !this.flowYearly.months) return [];
      var months = this.flowYearly.months;
      var cur = this.flowYearly.current;
      var curIdx = 0;
      for (var i = 0; i < months.length; i++) {
        if (months[i].id === cur) { curIdx = i; break; }
      }
      var baseP = this._pArr;
      var result = [];
      for (var j = -1; j <= 3; j++) {
        var idx = (curIdx + j + months.length) % months.length;
        var m = months[idx];
        var name = this.currentLocale === 'zh-CN' ? m.id : (m.name_en || m.id);
        // Base profile proportions — direction marked via arrows on the chart.
        var p = baseP ? baseP.slice() : [0.2, 0.2, 0.2, 0.2, 0.2];
        result.push({
          label: name,
          p: p,
          generates: m.generates,
          restrains: m.restrains,
        });
      }
      return result;
    },

    getTargetName() {
      var path = window.location.pathname.replace(/\/+$/, '');
      var segments = path.split('/').filter(Boolean);
      var loc = window.CURRENT_LOCALE || 'en';
      if (segments[0] === loc) segments.shift();
      // Path: /profile/{name} or /profile/{name}/profile.html
      if (segments[0] === 'profile') segments.shift();
      if (segments.length > 0 && segments[segments.length - 1] === 'profile.html') segments.pop();
      if (segments.length === 0) {
        var params = new URLSearchParams(window.location.search);
        var n = params.get('n');
        return n || '';
      }
      return segments[segments.length - 1];
    },

    init() {
      var name = this.getTargetName();
      if (!name) {
        this.loading = false;
        return;
      }
      document.title = name + ' — 25 Types';
      this.loadAll(name);
    },

    async loadAll(name) {
      await this.loadProfile(name);
      if (this.profile && this.profile.is_owner) {
        await Promise.all([
          this.loadBonds(name),
          this.loadReviewLinks(),
          this.loadMatchLinks(),
          this.loadMingliMatchLinks(),
          this.loadHistory(),
          this.loadFlowYearly(),
        ]);
      } else if (this.profile && this.auth.id) {
        await Promise.all([
          this.checkMyProfile(),
          this.checkTargetReviewLinks(),
        ]);
      }
    },

    async loadProfile(name) {
      try {
        var res = await api('/api/profiles/' + encodeURIComponent(name));
        this.profile = res.data || res;
        document.title = (this.profile && this.profile.user ? this.profile.user.name : name) + ' — 25 Types';
      } catch (e) {
        console.error(e);
        this.profile = null;
      }
      this.loading = false;

      // Sync BaZi chart and birth info from profile data
      this.chart = adaptMingliChart(this.profile && this.profile.mingli_chart);
      var bi = this.profile && this.profile.birth_info;
      if (bi) {
        this.birth = Object.assign({}, bi);
        var dt = fmtDatetimeLocal(bi.year, bi.month, bi.day, bi.hour, bi.minute).split('T');
        this.birthDate = dt[0];
        this.birthTime = dt[1];
      }

      if (this.profile && this.profile.profile) {
        this._pArr = namedToArray(this.profile.profile.p);
        this.$nextTick(function () {
          this.renderProfileRadar();
          if (this.profile.peers) this.renderPeerRadar();
          if (this.flowMonth) this.renderFlowMiniRadar();
        }.bind(this));
      }
    },

    async loadBonds(name) {
      try {
        var res = await api('/api/profiles/' + encodeURIComponent(name) + '/bonds');
        this.bonds = (res.data && res.data.items) || res.items || [];
        if (this.hasLatestBond) {
          var chartData = this.prepareLatestBondData();
          var self = this;
          this.$nextTick(function () { self.renderLatestBondCharts(chartData); });
        }
      } catch (e) { console.error(e); this.bonds = []; }
    },

    prepareLatestBondData() {
      var b = this.latestBond;
      if (!b || !b.bond) return null;
      var shapes = computeBondShapes(b.bond);
      this.bondA = identityFromD(b.bond.self);
      this.bondB = identityFromD(b.bond.other);
      this.bondInfluencerA = this.bondB ? this.bondB.label : '';
      this.bondInfluencerB = this.bondA ? this.bondA.label : '';
      this.bondDeltasA = shapes.selfDeltas;
      this.bondDeltasB = shapes.otherDeltas;
      var sp = concordProps(b.bond.concord);
      this.concordLabel = sp.label;
      this.concordBadgeClass = sp.badgeClass;
      this.concordDesc = sp.desc;
      return { origSelf: shapes.origSelf, origOther: shapes.origOther, pSelf: shapes.pSelf, pOther: shapes.pOther };
    },

    renderLatestBondCharts(d) {
      if (!d) return;
      this.renderBondInfluence(d.origSelf, d.origOther, d.pSelf, d.pOther);
    },

    async _loadLinks(path, field) {
      try {
        var res = await api(path);
        this[field] = (res.data && res.data.items) || res.items || [];
      } catch (e) { console.error(e); this[field] = []; }
    },
    async loadReviewLinks() { await this._loadLinks('/api/reviews', 'reviewLinks'); },
    async loadMatchLinks() { await this._loadLinks('/api/match-links', 'matchLinks'); },
    async loadMingliMatchLinks() { await this._loadLinks('/api/match-links', 'mingliMatchLinks'); },

    async loadHistory() {
      try {
        var res = await api('/api/assessments');
        this.history = (res.data && res.data.items) || res.items || [];
        if (this.history.length > 1) {
          this.$nextTick(function () { this.renderHistoryChart(); }.bind(this));
        }
      } catch (e) { console.error(e); this.history = []; }
    },

    async checkMyProfile() {
      if (!this.auth.id) return;
      try {
        var res = await api('/api/profiles/' + encodeURIComponent(this.auth.name));
        this.hasMyProfile = !!(res.data && res.data.profile);
        if (this.hasMyProfile) this.myProfile = res.data;
      } catch (e) { console.error(e); this.hasMyProfile = false; }
    },

    checkTargetReviewLinks() {
      this.targetReviewToken = (this.profile && this.profile.has_review_link) || null;
    },

    // --- Radar rendering ---
    async renderProfileRadar() {
      if (!this.profile || !this.profile.profile) return;
      var el = document.getElementById('profile-radar');
      if (!el) return;
      if (window._loadECharts) await window._loadECharts();
      this.profileRadarInst = Charts.renderElementRadar(el, {
        series: [{ value: this._pArr, name: this.profile.profile.identity.label }],
      }, {
        radar: { center: ['50%', '50%'], radius: '65%' },
      });
    },

    renderPeerRadar() {
      if (!this.profile || !this.profile.peers) return;
      var el = document.getElementById('peers-radar');
      if (!el) return;
      Charts.renderElementRadar(el, {
        series: [
          { value: this._pArr, name: this.$store.locale.t('peers.self'), lineStyle: { width: 1.5 }, areaStyle: { opacity: 0.04 } },
          { value: namedToArray(this.profile.peers.p), name: this.$store.locale.t('peers.peerReviews'), colorIdx: 1, symbol: 'diamond', areaStyle: { opacity: 0.06 } },
        ],
        legend: true,
      });
    },

    renderFlowMiniRadar() {
      var el = document.getElementById('flow-mini-radar');
      if (!el || !this.flowMonth || !this.profile || !this.profile.profile) return;
      var pOrig = this._pArr;
      // Visual proportions: base + direction offset
      var pFlow = pOrig ? pOrig.slice() : [0.2, 0.2, 0.2, 0.2, 0.2];
      if (typeof this.flowMonth.generates === 'number' && this.flowMonth.generates >= 0 && this.flowMonth.generates < 5) {
        pFlow[this.flowMonth.generates] += 0.02;
      }
      if (typeof this.flowMonth.restrains === 'number' && this.flowMonth.restrains >= 0 && this.flowMonth.restrains < 5) {
        pFlow[this.flowMonth.restrains] -= 0.02;
      }
      this.flowDeltas = ELEMENT_HTML_ORDER.map(function(i) {
        return { idx: i, delta: pFlow[i] - pOrig[i] };
      });
      Charts.renderElementRadar(el, {
        series: [
          { value: pOrig, name: this.$store.locale.t('peers.self'), lineStyle: { width: 1.5, type: 'dashed', opacity: 0.6 }, areaStyle: { opacity: 0.04 } },
          { value: pFlow, name: this.flowMonth.month_en || this.flowMonth.month_id, colorIdx: 2 },
        ],
        legend: true,
      });
    },

    renderHistoryChart() {
      var el = document.getElementById('history-chart');
      if (!el || !this.history.length || !window.Charts) return;

      var self = this;
      var dates = [], dAll = [], typeLabels = [];
      for (var i = this.history.length - 1; i >= 0; i--) {
        var a = this.history[i];
        var d = new Date(a.created_at);
        dates.push(d.toLocaleDateString(self.currentLocale === 'zh-CN' ? 'zh-CN' : 'en-US', { month: 'short', day: 'numeric' }));
        if (!a.profile) { dAll.push(null); }
        else if (a.profile.p) { dAll.push(namedToArray(a.profile.p)); }
        else if (a.profile.d) { dAll.push(deviationToProportion(a.profile.d)); }
        else { dAll.push(null); }
        typeLabels.push(a.identity ? a.identity.label : null);
      }

      Charts.renderElementLine(el, {
        categories: dates,
        series: ELEMENT_DISPLAY_ORDER.map(function(idx) {
          var s = { colorIdx: idx, data: dAll.map(function(d) { return d ? d[idx] : null; }) };
          if (idx === 0) {
            var bc = getComputedStyle(document.documentElement).getPropertyValue('--bc').trim() || '#1e293b';
            s.markPoint = {
              symbol: 'pin',
              symbolSize: 36,
              symbolOffset: [0, '-50%'],
              label: { fontSize: 10, fontWeight: 'bold', color: bc },
              data: typeLabels.map(function(label, i) {
                if (!label || !dAll[i]) return null;
                return { name: label, xAxis: i, yAxis: dAll[i][idx] };
              }).filter(Boolean),
            };
          }
          return s;
        }),
      });
    },

    async loadFlowYearly() {
      if (!this.auth.id) return;
      try {
        var res = await api('/api/flow/yearly');
        this.flowYearly = res.data || res;
        if (this.hasFlowRiver) {
          this.$nextTick(function () { this.renderFlowRiverChart(); }.bind(this));
        }
      } catch (e) { console.error(e); this.flowYearly = null; }
    },

    renderFlowRiverChart() {
      var el = document.getElementById('flow-river-chart');
      if (!el || !this.hasFlowRiver || !window.Charts) return;
      Charts.renderFlowRiver(el, this.flowRiverMonths, 1);
    },

    // --- Owner actions ---
    shareProfile() {
      var url = this.origin + '/' + this.currentLocale + '/profile/' + encodeURIComponent(this.profile.user.name);
      copyText(url, { quiet: true });
      this.$store.toast.success(this.$store.locale.t('profile.profileShared'));
    },

    async generateShareCard() {
      if (!this.profile || !this.profile.profile) return;
      await Charts.generateAndSaveShareCard(this.profileRadarInst, this.profile.profile.identity, this._pArr);
    },

    async _createLink(path, reloadFn) {
      try {
        await api(path, { method: 'POST' });
        this.$store.toast.success(this.$store.locale.t('profile.linkCreated'));
        await reloadFn.call(this);
      } catch (e) { console.error(e); }
    },
    async createReviewLink() { await this._createLink('/api/reviews', this.loadReviewLinks); },
    async createMatchLink() { await this._createLink('/api/match-links', this.loadMatchLinks); },
    async createMingliMatchLink() { await this._createLink('/api/match-links', this.loadMingliMatchLinks); },

    async _deleteLink(path, field, id) {
      if (!confirm(this.$store.locale.t('confirm.deleteLink'))) return;
      try {
        await api(path + '/' + id, { method: 'DELETE' });
        this[field] = this[field].filter(function(l) { return l.id !== id; });
      } catch (e) { console.error(e); }
    },
    async deleteReviewLink(id) { await this._deleteLink('/api/reviews', 'reviewLinks', id); },
    async deleteMatchLink(id) { await this._deleteLink('/api/match-links', 'matchLinks', id); },
    async deleteMingliMatchLink(id) { await this._deleteLink('/api/match-links', 'mingliMatchLinks', id); },

    // --- Visitor: instant compare (two influence cards) ---
    async compareWithMe() {
      if (!this.profile || !this.profile.user) return;
      try {
        var bond;
        var body = {};
        if (this.profile.user.id) {
          body.with_user_id = this.profile.user.id;
        } else {
          body.with_name = this.profile.user.name;
        }
        var res = await api('/api/bond', { method: 'POST', body: JSON.stringify(body) });
        bond = res.data || res;
        if (!bond) return;

        var shapes = computeBondShapes(bond);

        var myId = this.myProfile && this.myProfile.profile && this.myProfile.profile.identity;
        this.bondA = myId || identityFromD(bond.self);
        this.bondInfluencerA = this.identity ? this.identity.label : '';

        this.bondB = this.identity || identityFromD(bond.other);
        this.bondInfluencerB = myId ? myId.label : '';

        this.bondDeltasA = shapes.selfDeltas;
        this.bondDeltasB = shapes.otherDeltas;

        var sp = concordProps(bond.concord);
        this.concordLabel = sp.label;
        this.concordBadgeClass = sp.badgeClass;
        this.concordDesc = sp.desc;

        this.showBond = true;
        var self = this;
        this.$nextTick(function () { self.renderBondInfluence(shapes.origSelf, shapes.origOther, shapes.pSelf, shapes.pOther); });
      } catch (e) { console.error(e); }
    },

    renderBondInfluence(origSelf, origOther, pSelf, pOther) {
      var locale = this.$store.locale;
      Charts.renderBondInfluenceChart(document.getElementById('bond-influence-self'), {
        origData: origSelf, pData: pSelf,
        yourLabel: locale.t('bond.yourShape'), influencedLabel: locale.t('bond.influenced'),
        overrides: { radar: { center: ['50%', '50%'], radius: '60%' } },
      });
      Charts.renderBondInfluenceChart(document.getElementById('bond-influence-other'), {
        origData: origOther, pData: pOther,
        yourLabel: locale.t('bond.theirShape'), influencedLabel: locale.t('bond.influenced'),
        overrides: { radar: { center: ['50%', '50%'], radius: '60%' } },
      });
    },

    // --- BaZi form methods ---
    filterCities() {
      var self = this;
      clearTimeout(this._cityTimer);
      this._cityTimer = setTimeout(async function () {
        self.filteredCities = await fetchFilteredCities(self.cityQuery);
        self.showCityList = self.filteredCities.length > 0;
      }, 200);
    },

    selectCity(c) {
      applyCityToBirth(c, this.birth);
      this.cityQuery = c.name_zh || c.name;
      this.selectedCityName = cityDisplayName(c);
      this.showCityList = false;
      if (CHINESE_COUNTRIES[c.country]) this.showLunar = true;
    },

    startEdit() {
      if (this.profile && this.profile.birth_info) {
        var bi = this.profile.birth_info;
        this.birth = Object.assign({}, bi);
        var dt = fmtDatetimeLocal(bi.year, bi.month, bi.day, bi.hour, bi.minute);
        var parts = dt.split('T');
        this.birthDate = parts[0];
        this.birthTime = parts[1];
      }
      this.editing = true;
    },

    cancelEdit() {
      this.editing = false;
      if (this.profile && this.profile.birth_info) {
        this.birth = Object.assign({}, this.profile.birth_info);
      }
    },

    async computeAndSave() {
      this.parseDatetime();
      this.mingliLoading = true;
      this.mingliError = '';
      try {
        var [chartResp] = await Promise.all([
          api('/api/mingli/bazi/chart', {
            method: 'POST',
            body: JSON.stringify(birthToPayload(this.birth)),
          }),
          api('/api/users/me', {
            method: 'PATCH',
            body: JSON.stringify({ birth_info: this.birth }),
          }),
        ]);

        var raw = chartResp.data || chartResp;
        this.chart = adaptMingliChart(raw);
        // Update profile data so the card shows on refresh
        if (this.profile) {
          this.profile.birth_info = Object.assign({}, this.birth);
          this.profile.mingli_chart = raw;
        }
        this.editing = false;
        this.$store.toast.success(this.$store.locale.t('profile.mingli.baziSaveSuccess'));
      } catch (e) {
        this.mingliError = e.message || 'Unknown error';
      } finally {
        this.mingliLoading = false;
      }
    },

    // --- Helpers ---
    fmtDate(s) {
      return formatDate(s, this.currentLocale);
    },
  };
}
