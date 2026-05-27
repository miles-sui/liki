function assessPage() {
  return composeComponent(
    {
      allQuestions: [],
      answers: {},
      currentQIndex: 0,
      loading: true,
      error: '',
      submitting: false,
      lastResult: null,
      anonymousToken: '',
      showInterstitial: false,
      completedRound: 0,
      showOnboarding: false,

      get topElementLabel() {
        if (!this.lastResult || !this.lastResult.profile || !this.lastResult.profile.p) return '';
        var p = this.lastResult.profile.p;
        var maxIdx = 0, maxVal = p[0];
        for (var i = 1; i < p.length; i++) {
          if (p[i] > maxVal) { maxVal = p[i]; maxIdx = i; }
        }
        var codes = window.ELEMENT_CODES || ['W','F','E','M','R'];
        var names = window.ELEMENT_NAMES || {W:'Wood',F:'Fire',E:'Earth',M:'Metal',R:'Water'};
        return names[codes[maxIdx]] || '';
      },

      startAssessment() {
        this.showOnboarding = false;
        sessionStorage.setItem('assess_onboarded', '1');
      },

      continueAssessment() { this.showInterstitial = false; },

      get resultParams() {
        var r = this.lastResult;
        if (!r || !r.profile) return '';
        var p = new URLSearchParams();
        if (r.profile.d) p.set('d', JSON.stringify(r.profile.d));
        if (r.profile.p) p.set('p', JSON.stringify(r.profile.p));
        if (r.identity) {
          p.set('label', r.identity.label || r.identity.id);
          p.set('id', r.identity.id);
          p.set('category', r.identity.category);
        }
        return p.toString();
      },

      startOver() {
        localStorage.removeItem('anon_assessment_done');
        this.lastResult = null;
        this.allQuestions = [];
        this.answers = {};
        this.currentQIndex = 0;
        this.showInterstitial = false;
        this.showOnboarding = false;
        this.loading = true;
        this.error = '';
        this.init();
      },

      init() {
        this.anonymousToken = generateAnonToken();

        // Check onboarding first-visit
        if (!sessionStorage.getItem('assess_onboarded')) {
          this.showOnboarding = true;
        }

        var prev = localStorage.getItem('anon_assessment_done');
        if (prev) {
          try {
            var parsed = JSON.parse(prev);
            if (parsed.profile) {
              if (parsed.profile.p) parsed.profile.p = namedToArray(parsed.profile.p);
              if (parsed.profile.d) parsed.profile.d = namedToArray(parsed.profile.d);
            }
            this.lastResult = parsed;
            this.loading = false;
            return;
          } catch (_) {}
        }

        var self = this;
        this.$nextTick(function () {
          self.fetchQuestions();
        });
      },

      async fetchQuestions() {
        try {
          var locale = (Alpine.store('locale') || {}).current || 'en';
          this.allQuestions = await loadAssessmentQuestions(locale);
          if (!this.allQuestions || this.allQuestions.length === 0) {
            this.error = this.$store.locale.t('error.loadQuestions');
          }
        } catch (e) {
          console.error(e);
          this.error = this.$store.locale.t('error.loadQuestions');
        } finally {
          this.loading = false;
        }
      },

      async submitRound() {
        if (this.answeredCount === 0) {
          try { Alpine.store('toast').show(Alpine.store('locale').t('assess.emptySubmit'), 'warning'); } catch (_) {}
          return;
        }
        this.submitting = true;
        try {
          var answers = [];
          for (var i = 0; i <= this.currentQIndex; i++) {
            var q = this.allQuestions[i];
            var sel = this.answers[q.qid];
            answers.push({ qid: q.qid, selections: sel || [] });
          }
          var resp = await api('/api/assessments', {
            method: 'POST',
            body: JSON.stringify({ answers: answers, anonymous_token: this.anonymousToken }),
          });
          var data = (resp && resp.data) ? resp.data : resp;
          if (data.complete) {
            var auth = Alpine.store('auth');
            if (auth && auth.id) {
              window.location.href = '/app#/overview';
              return;
            }
            localStorage.setItem('anon_assessment_done', JSON.stringify({
              profile: data.profile,
              identity: data.identity,
            }));
            var params = new URLSearchParams();
            if (data.profile && data.profile.d) params.set('d', JSON.stringify(data.profile.d));
            if (data.profile && data.profile.p) params.set('p', JSON.stringify(data.profile.p));
            if (data.identity) {
              params.set('label', data.identity.label || data.identity.id);
              params.set('id', data.identity.id);
              params.set('category', data.identity.category);
            }
            window.location.href = localePath('/result') + '?' + params.toString();
          } else {
            if (data.profile) {
              if (data.profile.p) data.profile.p = namedToArray(data.profile.p);
              if (data.profile.d) data.profile.d = namedToArray(data.profile.d);
            }
            this.lastResult = data;
            this.completedRound = this.round - 1;
            this.showInterstitial = true;
          }
        } catch (e) {
          console.error(e);
        } finally {
          this.submitting = false;
        }
      },
    },
    assessmentNavigation(),
    makePickAny(function() { return this.answers; })
  );
}
