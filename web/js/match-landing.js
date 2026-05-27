// match-landing.js — Match link landing page /m/{token}
function matchLandingPage() {
  return composeComponent(
    {
      token: '',
      allQuestions: [],
      answers: {},
      currentQIndex: 0,
      loading: true,
      error: '',
      submitting: false,
      result: null,
      showInterstitial: false,

      get currentLocale() { return window.CURRENT_LOCALE || 'en'; },
      get auth() { return Alpine.store('auth'); },
      get allAnswered() { return this.answeredCount > 0; },

      get pValues() {
        if (!this.result || !this.result.profile || !this.result.profile.p) return [];
        return namedToArray(this.result.profile.p);
      },
      get bondA() {
        if (!this.result || !this.result.profile) return null;
        return this.result.profile.identity;
      },
      get bondB() {
        if (!this.result || !this.result.bond) return null;
        return identityFromD(this.result.bond.other);
      },
      get bondInfluencerA() {
        var b = this.bondB;
        return b ? b.label : '';
      },
      get bondInfluencerB() {
        var a = this.bondA;
        return a ? a.label : '';
      },
      get computedBond() {
        if (!this.result || !this.result.bond) return null;
        if (!this._cachedBondShapes) {
          this._cachedBondShapes = computeBondShapes(this.result.bond);
        }
        return this._cachedBondShapes;
      },
      get bondDeltasA() {
        var b = this.computedBond;
        return b ? b.selfDeltas : [];
      },
      get bondDeltasB() {
        var b = this.computedBond;
        return b ? b.otherDeltas : [];
      },
      get concordLabel() {
        var raw = this.result && this.result.bond ? this.result.bond.concord : null;
        return concordProps(raw).label;
      },
      get concordBadgeClass() {
        var raw = this.result && this.result.bond ? this.result.bond.concord : null;
        return concordProps(raw).badgeClass;
      },
      get concordDesc() {
        var raw = this.result && this.result.bond ? this.result.bond.concord : null;
        return concordProps(raw).desc;
      },

      anonymousToken: '',
      existingProfile: null,
      checkingProfile: false,

      get hasExistingProfile() {
        return !!(this.auth.id && this.existingProfile);
      },

      init() {
        var path = window.location.pathname;
        var match = path.match(/\/m\/([a-zA-Z0-9_-]+)/);
        this.token = match ? match[1] : '';

        if (!this.token) {
          this.error = this.$store.locale.t('error.invalidLink');
          this.loading = false;
          return;
        }

        this.anonymousToken = generateAnonToken();

        var self = this;
        this.$nextTick(function () {
          self.loadAndVerify();
        });
      },

      async loadAndVerify(skipProfileCheck) {
        try {
          await api('/api/m/' + this.token);
          var locale = (Alpine.store('locale') || {}).current || 'en';
          // Only load questions if user needs the assessment (no existing profile).
          if (!skipProfileCheck && this.auth.id) {
            this.checkingProfile = true;
            try {
              var res = await api('/api/profiles/' + encodeURIComponent(this.auth.name));
              if (res.data && res.data.profile) {
                this.existingProfile = res.data.profile;
                this.loading = false;
                this.checkingProfile = false;
                return;
              }
            } catch (e) { console.error(e); }
            this.checkingProfile = false;
          }
          this.allQuestions = await loadAssessmentQuestions(locale);
        } catch (e) {
          console.error(e);
          this.error = this.$store.locale.t('error.linkInvalid');
        }
        this.loading = false;
      },

      async computeDirectBond() {
        this.submitting = true;
        try {
          var resp = await api('/api/m/' + this.token, {
            method: 'POST',
            body: JSON.stringify({ use_existing: true }),
          });
          var data = (resp && resp.data) ? resp.data : resp;
          this._cachedBondShapes = null;
          this.result = data;
          this.$nextTick(function () {
            this.renderResults();
          }.bind(this));
        } catch (e) {
          console.error(e);
          this.$store.toast.error(e.message || this.$store.locale.t('error.generic'));
        } finally {
          this.submitting = false;
        }
      },

      submitRound() { this.submit(); },

      async submit() {
        if (!this.allAnswered) {
          try { Alpine.store('toast').show(Alpine.store('locale').t('assess.emptySubmit'), 'warning'); } catch (_) {}
          return;
        }
        this.submitting = true;
        try {
          var answers = [];
          for (var i = 0; i < this.allQuestions.length; i++) {
            var q = this.allQuestions[i];
            var sel = this.answers[q.qid];
            answers.push({ qid: q.qid, selections: sel || [] });
          }
          var resp = await api('/api/m/' + this.token, {
            method: 'POST',
            body: JSON.stringify({ answers: answers, anonymous_token: this.anonymousToken }),
          });
          var data = (resp && resp.data) ? resp.data : resp;
          this._cachedBondShapes = null;
          this.result = data;
          this.$nextTick(function () {
            this.renderResults();
          }.bind(this));
        } catch (e) {
          console.error(e);
          this.$store.toast.error(e.message || this.$store.locale.t('error.generic'));
        } finally {
          this.submitting = false;
        }
      },

      renderResults() {
        if (!this.result) return;
        var pArr = namedToArray(this.result.profile.p);
        var radarEl = document.getElementById('result-radar');
        if (radarEl) {
          Charts.renderElementRadar(radarEl, {
            series: [{ value: pArr, name: this.result.profile.identity.label }],
          });
        }

        if (this.result.bond) {
          var shapes = this.computedBond;

          Charts.renderBondInfluenceChart(document.getElementById('bond-influence-self'), {
            origData: shapes.origSelf, pData: shapes.pSelf,
            yourLabel: Alpine.store('locale').t('bond.yourShape'), influencedLabel: Alpine.store('locale').t('bond.influenced'),
            overrides: { radar: { center: ['50%', '50%'], radius: '60%' } },
          });
          Charts.renderBondInfluenceChart(document.getElementById('bond-influence-other'), {
            origData: shapes.origOther, pData: shapes.pOther,
            yourLabel: Alpine.store('locale').t('bond.theirShape'), influencedLabel: Alpine.store('locale').t('bond.influenced'),
            overrides: { radar: { center: ['50%', '50%'], radius: '60%' } },
          });
        }
      },
    },
    assessmentNavigation(),
    makePickAny(function() { return this.answers; })
  );
}
