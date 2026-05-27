function reviewLanding() {
  return composeComponent(
    {
      link: null,
      questions: [],
      selections: {},
      loading: true,
      error: '',
      submitted: false,
      submitting: false,
      reviewerName: '',

      init() {
        var path = location.pathname;
        var token = path.split('/r/')[1];
        if (!token) {
          this.error = Alpine.store('locale').t('error.invalidLink');
          this.loading = false;
          return;
        }
        var self = this;
        this.$nextTick(function () {
          self.loadReview(token);
        });
      },

      async loadReview(token) {
        try {
          var locale = (Alpine.store('locale') || {}).current || 'en';
          var resp = await api('/api/r/' + token + '?locale=' + locale);
          if (resp.data) {
            this.link = resp.data;
            if (!resp.data.valid) {
              this.error = resp.data.expired
                ? Alpine.store('locale').t('error.linkExpired')
                : Alpine.store('locale').t('error.linkInvalid');
            }
            if (resp.data.questions && resp.data.questions.length) {
              this.questions = resp.data.questions;
            }
          }
        } catch (e) {
          console.error(e);
          this.error = Alpine.store('locale').t('error.linkNotFound');
        }
        if (Array.isArray(this.questions)) {
          this.questions.forEach(function(q) { this.selections[q.qid] = []; }.bind(this));
        }
        this.loading = false;
      },

      async submitReview() {
        var path = location.pathname;
        var token = path.split('/r/')[1];
        this.submitting = true;
        try {
          var answers = Object.entries(this.selections).map(function(e) {
            return { qid: e[0], selections: e[1] };
          });
          await api('/api/r/' + token, {
            method: 'POST',
            body: JSON.stringify({ reviewer_name: this.reviewerName, answers: answers }),
          });
          this.submitted = true;
        } catch (e) {
          console.error(e);
        } finally {
          this.submitting = false;
        }
      },
    },
    makePickAny(function() { return this.selections; })
  );
}
