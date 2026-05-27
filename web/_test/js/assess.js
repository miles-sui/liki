function assessPage() {
  return Object.assign({
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
  }, assessmentNavigation(), {
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
      this.loading = true;
      this.error = '';
      this.init();
    },

    async init() {
      var stored = sessionStorage.getItem('anon_token');
      this.anonymousToken = stored || 'anon-' + (crypto.randomUUID ? crypto.randomUUID() : Math.random().toString(36).slice(2, 10));
      sessionStorage.setItem('anon_token', this.anonymousToken);

      // Check for prior anonymous completion
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
        } catch (_) { /* corrupted, ignore */ }
      }

      try {
        var locale = (Alpine.store('locale') || {}).current || 'en';
        this.allQuestions = await loadAssessmentQuestions(locale);
      } catch (_) { /* ignore */ }
      if (!this.allQuestions || this.allQuestions.length === 0) {
        this.error = 'Failed to load questions';
      }
      this.loading = false;
    },

    async submitRound() {
      this.submitting = true;
      try {
        var answers = [];
        for (var i = 0; i <= this.currentQIndex; i++) {
          var q = this.allQuestions[i];
          var sel = this.answers[q.qid];
          if (sel && sel.length === 2) {
            answers.push({ qid: q.qid, selections: sel });
          }
        }
        var resp = await api('/api/assessments', {
          method: 'POST',
          body: JSON.stringify({ answers: answers, anonymous_token: this.anonymousToken }),
        });
        var data = (resp && resp.data) ? resp.data : resp;
        if (data.complete) {
          var auth = Alpine.store('auth');
          if (!auth || !auth.id) {
            localStorage.setItem('anon_assessment_done', JSON.stringify({
              profile: data.profile,
              identity: data.identity,
            }));
          }
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
      } catch (_) {
      } finally {
        this.submitting = false;
      }
    },
  }, makePickTwo(function() { return this.answers; }));
}
