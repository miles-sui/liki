function resetPage() {
  return {
    password: '', showPw: false, loading: false, error: '', success: false,
    get strength() { return passwordStrength(this.password); },
    barClass(n) { return pwBarClass(this.strength.score, n); },

    async handleSubmit() {
      this.loading = true; this.error = '';
      try {
        var token = new URLSearchParams(location.search).get('token');
        await api('/api/auth/reset-password', {
          method: 'POST',
          body: JSON.stringify({ token: token, password: this.password }),
        });
        this.success = true;
      } catch (e) {
        console.error(e);
        this.error = e.message || Alpine.store('locale').t('error.failed');
      } finally { this.loading = false; }
    },
  };
}
