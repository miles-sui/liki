// profile.js — Profile page at /{name}
function profilePage() {
  return {
    loading: true,
    profile: null,
    bonds: [],
    reviewLinks: [],
    matchLinks: [],
    history: [],
    showBond: false,
    bondDeltaA: [],
    bondDeltaB: [],
    flowDeltas: [],
    hasMyProfile: false,
    targetReviewToken: null,

    get auth() { return Alpine.store('auth'); },
    get currentLocale() { return window.CURRENT_LOCALE || 'en'; },
    get origin() { return window.location.origin; },
    get identity() { return this.profile && this.profile.profile ? this.profile.profile.identity : null; },
    get pValues() { return this.profile && this.profile.profile ? namedToArray(this.profile.profile.p) : []; },
    get flowMonth() { return this.profile && this.profile.flow_month; },
    get peerIdentity() { return this.profile && this.profile.peers ? this.profile.peers.identity : null; },

    // Extract username from URL path: /en/john → john, /john → john
    getTargetName() {
      var path = window.location.pathname.replace(/\/+$/, '');
      var segments = path.split('/').filter(Boolean);
      if (segments[0] === 'en' || segments[0] === 'zh-CN') {
        segments.shift();
      }
      if (segments.length > 0 && segments[segments.length - 1] === 'profile.html') {
        segments.pop();
      }
      if (segments.length === 0) {
        var params = new URLSearchParams(window.location.search);
        var n = params.get('n');
        return n || '';
      }
      return segments[segments.length - 1];
    },

    loadMock(isOwner) {
      this.profile = {
        user: { id: 1, name: 'dev-mock', is_public: true },
        profile: {
          d: { wood: 0.65, fire: 0.15, earth: -0.20, metal: -0.35, water: -0.25 },
          p: { wood: 0.33, fire: 0.23, earth: 0.16, metal: 0.13, water: 0.15 },
          identity: { label: 'WF', id: 'WF', category: 'W' },
        },
        flow_month: {
          month_id: 'meng-chun',
          month_en: 'Early Spring',
          d_eff: { wood: 0.30, fire: -0.10, earth: 0.10, metal: -0.20, water: -0.10 },
        },
        peers: {
          d: { wood: 0.40, fire: 0.30, earth: -0.10, metal: -0.45, water: -0.15 },
          p: { wood: 0.28, fire: 0.26, earth: 0.18, metal: 0.11, water: 0.17 },
          identity: { label: 'FW', id: 'FW', category: 'F' },
          count: 4,
        },
        is_owner: isOwner,
        is_public: true,
        has_review_link: 'mock-review-token',
      };

      if (isOwner) {
        this.bonds = [
          { id: 101, other_name: 'dev-wood', created_at: '2026-05-01T10:30:00Z' },
          { id: 102, other_name: 'dev-fire', created_at: '2026-04-20T14:00:00Z' },
        ];
        this.reviewLinks = [
          { id: 201, token: 'mock-r-1', submission_count: 2, expires_at: '2026-06-14T00:00:00Z', created_at: '2026-05-10T10:00:00Z' },
          { id: 202, token: 'mock-r-2', submission_count: 0, expires_at: '2026-06-20T00:00:00Z', created_at: '2026-05-12T10:00:00Z' },
        ];
        this.matchLinks = [
          { id: 301, token: 'mock-m-1', created_at: '2026-05-08T08:00:00Z' },
          { id: 302, token: 'mock-m-2', created_at: '2026-05-11T09:00:00Z' },
        ];
        this.history = [
          { id: 1001, profile: { d: { wood: 0.45, fire: 0.10, earth: -0.10, metal: -0.25, water: -0.20 }, p: { wood: 0.29, fire: 0.22, earth: 0.18, metal: 0.15, water: 0.16 }, identity: { label: 'WF', id: 'WF', category: 'W' } }, identity: { label: 'WF', id: 'WF', category: 'W' }, created_at: '2026-03-15T12:00:00Z' },
          { id: 1002, profile: { d: { wood: 0.55, fire: 0.20, earth: -0.15, metal: -0.30, water: -0.30 }, p: { wood: 0.31, fire: 0.24, earth: 0.17, metal: 0.14, water: 0.14 }, identity: { label: 'WE', id: 'WE', category: 'W' } }, identity: { label: 'WE', id: 'WE', category: 'W' }, created_at: '2026-04-01T12:00:00Z' },
          { id: 1003, profile: { d: { wood: 0.65, fire: 0.15, earth: -0.20, metal: -0.35, water: -0.25 }, p: { wood: 0.33, fire: 0.23, earth: 0.16, metal: 0.13, water: 0.15 }, identity: { label: 'WF', id: 'WF', category: 'W' } }, identity: { label: 'WF', id: 'WF', category: 'W' }, created_at: '2026-05-01T12:00:00Z' },
        ];
      } else {
        this.bonds = [];
        this.reviewLinks = [];
        this.matchLinks = [];
        this.history = [];
        this.hasMyProfile = true;
        this.targetReviewToken = this.profile.has_review_link;
        var auth = Alpine.store('auth');
        if (auth) auth.id = 99;
      }

      document.title = this.profile.user.name + ' — 25 Types';
      this.loading = false;

      var self = this;
      this.$nextTick(function () {
        self.renderProfileRadar();
        if (self.profile.peers) self.renderPeerRadar();
        if (self.flowMonth) self.renderFlowMiniRadar();
        if (isOwner && self.history.length > 1) {
          self.$nextTick(function () { self.renderHistoryChart(); });
        }
      });
    },

    async init() {
      var params = new URLSearchParams(window.location.search);
      var mock = params.get('mock');
      if (mock === 'owner' || mock === 'visitor') {
        this.loadMock(mock === 'owner');
        return;
      }

      var name = this.getTargetName();
      if (!name) {
        this.loading = false;
        return;
      }

      document.title = name + ' — 25 Types';
      await this.loadProfile(name);

      if (this.profile && this.profile.is_owner) {
        await Promise.all([
          this.loadBonds(name),
          this.loadReviewLinks(),
          this.loadMatchLinks(),
          this.loadHistory(),
        ]);
      } else if (this.profile && this.auth.id) {
        await Promise.all([
          this.checkMyProfile(),
          this.checkTargetReviewLinks(name),
        ]);
      }
    },

    async loadProfile(name) {
      try {
        var res = await api('/api/profiles/' + encodeURIComponent(name));
        this.profile = res.data || res;
        document.title = (this.profile && this.profile.user ? this.profile.user.name : name) + ' — 25 Types';
      } catch (e) {
        this.profile = null;
      }
      this.loading = false;

      if (this.profile && this.profile.profile) {
        this.$nextTick(() => {
          this.renderProfileRadar();
          if (this.profile.peers) this.renderPeerRadar();
          if (this.flowMonth) this.renderFlowMiniRadar();
        });
      }
    },

    async loadBonds(name) {
      try {
        var res = await api('/api/profiles/' + encodeURIComponent(name) + '/bonds');
        this.bonds = (res.data && res.data.items) || res.items || [];
      } catch (_) { this.bonds = []; }
    },

    async loadReviewLinks() {
      try {
        var res = await api('/api/reviews');
        this.reviewLinks = (res.data && res.data.items) || res.items || [];
      } catch (_) { this.reviewLinks = []; }
    },

    async loadMatchLinks() {
      try {
        var res = await api('/api/match-links');
        this.matchLinks = (res.data && res.data.items) || res.items || [];
      } catch (_) { this.matchLinks = []; }
    },

    async loadHistory() {
      try {
        var res = await api('/api/assessments');
        this.history = (res.data && res.data.items) || res.items || [];
        if (this.history.length > 1) {
          this.$nextTick(() => { this.renderHistoryChart(); });
        }
      } catch (_) { this.history = []; }
    },

    async checkMyProfile() {
      if (!this.auth.id) return;
      try {
        var res = await api('/api/profiles/' + encodeURIComponent(this.auth.name));
        this.hasMyProfile = !!(res.data && res.data.profile);
      } catch (_) { this.hasMyProfile = false; }
    },

    async checkTargetReviewLinks(name) {
      try {
        var res = await api('/api/profiles/' + encodeURIComponent(name));
        this.targetReviewToken = (res.data && res.data.has_review_link) || null;
      } catch (_) { this.targetReviewToken = null; }
    },

    // --- Radar rendering ---
    renderProfileRadar() {
      if (!this.profile || !this.profile.profile) return;
      var el = document.getElementById('profile-radar');
      if (!el) return;
      var pArr = namedToArray(this.profile.profile.p);
      Charts.renderElementRadar(el, {
        series: [{ value: pArr, name: this.profile.profile.identity.label }],
      });
    },

    renderPeerRadar() {
      if (!this.profile || !this.profile.peers) return;
      var el = document.getElementById('peers-radar');
      if (!el) return;
      var pArr = namedToArray(this.profile.profile.p);
      var peerPArr = namedToArray(this.profile.peers.p);
      Charts.renderElementRadar(el, {
        series: [
          { value: pArr, name: this.$store.locale.t('peers.self'), lineStyle: { width: 1.5 }, areaStyle: { opacity: 0.04 } },
          { value: peerPArr, name: this.$store.locale.t('peers.peerReviews'), colorIdx: 1, symbol: 'diamond', areaStyle: { opacity: 0.06 } },
        ],
        legend: true,
      });
    },

    renderFlowMiniRadar() {
      var el = document.getElementById('flow-mini-radar');
      if (!el || !this.flowMonth || !this.profile || !this.profile.profile) return;
      var pOrig = namedToArray(this.profile.profile.p);
      var pFlow = deviationToProportion(namedToArray(this.flowMonth.d_eff));
      this.flowDeltas = ELEMENT_HTML_ORDER.map(function(i) {
        return { idx: i, delta: Math.round((pFlow[i] - pOrig[i]) * 100) / 100 };
      });
      Charts.renderElementRadar(el, {
        series: [
          { value: pOrig, name: this.$store.locale.t('peers.self'), lineStyle: { width: 1.5, type: 'dashed', opacity: 0.6 }, areaStyle: { opacity: 0.04 } },
          { value: pFlow, name: this.flowMonth.month_en || this.flowMonth.month_id, colorIdx: 2 },
        ],
        legend: true,
      });
    },

    // --- History river chart ---
    renderHistoryChart() {
      var el = document.getElementById('history-chart');
      if (!el || !this.history.length || !window.Charts) return;

      var dates = this.history.map((a) => {
        var d = new Date(a.created_at);
        return d.toLocaleDateString(this.currentLocale === 'zh-CN' ? 'zh-CN' : 'en-US', { month: 'short', day: 'numeric' });
      }).reverse();
      var dAll = this.history.map(function(a) {
        if (!a.profile) return null;
        if (a.profile.p) return namedToArray(a.profile.p);
        if (a.profile.d) return deviationToProportion(namedToArray(a.profile.d));
        return null;
      }).reverse();

      Charts.renderElementLine(el, {
        categories: dates,
        series: ELEMENT_DISPLAY_ORDER.map(function(idx) {
          return { colorIdx: idx, data: dAll.map(function(d) { return d ? d[idx] : null; }) };
        }),
      });

    },

    // --- Owner actions ---
    shareProfile() {
      var url = this.origin + '/' + this.currentLocale + '/' + encodeURIComponent(this.profile.user.name);
      this.copyText(url);
      this.$store.toast.success(this.$store.locale.t('profile.profileShared'));
    },

    editProfile() {
      var newName = prompt(this.$store.locale.t('register.username'), this.profile.user.name);
      if (!newName || newName === this.profile.user.name) return;
      api('/api/users/me', {
        method: 'PATCH',
        body: JSON.stringify({ name: newName }),
      }).then(() => {
        this.profile.user.name = newName;
        this.$store.auth.name = newName;
        this.$store.toast.success(this.$store.locale.t('toast.settingsUpdated'));
        var newPath = '/' + this.currentLocale + '/' + encodeURIComponent(newName);
        window.history.replaceState({}, '', newPath);
        document.title = newName + ' — 25 Types';
      }).catch(function() {});
    },

    async togglePrivacy() {
      var current = this.profile.is_public;
      try {
        await api('/api/users/me', {
          method: 'PATCH',
          body: JSON.stringify({ is_public: !current }),
        });
        this.profile.is_public = !current;
        this.profile.user.is_public = !current;
        if (this.$store.auth) this.$store.auth.isPublic = !current;
        this.$store.toast.success(current ? this.$store.locale.t('profile.madePrivate') : this.$store.locale.t('profile.madePublic'));
      } catch (_) {}
    },

    async createReviewLink() {
      try {
        await api('/api/reviews', { method: 'POST' });
        this.$store.toast.success(this.$store.locale.t('profile.linkCreated'));
        await this.loadReviewLinks();
      } catch (_) {}
    },

    async createMatchLink() {
      try {
        await api('/api/match-links', { method: 'POST' });
        this.$store.toast.success(this.$store.locale.t('profile.linkCreated'));
        await this.loadMatchLinks();
      } catch (_) {}
    },

    async exportData() {
      try {
        var res = await api('/api/users/me/export');
        var blob = new Blob([JSON.stringify(res.data || res, null, 2)], { type: 'application/json' });
        var a = document.createElement('a');
        a.href = URL.createObjectURL(blob);
        a.download = '25types-export-' + new Date().toISOString().slice(0, 10) + '.json';
        a.click();
        URL.revokeObjectURL(a.href);
        this.$store.toast.success(this.$store.locale.t('toast.dataExported'));
      } catch (_) {}
    },

    async deleteAccount() {
      if (!confirm(this.$store.locale.t('confirm.deleteAccount'))) return;
      try {
        await api('/api/users/me', { method: 'DELETE' });
        this.$store.toast.success(this.$store.locale.t('toast.accountDeactivated'));
        this.$store.auth.logout();
      } catch (_) {}
    },

    async deleteReviewLink(id) {
      if (!confirm(this.$store.locale.t('confirm.deleteLink'))) return;
      try {
        await api('/api/reviews/' + id, { method: 'DELETE' });
        this.reviewLinks = this.reviewLinks.filter(function(l) { return l.id !== id; });
      } catch (_) {}
    },

    async deleteMatchLink(id) {
      if (!confirm(this.$store.locale.t('confirm.deleteLink'))) return;
      try {
        await api('/api/match-links/' + id, { method: 'DELETE' });
        this.matchLinks = this.matchLinks.filter(function(l) { return l.id !== id; });
      } catch (_) {}
    },

    // --- Visitor: instant compare ---
    async compareWithMe() {
      if (!this.profile || !this.profile.user) return;
      try {
        var body = {};
        if (this.profile.user.id) {
          body.with_user_id = this.profile.user.id;
        } else {
          body.with_name = this.profile.user.name;
        }
        var res = await api('/api/bond', { method: 'POST', body: JSON.stringify(body) });
        var bond = res.data || res;
        var arrA = namedToArray(bond.delta_a);
        var arrB = namedToArray(bond.delta_b);
        this.bondDeltaA = ELEMENT_HTML_ORDER.map(function(i) { return { idx: i, delta: arrA[i] }; });
        this.bondDeltaB = ELEMENT_HTML_ORDER.map(function(i) { return { idx: i, delta: arrB[i] }; });
        this.showBond = true;
        this.$nextTick(() => { this.renderBond(bond); });
      } catch (_) {}
    },

    // --- Bond rendering ---
    renderBond(bond) {
      var el = document.getElementById('bond-radar');
      if (!el) return;
      this.$nextTick(() => {
        var pSelf = deviationToProportion(namedToArray(bond.self));
        var pOther = deviationToProportion(namedToArray(bond.other));
        Charts.renderElementRadar(el, {
          series: [
            { value: pSelf, name: this.$store.locale.t('bond.yourShape') },
            { value: pOther, name: this.$store.locale.t('bond.theirShape'), colorIdx: 1 },
          ],
          legend: true,
        }, {
          radar: { center: ['50%', '48%'], radius: '60%' },
        });
        var arrA = namedToArray(bond.delta_a);
        var arrB = namedToArray(bond.delta_b);
        this.bondDeltaA = ELEMENT_HTML_ORDER.map(function(i) { return { idx: i, delta: arrA[i] }; });
        this.bondDeltaB = ELEMENT_HTML_ORDER.map(function(i) { return { idx: i, delta: arrB[i] }; });
      });
    },

    // --- Helpers ---
    // elementColor/elementName are global helpers in common.js
    fmtDate(s) {
      if (!s) return '';
      var d = new Date(s);
      return d.toLocaleDateString(this.currentLocale === 'zh-CN' ? 'zh-CN' : 'en-US', { year: 'numeric', month: 'short', day: 'numeric' });
    },
    async copyText(text) {
      try {
        await navigator.clipboard.writeText(text);
        this.$store.toast.success(this.$store.locale.t('copied'));
      } catch (_) {
        this.$store.toast.error(this.$store.locale.t('toast.copyFailed'));
      }
    },
  };
}
