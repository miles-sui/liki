function reviewLanding() {
  return Object.assign({
    link: null, questions: [], selections: {}, loading: true, error: '', submitted: false, submitting: false, reviewerName: '',
    async init() {
      var path = location.pathname;
      var token = path.split('/r/')[1];
      if (!token) { this.error = Alpine.store('locale').t('error.invalidLink'); this.loading = false; return; }
      try {
        var resp = await api('/api/r/' + token);
        if (resp.data) {
          this.link = resp.data;
          if (!resp.data.valid) this.error = resp.data.expired ? Alpine.store('locale').t('error.linkExpired') : Alpine.store('locale').t('error.linkInvalid');
          if (resp.data.questions && resp.data.questions.length) {
            this.questions = resp.data.questions;
          }
        }
      } catch (_) { this.error = Alpine.store('locale').t('error.linkNotFound'); }
      this.questions.forEach((q) => { this.selections[q.qid] = []; });
      this.loading = false;
    },
    async submitReview() {
      var path = location.pathname;
      var token = path.split('/r/')[1];
      this.submitting = true;
      try {
        var answers = Object.entries(this.selections).map(function(e) { return { qid: e[0], selections: e[1] }; });
        await api('/api/r/' + token, { method: 'POST', body: JSON.stringify({ reviewer_name: this.reviewerName, answers: answers }) });
        this.submitted = true;
      } catch (_) {} finally { this.submitting = false; }
    },
  }, makePickTwo(function() { return this.selections; }));
}
