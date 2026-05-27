// match-landing.js — Match link landing page /m/{token}
function matchLandingPage() {
  return Object.assign({
    token: '',
    allQuestions: [],
    answers: {},
    currentQIndex: 0,
    loading: true,
    error: '',
    submitting: false,
    result: null,
    bondChordReady: false,

    get currentLocale() { return window.CURRENT_LOCALE || 'en'; },
    get auth() { return Alpine.store('auth'); },
  }, assessmentNavigation(), {
    get allAnswered() { return this.answeredCount >= this.totalQuestions; },

    get pValues() {
      if (!this.result || !this.result.profile || !this.result.profile.p) return [];
      return namedToArray(this.result.profile.p);
    },
    get deltaA() {
      if (!this.result || !this.result.bond) return [];
      var arr = namedToArray(this.result.bond.delta_a);
      return ELEMENT_HTML_ORDER.map(function(i) { return { idx: i, delta: arr[i] }; });
    },
    get deltaB() {
      if (!this.result || !this.result.bond) return [];
      var arr = namedToArray(this.result.bond.delta_b);
      return ELEMENT_HTML_ORDER.map(function(i) { return { idx: i, delta: arr[i] }; });
    },

    async init() {
      // Extract token from URL path: /m/{token} or /en/m/{token}
      var path = window.location.pathname;
      var match = path.match(/\/m\/([a-zA-Z0-9_-]+)/);
      this.token = match ? match[1] : '';

      if (!this.token) {
        this.error = this.$store.locale.t('error.invalidLink');
        this.loading = false;
        return;
      }

      // Verify link is valid
      try {
        await api('/api/m/' + this.token);
      } catch (e) {
        this.error = this.$store.locale.t('error.linkInvalid');
        this.loading = false;
        return;
      }

      // Load questions
      var locale = Alpine.store('locale').current;
      try {
        this.allQuestions = await loadAssessmentQuestions(locale);
      } catch (_) {
        this.error = Alpine.store('locale').t('error.generic');
      }
      this.loading = false;
    },

    async submit() {
      if (!this.allAnswered) return;
      this.submitting = true;
      try {
        var answers = [];
        for (var i = 0; i < this.allQuestions.length; i++) {
          var q = this.allQuestions[i];
          var sel = this.answers[q.qid];
          if (sel && sel.length === 2) {
            answers.push({ qid: q.qid, selections: sel });
          }
        }
        var resp = await api('/api/m/' + this.token, {
          method: 'POST',
          body: JSON.stringify({ answers: answers }),
        });
        var data = (resp && resp.data) ? resp.data : resp;
        this.result = data;
        this.$nextTick(() => {
          this.renderResults();
        });
      } catch (e) {
        this.$store.toast.error(e.message || this.$store.locale.t('error.generic'));
      } finally {
        this.submitting = false;
      }
    },

    renderResults() {
      if (!this.result) return;
      // Render profile radar
      var pArr = namedToArray(this.result.profile.p);
      var radarEl = document.getElementById('result-radar');
      if (radarEl) {
        Charts.renderElementRadar(radarEl, {
          series: [{ value: pArr, name: this.result.profile.identity.label }],
        });
      }

      // Render bond radar if available
      if (this.result.bond) {
        var bondEl = document.getElementById('bond-radar');
        if (bondEl) {
          var selfP = deviationToProportion(namedToArray(this.result.bond.self));
          var otherP = deviationToProportion(namedToArray(this.result.bond.other));
          Charts.renderElementRadar(bondEl, {
            series: [
              { value: selfP, name: 'You', symbolSize: 3, lineStyle: { width: 1.5 } },
              { value: otherP, name: 'Them', colorIdx: 1, symbol: 'diamond', symbolSize: 4, lineStyle: { width: 2 } },
            ],
            legend: true,
          });

          // Render chord diagram
          var chordEl = document.getElementById('bond-chord-chart');
          if (chordEl) {
            Charts.renderElementChord(chordEl, {
              personA: { name: 'You', p: selfP },
              personB: { name: 'Them', p: otherP },
            });
            this.bondChordReady = true;
          }
        }
      }
    },
  }, makePickTwo(function() { return this.answers; }));
}
